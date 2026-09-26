package attendance

import (
	"time"

	"github.com/google/uuid"
)

// Attendance status values.
const (
	StatusPresent = "present"
	StatusAbsent  = "absent"
	StatusLate    = "late"
	StatusHalfDay = "half_day"
	StatusLeave   = "leave"
)

// ValidStatuses lists accepted status values.
var ValidStatuses = map[string]bool{
	StatusPresent: true, StatusAbsent: true, StatusLate: true,
	StatusHalfDay: true, StatusLeave: true,
}

// AttendanceRecord is one student's attendance for one date in a
// class/division (subject-attendance is optional in v1).
type AttendanceRecord struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID        uuid.UUID  `gorm:"type:uuid;not null;index:idx_att_class_date;uniqueIndex:idx_att_daily,priority:1" json:"school_id"`
	ClassDivisionID uuid.UUID  `gorm:"type:uuid;not null;index:idx_att_class_date;uniqueIndex:idx_att_daily,priority:2" json:"class_division_id"`
	SubjectID       *uuid.UUID `gorm:"type:uuid" json:"subject_id,omitempty"`
	StudentID       uuid.UUID  `gorm:"type:uuid;not null;index:idx_att_student;uniqueIndex:idx_att_daily,priority:3" json:"student_id"`
	Date            time.Time  `gorm:"index:idx_att_class_date;uniqueIndex:idx_att_daily,priority:4" json:"date"`
	Status          string     `gorm:"size:20;not null" json:"status"`
	MarkedBy        uuid.UUID  `gorm:"type:uuid;not null" json:"marked_by"`
	EditedBy        *uuid.UUID `gorm:"type:uuid" json:"edited_by,omitempty"`
	EditedAt        *time.Time `json:"edited_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
