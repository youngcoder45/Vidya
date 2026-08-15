package notifications

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository is the persistence port for the notifications module.
type Repository interface {
	Create(ctx context.Context, n *Notification) error
	ListForUser(ctx context.Context, schoolID, userID uuid.UUID, unreadOnly bool, limit int) ([]Notification, error)
	MarkRead(ctx context.Context, schoolID, userID, id uuid.UUID) error
	MarkAllRead(ctx context.Context, schoolID, userID uuid.UUID) error
	ListPreferences(ctx context.Context, schoolID, userID uuid.UUID) ([]NotificationPreference, error)
	UpsertPreference(ctx context.Context, p *NotificationPreference) error
}

// GormRepo implements Repository.
type GormRepo struct{ db *gorm.DB }

// NewRepository creates the notifications repository.
func NewRepository(db *gorm.DB) *GormRepo { return &GormRepo{db: db} }

func (r *GormRepo) Create(ctx context.Context, n *Notification) error { return r.db.WithContext(ctx).Create(n).Error }

func (r *GormRepo) ListForUser(ctx context.Context, schoolID, userID uuid.UUID, unreadOnly bool, limit int) ([]Notification, error) {
	q := r.db.WithContext(ctx).Where("school_id = ? AND user_id = ?", schoolID, userID)
	if unreadOnly {
		q = q.Where("read_at IS NULL")
	}
	if limit <= 0 {
		limit = 50
	}
	var out []Notification
	err := q.Order("created_at DESC").Limit(limit).Find(&out).Error
	return out, err
}

func (r *GormRepo) MarkRead(ctx context.Context, schoolID, userID, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&Notification{}).
		Where("school_id = ? AND user_id = ? AND id = ? AND read_at IS NULL", schoolID, userID, id).
		Update("read_at", gorm.Expr("NOW()")).Error
}

func (r *GormRepo) MarkAllRead(ctx context.Context, schoolID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&Notification{}).
		Where("school_id = ? AND user_id = ? AND read_at IS NULL", schoolID, userID).
		Update("read_at", gorm.Expr("NOW()")).Error
}

func (r *GormRepo) ListPreferences(ctx context.Context, schoolID, userID uuid.UUID) ([]NotificationPreference, error) {
	var out []NotificationPreference
	err := r.db.WithContext(ctx).Where("school_id = ? AND user_id = ?", schoolID, userID).Find(&out).Error
	return out, err
}

func (r *GormRepo) UpsertPreference(ctx context.Context, p *NotificationPreference) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "school_id"}, {Name: "user_id"}, {Name: "event_type"}, {Name: "channel"}},
		DoUpdates: clause.AssignmentColumns([]string{"enabled"}),
	}).Create(p).Error
}
