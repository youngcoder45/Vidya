package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/schoolos/backend/internal/config"
	"github.com/schoolos/backend/internal/pkg/audit"
	"github.com/schoolos/backend/internal/pkg/ctxuser"
	"github.com/schoolos/backend/internal/pkg/jwtutil"
	"github.com/schoolos/backend/internal/pkg/passwd"
	"github.com/schoolos/backend/internal/pkg/randutil"
)

var (
	// ErrInvalidCredentials is returned for wrong identifier/password.
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
	// ErrUserSuspended is returned for suspended accounts.
	ErrUserSuspended = errors.New("auth: user suspended")
	// ErrInvalidOtp covers wrong/expired/attempt-limit OTPs.
	ErrInvalidOtp = errors.New("auth: invalid or expired otp")
	// ErrInvalidRefresh covers unknown/revoked refresh tokens.
	ErrInvalidRefresh = errors.New("auth: invalid refresh token")
	// ErrPasswordResetUnsupported guards the not-yet-implemented reset flow.
	ErrPasswordResetUnsupported = errors.New("auth: password reset not implemented")
)

// DeviceInput describes the calling device for login/session tracking.
type DeviceInput struct {
	DeviceID string `json:"device_id" binding:"required,max=128"`
	Name     string `json:"name" binding:"max=120"`
	Platform string `json:"platform" binding:"max=20"`
	FCMToken string `json:"fcm_token" binding:"max=255"`
}

// AuthResult is the token payload returned to clients.
type AuthResult struct {
	AccessToken  string    `json:"access_token"`
	ExpiresIn    int64     `json:"expires_in"`
	RefreshToken string    `json:"refresh_token"`
	User         *User     `json:"user"`
	Roles        []string  `json:"roles"`
	Permissions  []string  `json:"permissions"`
	SchoolID     uuid.UUID `json:"school_id"`
}

// Service implements auth use cases.
type Service struct {
	repo   Repository
	issuer *jwtutil.Issuer
	cfg    *config.Config
	audit  *audit.Writer
	log    *slog.Logger
}

// NewService wires the auth service.
func NewService(repo Repository, issuer *jwtutil.Issuer, cfg *config.Config, audit *audit.Writer, log *slog.Logger) *Service {
	return &Service{repo: repo, issuer: issuer, cfg: cfg, audit: audit, log: log}
}

// Login authenticates staff by email/phone + password.
func (s *Service) Login(ctx context.Context, identifier, password string, dev DeviceInput) (*AuthResult, error) {
	user, err := s.repo.FindUserByIdentity(ctx, nil, identifier)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, ErrAmbiguousIdentity) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if !passwd.Verify(user.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}
	if user.Status != UserActive {
		return nil, ErrUserSuspended
	}
	now := time.Now()
	user.LastLoginAt = &now
	_ = s.repo.UpdateUser(ctx, user)
	return s.issueTokens(ctx, user, dev)
}

// RequestOtp creates and "sends" an OTP. In dev the OTP is logged; the SMS
// adapter is the production channel.
func (s *Service) RequestOtp(ctx context.Context, schoolID uuid.UUID, phone, purpose string) error {
	// Supersede any still-valid codes for this phone/purpose.
	if err := s.repo.InvalidateOtps(ctx, schoolID, phone, purpose); err != nil {
		return err
	}
	code, err := randutil.Digits(6)
	if err != nil {
		return err
	}
	// Hash with bcrypt: a 6-digit space makes a plain hash trivially reversible.
	hash, err := passwd.Hash(code)
	if err != nil {
		return err
	}
	otp := &OtpCode{
		ID: uuid.New(), SchoolID: schoolID, Phone: phone, Purpose: purpose,
		CodeHash: hash, ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	if err := s.repo.CreateOtp(ctx, otp); err != nil {
		return err
	}
	// Channel adapter point: send OTP via SMS/WhatsApp/email here.
	if s.cfg.AppEnv == "local" {
		s.log.Info("otp generated (dev channel)", "school_id", schoolID, "phone", phone, "otp", code)
	} else {
		s.log.Info("otp generated", "school_id", schoolID, "phone", phone)
	}
	return nil
}

// VerifyOtp validates an OTP and returns tokens. Parents get a user account
// created on first verify (linked by phone).
func (s *Service) VerifyOtp(ctx context.Context, schoolID uuid.UUID, phone, purpose, code string, dev DeviceInput) (*AuthResult, error) {
	otp, err := s.repo.FindLatestOtp(ctx, schoolID, phone, purpose)
	if err != nil {
		return nil, ErrInvalidOtp
	}
	if otp.VerifiedAt != nil || time.Now().After(otp.ExpiresAt) || otp.Attempts >= 3 {
		return nil, ErrInvalidOtp
	}
	if !passwd.Verify(otp.CodeHash, code) {
		_ = s.repo.IncrementOtpAttempts(ctx, otp.ID)
		return nil, ErrInvalidOtp
	}
	if err := s.repo.MarkOtpVerified(ctx, otp.ID); err != nil {
		return nil, err
	}
	if purpose == OtpPurposePasswordReset {
		// A reset-purpose OTP must never mint a normal session.
		return nil, ErrPasswordResetUnsupported
	}

	user, err := s.repo.FindUserByIdentity(ctx, &schoolID, phone)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		// First-time parent: create account + parent role.
		randomPass, _ := randutil.Hex(16)
		hash, _ := passwd.Hash(randomPass)
		user = &User{
			ID: uuid.New(), SchoolID: &schoolID, FullName: phone,
			Phone: phone, PasswordHash: hash, Status: UserActive,
		}
		if err := s.repo.CreateUser(ctx, user); err != nil {
			return nil, err
		}
		role, err := s.repo.FindRoleByCode(ctx, "parent")
		if err != nil {
			return nil, err
		}
		if err := s.repo.AssignRole(ctx, schoolID, user.ID, role.ID); err != nil {
			return nil, err
		}
	}
	if user.Status != UserActive {
		return nil, ErrUserSuspended
	}
	s.audit.Record(ctx, audit.Entry{
		SchoolID: &schoolID, UserID: &user.ID, Action: "auth.otp_login",
		EntityType: "user", EntityID: user.ID.String(), IP: "",
	})
	return s.issueTokens(ctx, user, dev)
}

