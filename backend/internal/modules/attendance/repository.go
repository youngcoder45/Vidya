package attendance

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository is the persistence port for the attendance module.
type Repository interface {
	UpsertDaily(ctx context.Context, records []AttendanceRecord) error
	ListByClassDate(ctx context.Context, schoolID, classDivisionID uuid.UUID, date time.Time) ([]AttendanceRecord, error)
	ListByStudent(ctx context.Context, schoolID, studentID uuid.UUID, from, to time.Time) ([]AttendanceRecord, error)
	ClassDailyPercentage(ctx context.Context, schoolID, classDivisionID uuid.UUID, date time.Time) (present, total int64, err error)
}

// GormRepo implements Repository.
type GormRepo struct{ db *gorm.DB }

// NewRepository creates the attendance repository.
func NewRepository(db *gorm.DB) *GormRepo { return &GormRepo{db: db} }

func (r *GormRepo) UpsertDaily(ctx context.Context, records []AttendanceRecord) error {
	if len(records) == 0 {
		return nil
	}
	// One bulk upsert keyed by the natural key (school, class, student, date).
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "school_id"}, {Name: "class_division_id"},
			{Name: "student_id"}, {Name: "date"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"status", "subject_id", "edited_by", "edited_at", "updated_at"}),
	}).Create(&records).Error
}

func (r *GormRepo) ListByClassDate(ctx context.Context, schoolID, classDivisionID uuid.UUID, date time.Time) ([]AttendanceRecord, error) {
	var out []AttendanceRecord
	err := r.db.WithContext(ctx).
		Where("school_id = ? AND class_division_id = ? AND date = ?", schoolID, classDivisionID, date).
		Order("student_id ASC").Find(&out).Error
	return out, err
}

func (r *GormRepo) ListByStudent(ctx context.Context, schoolID, studentID uuid.UUID, from, to time.Time) ([]AttendanceRecord, error) {
	var out []AttendanceRecord
	err := r.db.WithContext(ctx).
		Where("school_id = ? AND student_id = ? AND date BETWEEN ? AND ?", schoolID, studentID, from, to).
		Order("date ASC").Find(&out).Error
	return out, err
}

func (r *GormRepo) ClassDailyPercentage(ctx context.Context, schoolID, classDivisionID uuid.UUID, date time.Time) (present, total int64, err error) {
	err = r.db.WithContext(ctx).Model(&AttendanceRecord{}).
		Where("school_id = ? AND class_division_id = ? AND date = ?", schoolID, classDivisionID, date).
		Count(&total).Error
	if err != nil {
		return 0, 0, err
	}
	err = r.db.WithContext(ctx).Model(&AttendanceRecord{}).
		Where("school_id = ? AND class_division_id = ? AND date = ? AND status IN ?",
			schoolID, classDivisionID, date, []string{StatusPresent, StatusLate, StatusHalfDay}).
		Count(&present).Error
	return present, total, err
}
