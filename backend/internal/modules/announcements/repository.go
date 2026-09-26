package announcements

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository is the persistence port for the announcements module.
type Repository interface {
	Create(ctx context.Context, a *Announcement) error
	List(ctx context.Context, schoolID uuid.UUID, status string, publishedOnly bool) ([]Announcement, error)
	Get(ctx context.Context, schoolID, id uuid.UUID) (*Announcement, error)
	SetStatus(ctx context.Context, schoolID, id uuid.UUID, status string) error
	MarkRead(ctx context.Context, schoolID, announcementID, userID uuid.UUID) error
	ReadCount(ctx context.Context, schoolID, announcementID uuid.UUID) (int64, error)

	CreateEvent(ctx context.Context, e *HolidayEvent) error
	ListEvents(ctx context.Context, schoolID, sessionID uuid.UUID) ([]HolidayEvent, error)
}

// GormRepo implements Repository.
type GormRepo struct{ db *gorm.DB }

// NewRepository creates the announcements repository.
func NewRepository(db *gorm.DB) *GormRepo { return &GormRepo{db: db} }

func (r *GormRepo) Create(ctx context.Context, a *Announcement) error { return r.db.WithContext(ctx).Create(a).Error }

func (r *GormRepo) List(ctx context.Context, schoolID uuid.UUID, status string, publishedOnly bool) ([]Announcement, error) {
	q := r.db.WithContext(ctx).Where("school_id = ?", schoolID)
	if publishedOnly {
		// Readers only ever see live, published announcements.
		q = q.Where("status = ?", StatusPublished).
			Where("(expires_at IS NULL OR expires_at > ?)", time.Now())
	} else if status != "" {
		q = q.Where("status = ?", status)
	}
	var out []Announcement
	err := q.Order("created_at DESC").Find(&out).Error
	return out, err
}

func (r *GormRepo) Get(ctx context.Context, schoolID, id uuid.UUID) (*Announcement, error) {
	var a Announcement
	if err := r.db.WithContext(ctx).First(&a, "school_id = ? AND id = ?", schoolID, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *GormRepo) SetStatus(ctx context.Context, schoolID, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&Announcement{}).
		Where("school_id = ? AND id = ?", schoolID, id).
		Updates(map[string]any{"status": status, "publish_at": time.Now()}).Error
}

func (r *GormRepo) MarkRead(ctx context.Context, schoolID, announcementID, userID uuid.UUID) error {
	// Upsert read receipt (unique school+announcement+user): repeated reads no-op.
	rec := &AnnouncementRead{
		ID: uuid.New(), SchoolID: schoolID, AnnouncementID: announcementID,
		UserID: userID, ReadAt: time.Now(),
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(rec).Error
}

func (r *GormRepo) ReadCount(ctx context.Context, schoolID, announcementID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&AnnouncementRead{}).
		Where("school_id = ? AND announcement_id = ?", schoolID, announcementID).
		Count(&n).Error
	return n, err
}

func (r *GormRepo) CreateEvent(ctx context.Context, e *HolidayEvent) error { return r.db.WithContext(ctx).Create(e).Error }

func (r *GormRepo) ListEvents(ctx context.Context, schoolID, sessionID uuid.UUID) ([]HolidayEvent, error) {
	q := r.db.WithContext(ctx).Where("school_id = ?", schoolID)
	if sessionID != uuid.Nil {
		q = q.Where("session_id = ?", sessionID)
	}
	var out []HolidayEvent
	err := q.Order("starts_on ASC").Find(&out).Error
	return out, err
}