// Refresh rotates a refresh token and issues a new pair. Reuse of a revoked
// token is rejected (rotation-invalidation).
func (s *Service) Refresh(ctx context.Context, refreshToken, deviceID string) (*AuthResult, error) {
	session, err := s.repo.FindSessionByRefreshHash(ctx, sha256Hex(refreshToken))
	if err != nil {
		return nil, ErrInvalidRefresh
	}
	if session.RevokedAt != nil || time.Now().After(session.ExpiresAt) {
		// Possible reuse: revoke the whole device session family.
		_ = s.repo.RevokeSessionsByDevice(ctx, session.SchoolID, session.UserID, session.DeviceID)
		return nil, ErrInvalidRefresh
	}
	user, err := s.repo.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, ErrInvalidRefresh
	}
	if user.Status != UserActive {
		return nil, ErrUserSuspended
	}
	_ = s.repo.RevokeSession(ctx, session.ID)
	return s.issueTokens(ctx, user, DeviceInput{DeviceID: deviceID})
}

// Logout revokes the current session and device.
func (s *Service) Logout(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID, deviceID string) error {
	if sessionID != uuid.Nil {
		if err := s.repo.RevokeSession(ctx, sessionID); err != nil {
			return err
		}
	}
	if claims, ok := ctxuser.From(ctx); ok && deviceID != "" {
		var schoolID *uuid.UUID
		if claims.SchoolID != uuid.Nil {
			schoolID = &claims.SchoolID
		}
		_ = s.repo.RevokeSessionsByDevice(ctx, schoolID, userID, deviceID)
	}
	return nil
}

// Me returns the authenticated user's profile.
func (s *Service) Me(ctx context.Context, userID uuid.UUID) (*User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

// ChangePassword verifies the old password and sets a new one.
func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if !passwd.Verify(user.PasswordHash, oldPassword) {
		return ErrInvalidCredentials
	}
	hash, err := passwd.Hash(newPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	return s.repo.UpdateUser(ctx, user)
}

// issueTokens builds the token pair, persists the session, and returns the
// full auth result with roles + permissions.
func (s *Service) issueTokens(ctx context.Context, user *User, dev DeviceInput) (*AuthResult, error) {
	schoolID := uuid.Nil
	if user.SchoolID != nil {
		schoolID = *user.SchoolID
	}

	var roles, perms []string
	if user.IsSuperadmin {
		roles = append(roles, "platform_admin")
		perms = append(perms, "*")
	} else if schoolID != uuid.Nil {
		roleRows, err := s.repo.ListRolesByUser(ctx, schoolID, user.ID)
		if err != nil {
			return nil, err
		}
		permRows, err := s.repo.ListPermissionsByUser(ctx, schoolID, user.ID)
		if err != nil {
			return nil, err
		}
		for _, r := range roleRows {
			roles = append(roles, r.Code)
		}
		for _, p := range permRows {
			perms = append(perms, p.Code)
		}
	}

	access, _, err := s.issuer.Issue(user.ID, schoolID, roles, perms, dev.DeviceID)
	if err != nil {
		return nil, err
	}
	rawRefresh, err := randutil.Hex(32)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := &AuthSession{
		ID: uuid.New(), SchoolID: user.SchoolID, UserID: user.ID, DeviceID: dev.DeviceID,
		RefreshTokenHash: sha256Hex(rawRefresh),
		ExpiresAt:        now.Add(s.cfg.JWTRefreshTTL),
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, err
	}
	if schoolID != uuid.Nil && dev.DeviceID != "" {
		_ = s.repo.UpsertDevice(ctx, &UserDevice{
			SchoolID: schoolID, UserID: user.ID, DeviceID: dev.DeviceID,
			DeviceName: dev.Name, Platform: dev.Platform, FCMToken: dev.FCMToken,
			LastSeenAt: now,
		})
	}

	return &AuthResult{
		AccessToken: access, ExpiresIn: int64(s.cfg.JWTAccessTTL.Seconds()),
		RefreshToken: rawRefresh, User: user, Roles: roles,
		Permissions: perms, SchoolID: schoolID,
	}, nil
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
