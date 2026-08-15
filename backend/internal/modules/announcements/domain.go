package announcements

import (
	"time"

	"github.com/google/uuid"
)

// Audience scopes.
const (
	AudienceSchool   = "school"
	AudienceClass    = "class"
	AudienceDivision = "division"
	AudienceStudent  = "student"
	AudienceStaff    = "staff"
)

// Announcement status values.
const (
	StatusDraft     = "draft"
	StatusScheduled = "scheduled"
	StatusPublished = "published"
	StatusArchived  = "archived"
)

// Announcement targets an audience scope (school/class/division/student/staff)
// and can be scheduled for future publication.
type Announcement struct {
	ID                      uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID                uuid.UUID  `gorm:"type:uuid;not null;index:idx_ann_status" json:"school_id"`
	Title                   string     `gorm:"size:160;not null" json:"title"`
	Body                    string     `gorm:"type:text;not null" json:"body"`
	AudienceScope           string     `gorm:"size:20;not null" json:"audience_scope"`
	AudienceClassDivisionID *uuid.UUID `gorm:"type:uuid" json:"audience_class_division_id,omitempty"`
	AudienceStudentID       *uuid.UUID `gorm:"type:uuid" json:"audience_student_id,omitempty"`
	PublishAt               *time.Time `json:"publish_at,omitempty"`
	ExpiresAt               *time.Time `json:"expires_at,omitempty"`
	Status                  string     `gorm:"size:20;not null;default:draft;index:idx_ann_status" json:"status"`
	CreatedBy               uuid.UUID  `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

// AnnouncementRead tracks read receipts per user.
type AnnouncementRead struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID       uuid.UUID `gorm:"type:uuid;not null;index" json:"school_id"`
	AnnouncementID uuid.UUID `gorm:"type:uuid;not null;index:idx_ann_read" json:"announcement_id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;index:idx_ann_read" json:"user_id"`
	ReadAt         time.Time `json:"read_at"`
}

// Event types for the school calendar.
const (
	EventHoliday   = "holiday"
	EventActivity  = "activity"
	EventEvent     = "event"
)

// HolidayEvent is a calendar entry (holiday / event / activity).
type HolidayEvent struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID      uuid.UUID `gorm:"type:uuid;not null;index:idx_ev_start" json:"school_id"`
	SessionID     uuid.UUID `gorm:"type:uuid;not null" json:"session_id"`
	Title         string    `gorm:"size:160;not null" json:"title"`
	Type          string    `gorm:"size:20;not null" json:"type"`
	StartsOn      time.Time `gorm:"index:idx_ev_start" json:"starts_on"`
	EndsOn        time.Time `json:"ends_on"`
	AudienceScope string    `gorm:"size:20;default:school" json:"audience_scope"`
	Description   string    `gorm:"size:500" json:"description,omitempty"`
	CreatedBy     uuid.UUID `gorm:"type:uuid" json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
}
