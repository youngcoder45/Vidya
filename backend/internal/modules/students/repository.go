package students

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/schoolos/backend/internal/pkg/httpx"
)

// Repository is the persistence port for the students module.
type Repository interface {
	CreateStudentFull(ctx context.Context, s *Student, en *StudentEnrollment, guardians []Guardian, links []StudentGuardian) error
	GetStudent(ctx context.Context, schoolID, id uuid.UUID) (*Student, error)
	ListStudents(ctx context.Context, schoolID, sessionID, classDivisionID uuid.UUID, search string, page, limit int) ([]Student, int64, error)
	UpdateStudent(ctx context.Context, s *Student) error
	SoftDeleteStudent(ctx context.Context, schoolID, id uuid.UUID) error

	CreateGuardian(ctx context.Context, g *Guardian) error
	ListGuardiansByStudent(ctx context.Context, schoolID, studentID uuid.UUID) ([]Guardian, error)
	LinkGuardian(ctx context.Context, schoolID, studentID, guardianID uuid.UUID, relation string, priority int) error

	Enroll(ctx context.Context, en *StudentEnrollment) error
	GetActiveEnrollment(ctx context.Context, schoolID, studentID uuid.UUID) (*StudentEnrollment, error)
	ListRoster(ctx context.Context, schoolID, sessionID, classDivisionID uuid.UUID) ([]StudentEnrollment, error)
	Promote(ctx context.Context, schoolID, studentID, fromSessionID, toSessionID, toClassDivisionID uuid.UUID, rollNo int) error
	ListEnrollmentHistory(ctx context.Context, schoolID, studentID uuid.UUID) ([]StudentEnrollment, error)

	CreateDocument(ctx context.Context, d *StudentDocument) error
	ListDocuments(ctx context.Context, schoolID, studentID uuid.UUID) ([]StudentDocument, error)
}

// GormRepo implements Repository.
type GormRepo struct{ db *gorm.DB }

// NewRepository creates the students repository.
func NewRepository(db *gorm.DB) *GormRepo { return &GormRepo{db: db} }

