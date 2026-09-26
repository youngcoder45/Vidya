// Package audit records append-only audit entries for sensitive operations
// (logins, fee/payment ops, result publish, payroll, student edits, RBAC).
package audit

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Entry is an audit log row. Kept append-only: no updates, no deletes.
type Entry struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID   *uuid.UUID `gorm:"type:uuid;index" json:"school_id,omitempty"`
	UserID     *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Action     string     `gorm:"size:64;not null" json:"action"`
	EntityType string     `gorm:"size:64;not null" json:"entity_type"`
	EntityID   string     `gorm:"size:64" json:"entity_id,omitempty"`
	Changes    string     `gorm:"type:jsonb" json:"changes,omitempty"`
	IP         string     `gorm:"size:64" json:"ip,omitempty"`
	UserAgent  string     `gorm:"size:255" json:"user_agent,omitempty"`
	DeviceID   string     `gorm:"size:128" json:"device_id,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// TableName keeps the plural convention.
func (Entry) TableName() string { return "audit_logs" }

// Writer persists audit entries.
type Writer struct{ db *gorm.DB }

// NewWriter creates an audit writer.
func NewWriter(db *gorm.DB) *Writer { return &Writer{db: db} }

// Record inserts an audit entry. Failures are logged and swallowed so audit
// problems never break the primary operation.
func (w *Writer) Record(ctx context.Context, e Entry) {
	if w == nil || w.db == nil {
		return
	}
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	e.CreatedAt = time.Now().UTC()
	if err := w.db.WithContext(ctx).Create(&e).Error; err != nil {
		// best-effort: never fail the request, but never hide the failure either
		slog.Default().Warn("audit write failed", "action", e.Action, "err", err)
	}
}
