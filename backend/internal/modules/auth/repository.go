package auth

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository is the persistence port for the auth module.
type Repository interface {
	FindUserByIdentity(ctx context.Context, schoolID *uuid.UUID, identifier string) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	CreateUser(ctx context.Context, u *User) error
	UpdateUser(ctx context.Context, u *User) error
	FindRoleByCode(ctx context.Context, code string) (*Role, error)
	AssignRole(ctx context.Context, schoolID, userID, roleID uuid.UUID) error
	ListRolesByUser(ctx context.Context, schoolID, userID uuid.UUID) ([]Role, error)
	ListPermissionsByUser(ctx context.Context, schoolID, userID uuid.UUID) ([]Permission, error)

	CreateSession(ctx context.Context, s *AuthSession) error
	FindSessionByRefreshHash(ctx context.Context, hash string) (*AuthSession, error)
	RevokeSession(ctx context.Context, id uuid.UUID) error
	RevokeSessionsByDevice(ctx context.Context, schoolID, userID uuid.UUID, deviceID string) error

	UpsertDevice(ctx context.Context, d *UserDevice) error
	ListDevices(ctx context.Context, schoolID, userID uuid.UUID) ([]UserDevice, error)
	RevokeDevice(ctx context.Context, schoolID, userID uuid.UUID, deviceID string) error

	CreateOtp(ctx context.Context, o *OtpCode) error
	FindLatestOtp(ctx context.Context, schoolID uuid.UUID, phone, purpose string) (*OtpCode, error)
	MarkOtpVerified(ctx context.Context, id uuid.UUID) error
	IncrementOtpAttempts(ctx context.Context, id uuid.UUID) error
}

// GormRepo implements Repository with tenant-scoped queries.
type GormRepo struct{ db *gorm.DB }

// NewRepository creates the GORM-backed repository.
func NewRepository(db *gorm.DB) *GormRepo { return &GormRepo{db: db} }

func (r *GormRepo) FindUserByIdentity(ctx context.Context, schoolID *uuid.UUID, identifier string) (*User, error) {
	q := r.db.WithContext(ctx).Where("email = ? OR phone = ?", identifier, identifier)
	if schoolID != nil {
		q = q.Where("school_id = ?", *schoolID)
	}
	var u User
	if err := q.First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *GormRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var u User
	if err := r.db.WithContext(ctx).First(&u, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *GormRepo) CreateUser(ctx context.Context, u *User) error { return r.db.WithContext(ctx).Create(u).Error }

func (r *GormRepo) UpdateUser(ctx context.Context, u *User) error {
	return r.db.WithContext(ctx).Model(u).Updates(map[string]any{
		"full_name": u.FullName, "email": u.Email, "phone": u.Phone,
		"password_hash": u.PasswordHash, "avatar_url": u.AvatarURL,
		"status": u.Status, "last_login_at": u.LastLoginAt,
	}).Error
}

func (r *GormRepo) FindRoleByCode(ctx context.Context, code string) (*Role, error) {
	var role Role
	if err := r.db.WithContext(ctx).First(&role, "code = ?", code).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *GormRepo) AssignRole(ctx context.Context, schoolID, userID, roleID uuid.UUID) error {
	ur := UserRole{ID: uuid.New(), SchoolID: schoolID, UserID: userID, RoleID: roleID}
	return r.db.WithContext(ctx).Create(&ur).Error
}

func (r *GormRepo) ListRolesByUser(ctx context.Context, schoolID, userID uuid.UUID) ([]Role, error) {
	var roles []Role
	err := r.db.WithContext(ctx).
		Joins("JOIN user_roles ur ON ur.role_id = roles.id").
		Where("ur.school_id = ? AND ur.user_id = ?", schoolID, userID).
		Find(&roles).Error
	return roles, err
}

func (r *GormRepo) ListPermissionsByUser(ctx context.Context, schoolID, userID uuid.UUID) ([]Permission, error) {
	var perms []Permission
	err := r.db.WithContext(ctx).
		Joins("JOIN role_permissions rp ON rp.permission_id = permissions.id").
		Joins("JOIN user_roles ur ON ur.role_id = rp.role_id").
		Where("ur.school_id = ? AND ur.user_id = ?", schoolID, userID).
		Distinct("permissions.id", "permissions.code", "permissions.name", "permissions.module").
		Find(&perms).Error
	return perms, err
}

func (r *GormRepo) CreateSession(ctx context.Context, s *AuthSession) error { return r.db.WithContext(ctx).Create(s).Error }

func (r *GormRepo) FindSessionByRefreshHash(ctx context.Context, hash string) (*AuthSession, error) {
	var s AuthSession
	if err := r.db.WithContext(ctx).First(&s, "refresh_token_hash = ?", hash).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *GormRepo) RevokeSession(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&AuthSession{}).Where("id = ?", id).
		Update("revoked_at", gorm.Expr("NOW()")).Error
}

func (r *GormRepo) RevokeSessionsByDevice(ctx context.Context, schoolID, userID uuid.UUID, deviceID string) error {
	return r.db.WithContext(ctx).Model(&AuthSession{}).
		Where("school_id = ? AND user_id = ? AND device_id = ?", schoolID, userID, deviceID).
		Update("revoked_at", gorm.Expr("NOW()")).Error
}

func (r *GormRepo) UpsertDevice(ctx context.Context, d *UserDevice) error {
	var existing UserDevice
	err := r.db.WithContext(ctx).
		Where("school_id = ? AND user_id = ? AND device_id = ?", d.SchoolID, d.UserID, d.DeviceID).
		First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		d.ID = uuid.New()
		return r.db.WithContext(ctx).Create(d).Error
	}
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&existing).Updates(map[string]any{
		"device_name": d.DeviceName, "platform": d.Platform, "fcm_token": d.FCMToken,
		"last_seen_at": d.LastSeenAt, "revoked_at": nil,
	}).Error
}

func (r *GormRepo) ListDevices(ctx context.Context, schoolID, userID uuid.UUID) ([]UserDevice, error) {
	var devices []UserDevice
	err := r.db.WithContext(ctx).Where("school_id = ? AND user_id = ?", schoolID, userID).
		Order("last_seen_at DESC").Find(&devices).Error
	return devices, err
}

func (r *GormRepo) RevokeDevice(ctx context.Context, schoolID, userID uuid.UUID, deviceID string) error {
	return r.db.WithContext(ctx).Model(&UserDevice{}).
		Where("school_id = ? AND user_id = ? AND device_id = ?", schoolID, userID, deviceID).
		Update("revoked_at", gorm.Expr("NOW()")).Error
}

func (r *GormRepo) CreateOtp(ctx context.Context, o *OtpCode) error { return r.db.WithContext(ctx).Create(o).Error }

func (r *GormRepo) FindLatestOtp(ctx context.Context, schoolID uuid.UUID, phone, purpose string) (*OtpCode, error) {
	var o OtpCode
	err := r.db.WithContext(ctx).
		Where("school_id = ? AND phone = ? AND purpose = ?", schoolID, phone, purpose).
		Order("created_at DESC").First(&o).Error
	return &o, err
}

func (r *GormRepo) MarkOtpVerified(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&OtpCode{}).Where("id = ?", id).
		Update("verified_at", gorm.Expr("NOW()")).Error
}

func (r *GormRepo) IncrementOtpAttempts(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&OtpCode{}).Where("id = ?", id).
		UpdateColumn("attempts", gorm.Expr("attempts + 1")).Error
}
