package fees

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/schoolos/backend/internal/pkg/httpx"
)

// ErrNotFound maps record-not-found to a 404.
var ErrNotFound = httpx.ErrNotFound

// Repository is the persistence port for the fees module.
type Repository interface {
	CreateFeeHead(ctx context.Context, h *FeeHead) error
	ListFeeHeads(ctx context.Context, schoolID uuid.UUID) ([]FeeHead, error)

	CreateFeeStructure(ctx context.Context, s *FeeStructure) error
	ListFeeStructures(ctx context.Context, schoolID, sessionID uuid.UUID) ([]FeeStructure, error)

	StudentExists(ctx context.Context, schoolID, studentID uuid.UUID) (bool, error)
	GuardianUserIDs(ctx context.Context, schoolID, studentID uuid.UUID) ([]uuid.UUID, error)
	GenerateLedgers(ctx context.Context, schoolID, sessionID uuid.UUID) (int, error)
	ListLedgersByStudent(ctx context.Context, schoolID, studentID uuid.UUID) ([]FeeLedger, error)
	ListDues(ctx context.Context, schoolID, sessionID, classDivisionID uuid.UUID, status string) ([]FeeLedger, error)

	CreateOrder(ctx context.Context, o *FeePaymentOrder) error
	GetOrderByID(ctx context.Context, schoolID, id uuid.UUID) (*FeePaymentOrder, error)
	GetOrderByGatewayID(ctx context.Context, gatewayOrderID string) (*FeePaymentOrder, error)
	UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string) error
	SetOrderGatewayID(ctx context.Context, id uuid.UUID, gatewayOrderID string) error

	// CapturePayment atomically records a payment + allocations + receipt and
	// credits the ledger. Append-only money path.
	CapturePayment(ctx context.Context, p *Payment, allocations []PaymentAllocation, ledgerUpdates []FeeLedger) error
	ListPaymentsByStudent(ctx context.Context, schoolID, studentID uuid.UUID) ([]Payment, error)
	CountReceiptsForYear(ctx context.Context, schoolID uuid.UUID, year int) (int64, error)
}

// GormRepo implements Repository.
type GormRepo struct{ db *gorm.DB }

// NewRepository creates the fees repository.
func NewRepository(db *gorm.DB) *GormRepo { return &GormRepo{db: db} }

func (r *GormRepo) CreateFeeHead(ctx context.Context, h *FeeHead) error { return r.db.WithContext(ctx).Create(h).Error }
func (r *GormRepo) ListFeeHeads(ctx context.Context, schoolID uuid.UUID) ([]FeeHead, error) {
	var out []FeeHead
	err := r.db.WithContext(ctx).Where("school_id = ?", schoolID).Order("name ASC").Find(&out).Error
	return out, err
}

func (r *GormRepo) CreateFeeStructure(ctx context.Context, s *FeeStructure) error {
	return r.db.WithContext(ctx).Create(s).Error
}
func (r *GormRepo) ListFeeStructures(ctx context.Context, schoolID, sessionID uuid.UUID) ([]FeeStructure, error) {
	var out []FeeStructure
	err := r.db.WithContext(ctx).
		Where("school_id = ? AND session_id = ?", schoolID, sessionID).
		Order("created_at ASC").Find(&out).Error
	return out, err
}

// GenerateLedgers creates one ledger row per active student per applicable
// structure for the session. Idempotent-ish: skips (student, head) pairs that
// already have a ledger row.
func (r *GormRepo) StudentExists(ctx context.Context, schoolID, studentID uuid.UUID) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Table("students").
		Where("school_id = ? AND id = ?", schoolID, studentID).Count(&n).Error
	return n > 0, err
}

// GuardianUserIDs returns registered parent user accounts for a student.
func (r *GormRepo) GuardianUserIDs(ctx context.Context, schoolID, studentID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Table("guardians g").
		Select("DISTINCT g.user_id").
		Joins("JOIN student_guardians sg ON sg.guardian_id = g.id").
		Where("g.school_id = ? AND sg.student_id = ? AND g.user_id IS NOT NULL", schoolID, studentID).
		Scan(&ids).Error
	return ids, err
}

func (r *GormRepo) GenerateLedgers(ctx context.Context, schoolID, sessionID uuid.UUID) (int, error) {
	var structures []FeeStructure
	if err := r.db.WithContext(ctx).
		Where("school_id = ? AND session_id = ? AND is_active = ?", schoolID, sessionID, true).
		Find(&structures).Error; err != nil {
		return 0, err
	}
	type enrollment struct {
		StudentID       uuid.UUID
		ClassDivisionID uuid.UUID
	}
	var enrollments []enrollment
	if err := r.db.WithContext(ctx).Table("student_enrollments").
		Select("student_id, class_division_id").
		Where("school_id = ? AND session_id = ? AND status = ?", schoolID, sessionID, "active").
		Scan(&enrollments).Error; err != nil {
		return 0, err
	}

	var existing []FeeLedger
	if err := r.db.WithContext(ctx).
		Where("school_id = ? AND session_id = ?", schoolID, sessionID).
		Find(&existing).Error; err != nil {
		return 0, err
	}
	seen := map[string]bool{}
	for _, l := range existing {
		seen[l.StudentID.String()+":"+l.FeeHeadID.String()] = true
	}

	created := 0
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, en := range enrollments {
			for _, st := range structures {
				if st.ClassDivisionID != nil && *st.ClassDivisionID != en.ClassDivisionID {
					continue
				}
				key := en.StudentID.String() + ":" + st.FeeHeadID.String()
				if seen[key] {
					continue
				}
				ledger := &FeeLedger{
					ID: uuid.New(), SchoolID: schoolID, SessionID: sessionID,
					StudentID: en.StudentID, FeeHeadID: st.FeeHeadID,
					AmountINR: st.AmountINR, DueDate: dueDateFor(st),
					Status: LedgerDue,
				}
				if err := tx.Create(ledger).Error; err != nil {
					return err
				}
				// Mark in-run so an overlapping structure (whole-school + class)
				// cannot create a second ledger for the same student/head.
				seen[key] = true
				created++
			}
		}
		return nil
	})
	return created, err
}

