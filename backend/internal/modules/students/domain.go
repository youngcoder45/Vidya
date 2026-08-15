package students

import (
	"time"

	"github.com/google/uuid"
)

// StudentStatus values.
const (
	StudentActive = "active"
	StudentDropped = "dropped"
)

// Student is the core student record (not session-scoped; enrollments are).
type Student struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID     uuid.UUID `gorm:"type:uuid;not null;index" json:"school_id"`
	AdmissionNo  string    `gorm:"size:40;not null" json:"admission_no"`
	FirstName    string    `gorm:"size:80;not null" json:"first_name"`
	LastName     string    `gorm:"size:80" json:"last_name,omitempty"`
	DOB          time.Time `json:"dob"`
	Gender       string    `gorm:"size:10" json:"gender,omitempty"`
	BloodGroup   string    `gorm:"size:8" json:"blood_group,omitempty"`
	Address      string    `gorm:"size:300" json:"address,omitempty"`
	PhotoURL     string    `gorm:"size:500" json:"photo_url,omitempty"`
	MedicalNotes string    `gorm:"size:500" json:"medical_notes,omitempty"`
	Status       string    `gorm:"size:20;not null;default:active" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Guardian is a parent/guardian record. Linked to a user account when the
// parent registers with OTP.
type Guardian struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"school_id"`
	UserID    *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Name      string     `gorm:"size:120;not null" json:"name"`
	Relation  string     `gorm:"size:20" json:"relation,omitempty"`
	Phone     string     `gorm:"size:20;not null" json:"phone"`
	Email     string     `gorm:"size:160" json:"email,omitempty"`
	Occupation string    `gorm:"size:80" json:"occupation,omitempty"`
	IsPrimary bool       `gorm:"default:false" json:"is_primary"`
	CreatedAt time.Time  `json:"created_at"`
}

// StudentGuardian links a student to a guardian.
type StudentGuardian struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID    uuid.UUID `gorm:"type:uuid;not null;index" json:"school_id"`
	StudentID   uuid.UUID `gorm:"type:uuid;not null;index" json:"student_id"`
	GuardianID  uuid.UUID `gorm:"type:uuid;not null" json:"guardian_id"`
	Relation    string    `gorm:"size:20" json:"relation,omitempty"`
	Priority    int       `gorm:"default:0" json:"priority"`
}

// EnrollmentStatus values.
const (
	EnrollmentActive   = "active"
	EnrollmentPromoted = "promoted"
	EnrollmentDropped  = "dropped"
)

// StudentEnrollment is the session-scoped membership of a student in a
// class/division with a roll number.
type StudentEnrollment struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID        uuid.UUID `gorm:"type:uuid;not null;index:idx_enr_session_class" json:"school_id"`
	SessionID       uuid.UUID `gorm:"type:uuid;not null;index:idx_enr_session_class" json:"session_id"`
	StudentID       uuid.UUID `gorm:"type:uuid;not null;index:idx_enr_student" json:"student_id"`
	ClassDivisionID uuid.UUID `gorm:"type:uuid;not null;index:idx_enr_session_class" json:"class_division_id"`
	RollNo          int       `json:"roll_no"`
	AdmissionDate   time.Time `json:"admission_date"`
	Status          string    `gorm:"size:20;not null;default:active" json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

// StudentDocument is an uploaded document reference (S3 object URL).
type StudentDocument struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"school_id"`
	StudentID uuid.UUID  `gorm:"type:uuid;not null;index" json:"student_id"`
	DocType   string     `gorm:"size:40;not null" json:"doc_type"`
	FileURL   string     `gorm:"size:500;not null" json:"file_url"`
	IssuedOn  *time.Time `json:"issued_on,omitempty"`
	ExpiresOn *time.Time `json:"expires_on,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