func (r *GormRepo) CreateStudentFull(ctx context.Context, s *Student, en *StudentEnrollment, guardians []Guardian, links []StudentGuardian) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(s).Error; err != nil {
			return err
		}
		if en != nil {
			if err := tx.Create(en).Error; err != nil {
				return err
			}
		}
		for i := range guardians {
			if err := tx.Create(&guardians[i]).Error; err != nil {
				return err
			}
			links[i].GuardianID = guardians[i].ID
			if err := tx.Create(&links[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *GormRepo) GetStudent(ctx context.Context, schoolID, id uuid.UUID) (*Student, error) {
	var s Student
	if err := r.db.WithContext(ctx).First(&s, "school_id = ? AND id = ?", schoolID, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *GormRepo) ListStudents(ctx context.Context, schoolID, sessionID, classDivisionID uuid.UUID, search string, page, limit int) ([]Student, int64, error) {
	q := r.db.WithContext(ctx).Model(&Student{}).Where("students.school_id = ?", schoolID)
	if sessionID != uuid.Nil || classDivisionID != uuid.Nil {
		q = q.Joins("JOIN student_enrollments en ON en.student_id = students.id")
		if sessionID != uuid.Nil {
			q = q.Where("en.session_id = ?", sessionID)
		}
		if classDivisionID != uuid.Nil {
			q = q.Where("en.class_division_id = ?", classDivisionID)
		}
	}
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("students.first_name ILIKE ? OR students.last_name ILIKE ? OR students.admission_no ILIKE ?", like, like, like)
	}
	var total int64
	// Count distinct students: the enrollment join can duplicate rows.
	if err := q.Session(&gorm.Session{}).Distinct("students.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var out []Student
	err := q.Distinct("students.*").Order("students.first_name ASC").
		Offset((page - 1) * limit).Limit(limit).Find(&out).Error
	return out, total, err
}

func (r *GormRepo) UpdateStudent(ctx context.Context, s *Student) error {
	return r.db.WithContext(ctx).Model(s).Updates(map[string]any{
		"admission_no": s.AdmissionNo, "first_name": s.FirstName, "last_name": s.LastName,
		"dob": s.DOB, "gender": s.Gender, "blood_group": s.BloodGroup,
		"address": s.Address, "photo_url": s.PhotoURL, "medical_notes": s.MedicalNotes,
		"status": s.Status,
	}).Error
}

func (r *GormRepo) SoftDeleteStudent(ctx context.Context, schoolID, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&Student{}).
		Where("school_id = ? AND id = ?", schoolID, id).
		Update("status", StudentDropped).Error
}

func (r *GormRepo) CreateGuardian(ctx context.Context, g *Guardian) error { return r.db.WithContext(ctx).Create(g).Error }

func (r *GormRepo) ListGuardiansByStudent(ctx context.Context, schoolID, studentID uuid.UUID) ([]Guardian, error) {
	var out []Guardian
	err := r.db.WithContext(ctx).
		Joins("JOIN student_guardians sg ON sg.guardian_id = guardians.id").
		Where("guardians.school_id = ? AND sg.student_id = ?", schoolID, studentID).
		Order("guardians.is_primary DESC").Find(&out).Error
	return out, err
}

func (r *GormRepo) LinkGuardian(ctx context.Context, schoolID, studentID, guardianID uuid.UUID, relation string, priority int) error {
	link := &StudentGuardian{
		ID: uuid.New(), SchoolID: schoolID, StudentID: studentID,
		GuardianID: guardianID, Relation: relation, Priority: priority,
	}
	return r.db.WithContext(ctx).Create(link).Error
}

func (r *GormRepo) Enroll(ctx context.Context, en *StudentEnrollment) error { return r.db.WithContext(ctx).Create(en).Error }

func (r *GormRepo) GetActiveEnrollment(ctx context.Context, schoolID, studentID uuid.UUID) (*StudentEnrollment, error) {
	var en StudentEnrollment
	err := r.db.WithContext(ctx).
		Where("school_id = ? AND student_id = ? AND status = ?", schoolID, studentID, EnrollmentActive).
		First(&en).Error
	return &en, err
}

func (r *GormRepo) ListRoster(ctx context.Context, schoolID, sessionID, classDivisionID uuid.UUID) ([]StudentEnrollment, error) {
	var out []StudentEnrollment
	err := r.db.WithContext(ctx).
		Where("school_id = ? AND session_id = ? AND class_division_id = ? AND status = ?",
			schoolID, sessionID, classDivisionID, EnrollmentActive).
		Order("roll_no ASC").Find(&out).Error
	return out, err
}

func (r *GormRepo) Promote(ctx context.Context, schoolID, studentID, fromSessionID, toSessionID, toClassDivisionID uuid.UUID, rollNo int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&StudentEnrollment{}).
			Where("school_id = ? AND student_id = ? AND session_id = ?", schoolID, studentID, fromSessionID).
			Update("status", EnrollmentPromoted).Error; err != nil {
			return err
		}
		en := &StudentEnrollment{
			ID: uuid.New(), SchoolID: schoolID, SessionID: toSessionID,
			StudentID: studentID, ClassDivisionID: toClassDivisionID,
			RollNo: rollNo, AdmissionDate: time.Now(), Status: EnrollmentActive,
		}
		return tx.Create(en).Error
	})
}

func (r *GormRepo) ListEnrollmentHistory(ctx context.Context, schoolID, studentID uuid.UUID) ([]StudentEnrollment, error) {
	var out []StudentEnrollment
	err := r.db.WithContext(ctx).
		Where("school_id = ? AND student_id = ?", schoolID, studentID).
		Order("admission_date DESC").Find(&out).Error
	return out, err
}

func (r *GormRepo) CreateDocument(ctx context.Context, d *StudentDocument) error { return r.db.WithContext(ctx).Create(d).Error }

func (r *GormRepo) ListDocuments(ctx context.Context, schoolID, studentID uuid.UUID) ([]StudentDocument, error) {
	var out []StudentDocument
	err := r.db.WithContext(ctx).
		Where("school_id = ? AND student_id = ?", schoolID, studentID).
		Order("created_at DESC").Find(&out).Error
	return out, err
}

// notFound maps gorm not-found to a 404 AppError.
func notFound() error { return httpx.ErrNotFound }