// dueDateFor derives a due date from the structure's due_day, anchored to its
// applicability start month when set.
func dueDateFor(st FeeStructure) time.Time {
	// Anchor to the structure's applicability start when set, otherwise today.
	base := time.Now()
	if st.ApplicableFrom != nil {
		base = *st.ApplicableFrom
	}
	day := st.DueDay
	if day <= 0 {
		day = 10
	}
	lastDay := time.Date(base.Year(), base.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
	if day > lastDay {
		day = lastDay
	}
	return time.Date(base.Year(), base.Month(), day, 0, 0, 0, 0, time.UTC)
}

func (r *GormRepo) ListLedgersByStudent(ctx context.Context, schoolID, studentID uuid.UUID) ([]FeeLedger, error) {
	var out []FeeLedger
	err := r.db.WithContext(ctx).
		Where("school_id = ? AND student_id = ?", schoolID, studentID).
		Order("due_date ASC").Find(&out).Error
	return out, err
}

func (r *GormRepo) ListDues(ctx context.Context, schoolID, sessionID, classDivisionID uuid.UUID, status string) ([]FeeLedger, error) {
	q := r.db.WithContext(ctx).
		Joins("JOIN student_enrollments en ON en.student_id = fee_ledgers.student_id AND en.school_id = fee_ledgers.school_id").
		Where("fee_ledgers.school_id = ? AND en.session_id = ?", schoolID, sessionID)
	if classDivisionID != uuid.Nil {
		q = q.Where("en.class_division_id = ?", classDivisionID)
	}
	if status != "" {
		q = q.Where("fee_ledgers.status = ?", status)
	} else {
		q = q.Where("fee_ledgers.status IN ?", []string{LedgerDue, LedgerPartial})
	}
	var out []FeeLedger
	err := q.Order("fee_ledgers.due_date ASC").Find(&out).Error
	return out, err
}

func (r *GormRepo) CreateOrder(ctx context.Context, o *FeePaymentOrder) error { return r.db.WithContext(ctx).Create(o).Error }

func (r *GormRepo) GetOrderByID(ctx context.Context, schoolID, id uuid.UUID) (*FeePaymentOrder, error) {
	var o FeePaymentOrder
	if err := r.db.WithContext(ctx).First(&o, "school_id = ? AND id = ?", schoolID, id).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *GormRepo) GetOrderByGatewayID(ctx context.Context, gatewayOrderID string) (*FeePaymentOrder, error) {
	var o FeePaymentOrder
	if err := r.db.WithContext(ctx).First(&o, "gateway_order_id = ?", gatewayOrderID).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *GormRepo) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&FeePaymentOrder{}).Where("id = ?", id).Update("status", status).Error
}

func (r *GormRepo) SetOrderGatewayID(ctx context.Context, id uuid.UUID, gatewayOrderID string) error {
	return r.db.WithContext(ctx).Model(&FeePaymentOrder{}).Where("id = ?", id).
		Update("gateway_order_id", gatewayOrderID).Error
}

// CapturePayment records the payment + allocations + receipt and credits the
// fee ledger in one transaction.
func (r *GormRepo) CapturePayment(ctx context.Context, p *Payment, allocations []PaymentAllocation, ledgerUpdates []FeeLedger) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(p).Error; err != nil {
			return err
		}
		for i := range allocations {
			if err := tx.Create(&allocations[i]).Error; err != nil {
				return err
			}
		}
		for _, lu := range ledgerUpdates {
			if err := tx.Model(&FeeLedger{}).Where("id = ?", lu.ID).
				Updates(map[string]any{
					"paid_amount_inr": lu.PaidAmountINR,
					"status":          lu.Status,
				}).Error; err != nil {
				return err
			}
		}
		return tx.Create(&Receipt{
			ID: uuid.New(), SchoolID: p.SchoolID, PaymentID: p.ID,
			ReceiptNo: p.ReceiptNo, IssuedAt: time.Now(),
		}).Error
	})
}

func (r *GormRepo) ListPaymentsByStudent(ctx context.Context, schoolID, studentID uuid.UUID) ([]Payment, error) {
	var out []Payment
	err := r.db.WithContext(ctx).
		Where("school_id = ? AND student_id = ?", schoolID, studentID).
		Order("paid_at DESC").Find(&out).Error
	return out, err
}

func (r *GormRepo) CountReceiptsForYear(ctx context.Context, schoolID uuid.UUID, year int) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&Receipt{}).
		Where("school_id = ? AND EXTRACT(YEAR FROM issued_at) = ?", schoolID, year).
		Count(&n).Error
	return n, err
}
