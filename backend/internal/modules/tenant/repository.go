package tenant

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/schoolos/backend/internal/pkg/httpx"
)

// ErrNotFound maps to a 404.
var ErrNotFound = httpx.ErrNotFound

// Repository is the persistence port for the tenant module.
type Repository interface {
	CreateSchool(ctx context.Context, s *School) error
	GetSchool(ctx context.Context, id uuid.UUID) (*School, error)
	ListSchools(ctx context.Context) ([]School, error)

	CreateSession(ctx context.Context, s *AcademicSession) error
	ListSessions(ctx context.Context, schoolID uuid.UUID) ([]AcademicSession, error)
	GetSession(ctx context.Context, schoolID, id uuid.UUID) (*AcademicSession, error)
	ActivateSession(ctx context.Context, schoolID, id uuid.UUID) error

	CreateClassGroup(ctx context.Context, c *ClassGroup) error
	ListClassGroups(ctx context.Context, schoolID uuid.UUID) ([]ClassGroup, error)
	CreateDivision(ctx context.Context, d *Division) error
	ListDivisions(ctx context.Context, schoolID uuid.UUID) ([]Division, error)

	CreateClassDivision(ctx context.Context, cd *ClassDivision) error
	ListClassDivisions(ctx context.Context, schoolID, sessionID uuid.UUID) ([]ClassDivision, error)
	GetClassDivision(ctx context.Context, schoolID, id uuid.UUID) (*ClassDivision, error)
	SetClassTeacher(ctx context.Context, schoolID, id, teacherID uuid.UUID) error

	CreateSubject(ctx context.Context, s *Subject) error
	ListSubjects(ctx context.Context, schoolID uuid.UUID) ([]Subject, error)
	CreateClassSubject(ctx context.Context, cs *ClassSubject) error
	ListClassSubjects(ctx context.Context, schoolID uuid.UUID) ([]ClassSubject, error)

	CreateTeacher(ctx context.Context, t *Teacher) error
	GetTeacher(ctx context.Context, schoolID, id uuid.UUID) (*Teacher, error)
	ListTeachers(ctx context.Context, schoolID uuid.UUID) ([]Teacher, error)
}

// GormRepo implements Repository.
type GormRepo struct{ db *gorm.DB }

// NewRepository creates the tenant repository.
func NewRepository(db *gorm.DB) *GormRepo { return &GormRepo{db: db} }

func (r *GormRepo) CreateSchool(ctx context.Context, s *School) error { return r.db.WithContext(ctx).Create(s).Error }
func (r *GormRepo) GetSchool(ctx context.Context, id uuid.UUID) (*School, error) {
	var s School
	if err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}
func (r *GormRepo) ListSchools(ctx context.Context) ([]School, error) {
	var schools []School
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&schools).Error
	return schools, err
}

func (r *GormRepo) CreateSession(ctx context.Context, s *AcademicSession) error { return r.db.WithContext(ctx).Create(s).Error }
func (r *GormRepo) ListSessions(ctx context.Context, schoolID uuid.UUID) ([]AcademicSession, error) {
	var sessions []AcademicSession
	err := r.db.WithContext(ctx).Where("school_id = ?", schoolID).Order("starts_on DESC").Find(&sessions).Error
	return sessions, err
}
func (r *GormRepo) GetSession(ctx context.Context, schoolID, id uuid.UUID) (*AcademicSession, error) {
	var s AcademicSession
	if err := r.db.WithContext(ctx).First(&s, "school_id = ? AND id = ?", schoolID, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}
func (r *GormRepo) ActivateSession(ctx context.Context, schoolID, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&AcademicSession{}).Where("school_id = ?", schoolID).Update("is_active", false).Error; err != nil {
			return err
		}
		return tx.Model(&AcademicSession{}).Where("school_id = ? AND id = ?", schoolID, id).Update("is_active", true).Error
	})
}

func (r *GormRepo) CreateClassGroup(ctx context.Context, c *ClassGroup) error { return r.db.WithContext(ctx).Create(c).Error }
func (r *GormRepo) ListClassGroups(ctx context.Context, schoolID uuid.UUID) ([]ClassGroup, error) {
	var out []ClassGroup
	err := r.db.WithContext(ctx).Where("school_id = ?", schoolID).Order("level ASC").Find(&out).Error
	return out, err
}
func (r *GormRepo) CreateDivision(ctx context.Context, d *Division) error { return r.db.WithContext(ctx).Create(d).Error }
func (r *GormRepo) ListDivisions(ctx context.Context, schoolID uuid.UUID) ([]Division, error) {
	var out []Division
	err := r.db.WithContext(ctx).Where("school_id = ?", schoolID).Order("name ASC").Find(&out).Error
	return out, err
}

func (r *GormRepo) CreateClassDivision(ctx context.Context, cd *ClassDivision) error {
	return r.db.WithContext(ctx).Create(cd).Error
}
func (r *GormRepo) ListClassDivisions(ctx context.Context, schoolID, sessionID uuid.UUID) ([]ClassDivision, error) {
	var out []ClassDivision
	err := r.db.WithContext(ctx).Where("school_id = ? AND session_id = ?", schoolID, sessionID).Find(&out).Error
	return out, err
}
func (r *GormRepo) GetClassDivision(ctx context.Context, schoolID, id uuid.UUID) (*ClassDivision, error) {
	var cd ClassDivision
	if err := r.db.WithContext(ctx).First(&cd, "school_id = ? AND id = ?", schoolID, id).Error; err != nil {
		return nil, err
	}
	return &cd, nil
}
func (r *GormRepo) SetClassTeacher(ctx context.Context, schoolID, id, teacherID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&ClassDivision{}).
		Where("school_id = ? AND id = ?", schoolID, id).Update("class_teacher_id", teacherID).Error
}

func (r *GormRepo) CreateSubject(ctx context.Context, s *Subject) error { return r.db.WithContext(ctx).Create(s).Error }
func (r *GormRepo) ListSubjects(ctx context.Context, schoolID uuid.UUID) ([]Subject, error) {
	var out []Subject
	err := r.db.WithContext(ctx).Where("school_id = ?", schoolID).Order("name ASC").Find(&out).Error
	return out, err
}
func (r *GormRepo) CreateClassSubject(ctx context.Context, cs *ClassSubject) error {
	return r.db.WithContext(ctx).Create(cs).Error
}
func (r *GormRepo) ListClassSubjects(ctx context.Context, schoolID uuid.UUID) ([]ClassSubject, error) {
	var out []ClassSubject
	err := r.db.WithContext(ctx).Where("school_id = ?", schoolID).Find(&out).Error
	return out, err
}

func (r *GormRepo) CreateTeacher(ctx context.Context, t *Teacher) error { return r.db.WithContext(ctx).Create(t).Error }
func (r *GormRepo) GetTeacher(ctx context.Context, schoolID, id uuid.UUID) (*Teacher, error) {
	var t Teacher
	if err := r.db.WithContext(ctx).First(&t, "school_id = ? AND id = ?", schoolID, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}
func (r *GormRepo) ListTeachers(ctx context.Context, schoolID uuid.UUID) ([]Teacher, error) {
	var out []Teacher
	err := r.db.WithContext(ctx).Where("school_id = ?", schoolID).Order("employee_code ASC").Find(&out).Error
	return out, err
}
