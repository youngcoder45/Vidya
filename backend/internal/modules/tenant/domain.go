package tenant

import (
	"time"

	"github.com/google/uuid"
)

// SchoolStatus values.
const (
	SchoolActive    = "active"
	SchoolSuspended = "suspended"
)

// Plan is a subscription tier (platform-level).
type Plan struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name         string    `gorm:"size:60;not null" json:"name"`
	PriceINR     int64     `json:"price_inr"`
	MaxStudents  int       `json:"max_students"`
	MaxTeachers  int       `json:"max_teachers"`
	FeatureFlags string    `gorm:"type:jsonb" json:"feature_flags,omitempty"` // json text
	CreatedAt    time.Time `json:"created_at"`
}

// School is a tenant. Branding holds per-school theme config (JSON text).
type School struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	PlanID     *uuid.UUID `gorm:"type:uuid" json:"plan_id,omitempty"`
	Name       string     `gorm:"size:160;not null" json:"name"`
	Code       string     `gorm:"size:40;uniqueIndex;not null" json:"code"`
	Board      string     `gorm:"size:60" json:"board,omitempty"`
	Address    string     `gorm:"size:300" json:"address,omitempty"`
	Phone      string     `gorm:"size:20" json:"phone,omitempty"`
	Email      string     `gorm:"size:160" json:"email,omitempty"`
	Timezone   string     `gorm:"size:40;default:Asia/Kolkata" json:"timezone"`
	Currency   string     `gorm:"size:8;default:INR" json:"currency"`
	Branding   string     `gorm:"type:jsonb" json:"branding,omitempty"` // {primary_color, logo_url, letterhead}
	Status     string     `gorm:"size:20;not null;default:active" json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// AcademicSession anchors all session-scoped data (April–March default).
type AcademicSession struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID      uuid.UUID `gorm:"type:uuid;not null;index" json:"school_id"`
	Name          string    `gorm:"size:40;not null" json:"name"`
	StartsOn      time.Time `json:"starts_on"`
	EndsOn        time.Time `json:"ends_on"`
	IsActive      bool      `gorm:"default:false" json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

// ClassGroup is a stable class ("Class 6") across sessions.
type ClassGroup struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID  uuid.UUID `gorm:"type:uuid;not null;index" json:"school_id"`
	Name      string    `gorm:"size:60;not null" json:"name"`
	Level     int       `json:"level"`
	CreatedAt time.Time `json:"created_at"`
}

// Division is a section ("A", "B").
type Division struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID  uuid.UUID `gorm:"type:uuid;not null;index" json:"school_id"`
	Name      string    `gorm:"size:20;not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// ClassDivision is the concrete session-scoped class+section taught by a
// class teacher. This is the anchor for rosters, attendance, homework, fees.
type ClassDivision struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID       uuid.UUID  `gorm:"type:uuid;not null;index:idx_cd_session" json:"school_id"`
	SessionID      uuid.UUID  `gorm:"type:uuid;not null;index:idx_cd_session" json:"session_id"`
	ClassGroupID   uuid.UUID  `gorm:"type:uuid;not null" json:"class_group_id"`
	DivisionID     uuid.UUID  `gorm:"type:uuid;not null" json:"division_id"`
	ClassTeacherID *uuid.UUID `gorm:"type:uuid" json:"class_teacher_id,omitempty"`
	Strength       int        `json:"strength"`
	CreatedAt      time.Time  `json:"created_at"`
}

// Subject is a subject catalog entry.
type Subject struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID   uuid.UUID `gorm:"type:uuid;not null;index" json:"school_id"`
	Name       string    `gorm:"size:80;not null" json:"name"`
	Code       string    `gorm:"size:20" json:"code,omitempty"`
	IsLanguage bool      `gorm:"default:false" json:"is_language"`
	CreatedAt  time.Time `json:"created_at"`
}

// ClassSubject assigns a subject to a class and optionally a teacher — the
// teacher's authorization scope.
type ClassSubject struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"school_id"`
	SessionID      uuid.UUID  `gorm:"type:uuid;not null" json:"session_id"`
	ClassDivisionID uuid.UUID `gorm:"type:uuid;not null;index" json:"class_division_id"`
	SubjectID      uuid.UUID  `gorm:"type:uuid;not null" json:"subject_id"`
	TeacherID      *uuid.UUID `gorm:"type:uuid;index" json:"teacher_id,omitempty"`
	IsElective     bool       `gorm:"default:false" json:"is_elective"`
	CreatedAt      time.Time  `json:"created_at"`
}

// TeacherStatus values.
const TeacherActive = "active"

// Teacher is the staff record linked to a user account.
type Teacher struct {
	ID                    uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID              uuid.UUID  `gorm:"type:uuid;not null;index" json:"school_id"`
	UserID                uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	EmployeeCode          string     `gorm:"size:40;not null" json:"employee_code"`
	Designation           string     `gorm:"size:80" json:"designation,omitempty"`
	JoinDate              *time.Time `json:"join_date,omitempty"`
	Qualification         string     `gorm:"size:160" json:"qualification,omitempty"`
	BankDetailsEncrypted  string     `gorm:"size:500" json:"-"`
	Status                string     `gorm:"size:20;not null;default:active" json:"status"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}
