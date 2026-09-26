package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/schoolos/backend/internal/deps"
	"github.com/schoolos/backend/internal/pkg/ctxuser"
	"github.com/schoolos/backend/internal/pkg/httpx"
	"github.com/schoolos/backend/internal/server/middleware"
)

const rateWindow = time.Minute

// RegisterPublic wires the unauthenticated auth routes (login, OTP, refresh).
// These must NOT sit behind the JWT middleware — the router registers them on
// the public group.
func RegisterPublic(rg *gin.RouterGroup, d *deps.Deps) {
	repo := NewRepository(d.DB)
	svc := NewService(repo, d.Issuer, d.Cfg, d.Audit, d.Log)
	h := &handler{svc: svc, repo: repo}

	authGroup := rg.Group("/auth")
	authGroup.POST("/login", middleware.AuthRateLimit(d.RDB, d.Cfg.AuthRateLimitPerMin, rateWindow, d.Log), h.login)
	authGroup.POST("/otp/request", middleware.AuthRateLimit(d.RDB, d.Cfg.AuthRateLimitPerMin, rateWindow, d.Log), h.requestOtp)
	authGroup.POST("/otp/verify", middleware.AuthRateLimit(d.RDB, d.Cfg.AuthRateLimitPerMin, rateWindow, d.Log), h.verifyOtp)
	authGroup.POST("/refresh", middleware.AuthRateLimit(d.RDB, d.Cfg.AuthRateLimitPerMin, rateWindow, d.Log), h.refresh)
}

// RegisterAuthed wires the authenticated auth routes. The router registers
// these on the JWT+tenant-protected group, so no auth middleware is added here.
func RegisterAuthed(rg *gin.RouterGroup, d *deps.Deps) {
	repo := NewRepository(d.DB)
	svc := NewService(repo, d.Issuer, d.Cfg, d.Audit, d.Log)
	h := &handler{svc: svc, repo: repo}

	authGroup := rg.Group("/auth")
	authGroup.POST("/logout", h.logout)
	authGroup.GET("/me", h.me)
	authGroup.GET("/devices", h.devices)
	authGroup.PUT("/devices/me", h.updateDevice)
	authGroup.POST("/devices/:deviceID/revoke", h.revokeDevice)
	authGroup.POST("/password/change", h.changePassword)
}

type handler struct {
	svc  *Service
	repo Repository
}

type loginRequest struct {
	Identifier string      `json:"identifier" binding:"required"`
	Password   string      `json:"password" binding:"required"`
	Device     DeviceInput `json:"device"`
}

func (h *handler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	res, err := h.svc.Login(c.Request.Context(), req.Identifier, req.Password, req.Device)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			httpx.WriteError(c, httpx.NewError(http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email/phone or password"))
			return
		}
		if errors.Is(err, ErrUserSuspended) {
			httpx.WriteError(c, httpx.NewError(http.StatusForbidden, "USER_SUSPENDED", "account is suspended"))
			return
		}
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, res)
}

type otpRequest struct {
	SchoolID uuid.UUID `json:"school_id" binding:"required"`
	Phone    string    `json:"phone" binding:"required,min=10,max=15"`
	Purpose  string    `json:"purpose" binding:"required,oneof=parent_login password_reset"`
}

func (h *handler) requestOtp(c *gin.Context) {
	var req otpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	if err := h.svc.RequestOtp(c.Request.Context(), req.SchoolID, req.Phone, req.Purpose); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, gin.H{"expires_in": 300})
}

type otpVerifyRequest struct {
	SchoolID uuid.UUID   `json:"school_id" binding:"required"`
	Phone    string      `json:"phone" binding:"required,min=10,max=15"`
	Purpose  string      `json:"purpose" binding:"required,oneof=parent_login password_reset"`
	OTP      string      `json:"otp" binding:"required,len=6"`
	Device   DeviceInput `json:"device"`
}

func (h *handler) verifyOtp(c *gin.Context) {
	var req otpVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	res, err := h.svc.VerifyOtp(c.Request.Context(), req.SchoolID, req.Phone, req.Purpose, req.OTP, req.Device)
	if err != nil {
		if errors.Is(err, ErrInvalidOtp) {
			httpx.WriteError(c, httpx.NewError(http.StatusUnauthorized, "INVALID_OTP", "invalid or expired OTP"))
			return
		}
		if errors.Is(err, ErrPasswordResetUnsupported) {
			httpx.WriteError(c, httpx.NewError(http.StatusNotImplemented, "PASSWORD_RESET_NOT_IMPLEMENTED", "password reset is not available yet"))
			return
		}
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, res)
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	DeviceID     string `json:"device_id" binding:"required"`
}

func (h *handler) refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	res, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken, req.DeviceID)
	if err != nil {
		if errors.Is(err, ErrInvalidRefresh) {
			httpx.WriteError(c, httpx.NewError(http.StatusUnauthorized, "TOKEN_REVOKED", "refresh token is invalid or revoked"))
			return
		}
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, res)
}

type logoutRequest struct {
	SessionID uuid.UUID `json:"session_id"`
	DeviceID  string    `json:"device_id"`
}

func (h *handler) logout(c *gin.Context) {
	var req logoutRequest
	_ = c.ShouldBindJSON(&req)
	uid := ctxuser.MustUserID(c.Request.Context())
	if err := h.svc.Logout(c.Request.Context(), uid, req.SessionID, req.DeviceID); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, gin.H{"logged_out": true})
}

func (h *handler) me(c *gin.Context) {
	uid := ctxuser.MustUserID(c.Request.Context())
	user, err := h.svc.Me(c.Request.Context(), uid)
	if err != nil {
		httpx.WriteError(c, httpx.ErrNotFound)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, user)
}

func (h *handler) devices(c *gin.Context) {
	uid := ctxuser.MustUserID(c.Request.Context())
	schoolID := ctxuser.MustSchoolID(c.Request.Context())
	devices, err := h.repo.ListDevices(c.Request.Context(), schoolID, uid)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, devices)
}

type updateDeviceRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
	Name     string `json:"name"`
	FCMToken string `json:"fcm_token"`
}

func (h *handler) updateDevice(c *gin.Context) {
	var req updateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	uid := ctxuser.MustUserID(c.Request.Context())
	schoolID := ctxuser.MustSchoolID(c.Request.Context())
	err := h.repo.UpsertDevice(c.Request.Context(), &UserDevice{
		SchoolID: schoolID, UserID: uid, DeviceID: req.DeviceID,
		DeviceName: req.Name, FCMToken: req.FCMToken,
	})
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, gin.H{"updated": true})
}

func (h *handler) revokeDevice(c *gin.Context) {
	deviceID := c.Param("deviceID")
	uid := ctxuser.MustUserID(c.Request.Context())
	schoolID := ctxuser.MustSchoolID(c.Request.Context())
	if err := h.repo.RevokeDevice(c.Request.Context(), schoolID, uid, deviceID); err != nil {
		httpx.WriteError(c, err)
		return
	}
	_ = h.repo.RevokeSessionsByDevice(c.Request.Context(), &schoolID, uid, deviceID)
	httpx.WriteJSON(c, http.StatusOK, gin.H{"revoked": true})
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=8"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func (h *handler) changePassword(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	uid := ctxuser.MustUserID(c.Request.Context())
	if err := h.svc.ChangePassword(c.Request.Context(), uid, req.OldPassword, req.NewPassword); err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			httpx.WriteError(c, httpx.NewError(http.StatusBadRequest, "INVALID_PASSWORD", "current password is incorrect"))
			return
		}
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, gin.H{"changed": true})
}
