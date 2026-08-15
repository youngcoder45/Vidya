package auth

import (
	"time"

	"github.com/google/uuid"
)

// UserStatus values.
const (
	UserActive  = "active"
	UserSuspended = "suspended"
)

// User is the auth principal. Every user belongs to one school (platform
// admins use SchoolID = nil).
type User struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID     *uuid.UUID `gorm:"type:uuid;index" json:"school_id"`
	FullName     string     `gorm:"size:120;not null" json:"full_name"`
	Email        string     `gorm:"size:160;index" json:"email"`
	Phone        string     `gorm:"size:20;index" json:"phone"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	AvatarURL    string     `gorm:"size:500" json:"avatar_url,omitempty"`
	Status       string     `gorm:"size:20;not null;default:active" json:"status"`
	IsSuperadmin bool       `gorm:"default:false" json:"is_superadmin"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Role is a named set of permissions (shared catalog across tenants).
type Role struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Code        string    `gorm:"size:40;uniqueIndex;not null" json:"code"`
	Name        string    `gorm:"size:80;not null" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
}

// Permission is an atomic capability, e.g. "fees.payment.create".
type Permission struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Code   string    `gorm:"size:80;uniqueIndex;not null" json:"code"`
	Name   string    `gorm:"size:120;not null" json:"name"`
	Module string    `gorm:"size:40;index" json:"module"`
}

// RolePermission links roles to permissions.
type RolePermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"role_id"`
	PermissionID uuid.UUID `gorm:"type:uuid;primaryKey" json:"permission_id"`
}

// UserRole grants a role to a user within a school. ScopeClassDivisionID
// optionally narrows the grant (e.g., teacher only for Class 6-A).
type UserRole struct {
	ID                    uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID              uuid.UUID  `gorm:"type:uuid;not null;index" json:"school_id"`
	UserID                uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	RoleID                uuid.UUID  `gorm:"type:uuid;not null" json:"role_id"`
	ScopeClassDivisionID  *uuid.UUID `gorm:"type:uuid" json:"scope_class_division_id,omitempty"`
}

// UserDevice tracks a device for push targets and session revocation.
type UserDevice struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"school_id"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	DeviceID   string     `gorm:"size:128;not null" json:"device_id"`
	DeviceName string     `gorm:"size:120" json:"device_name"`
	Platform   string     `gorm:"size:20" json:"platform"`
	FCMToken   string     `gorm:"size:255" json:"fcm_token,omitempty"`
	LastSeenAt time.Time  `json:"last_seen_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}

// AuthSession stores hashed refresh tokens with rotation tracking.
type AuthSession struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID          uuid.UUID  `gorm:"type:uuid;not null;index" json:"school_id"`
	UserID            uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	DeviceID          string     `gorm:"size:128" json:"device_id"`
	RefreshTokenHash  string     `gorm:"size:128;not null" json:"-"`
	IP                string     `gorm:"size:64" json:"ip,omitempty"`
	UserAgent         string     `gorm:"size:255" json:"user_agent,omitempty"`
	ExpiresAt         time.Time  `json:"expires_at"`
	RevokedAt         *time.Time `json:"revoked_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

// OtpPurpose values.
const (
	OtpPurposeParentLogin = "parent_login"
	OtpPurposePasswordReset = "password_reset"
)

// OtpCode stores hashed OTPs with attempt limits.
type OtpCode struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"school_id"`
	Phone     string     `gorm:"size:20;index" json:"phone"`
	Email     string     `gorm:"size:160" json:"email"`
	Purpose   string     `gorm:"size:40;not null" json:"purpose"`
	CodeHash  string     `gorm:"size:128;not null" json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	Attempts  int        `gorm:"default:0" json:"attempts"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
