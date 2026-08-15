package notifications

import (
	"time"

	"github.com/google/uuid"
)

// Channels.
const (
	ChannelInApp = "inapp"
	ChannelPush  = "push"
	ChannelEmail = "email"
	ChannelSMS   = "sms" // future
)

// Delivery status values.
const (
	DeliveryQueued = "queued"
	DeliverySent   = "sent"
	DeliveryFailed = "failed"
)

// Event types the notification service understands.
const (
	EvHomeworkAssigned   = "homework.assigned"
	EvAssignmentDeadline = "assignment.deadline"
	EvExamSchedule       = "exam.scheduled"
	EvResultsPublished   = "results.published"
	EvFeeDue             = "fees.due"
	EvFeePaid            = "fees.paid"
	EvAnnouncement       = "announcement.published"
	EvHoliday            = "holiday.announced"
	EvSalaryPaid         = "payroll.salary_paid"
)

// Notification is the in-app feed item (also the fan-out source for
// push/email deliveries).
type Notification struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID  uuid.UUID `gorm:"type:uuid;not null;index:idx_notif_user" json:"school_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_notif_user" json:"user_id"`
	Type      string    `gorm:"size:40;not null" json:"type"`
	Title     string    `gorm:"size:160;not null" json:"title"`
	Body      string    `gorm:"size:500" json:"body,omitempty"`
	Data      string    `gorm:"type:jsonb" json:"data,omitempty"`
	Channel   string    `gorm:"size:20;not null;default:inapp" json:"channel"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// NotificationDelivery records per-channel delivery attempts (idempotent by
// unique (school_id, notification_id, channel)).
type NotificationDelivery struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"school_id"`
	NotificationID uuid.UUID  `gorm:"type:uuid;not null;index:idx_delivery" json:"notification_id"`
	Channel        string     `gorm:"size:20;not null;index:idx_delivery" json:"channel"`
	Status         string     `gorm:"size:20;not null;default:queued" json:"status"`
	Attempts       int        `gorm:"default:0" json:"attempts"`
	LastError      string     `gorm:"size:300" json:"last_error,omitempty"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// NotificationPreference is per-user opt-in per event type per channel.
type NotificationPreference struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID  uuid.UUID `gorm:"type:uuid;not null;index" json:"school_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	EventType string    `gorm:"size:40;not null" json:"event_type"`
	Channel   string    `gorm:"size:20;not null" json:"channel"`
	Enabled   bool      `gorm:"default:true" json:"enabled"`
}

// NotificationTemplate renders message content per event/channel. SchoolID nil
// = platform default template.
type NotificationTemplate struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID     *uuid.UUID `gorm:"type:uuid;index" json:"school_id,omitempty"`
	EventType    string     `gorm:"size:40;not null" json:"event_type"`
	Channel      string     `gorm:"size:20;not null" json:"channel"`
	Subject      string     `gorm:"size:160" json:"subject,omitempty"`
	BodyTemplate string     `gorm:"type:text" json:"body_template"`
}
