package fees

import (
	"time"

	"github.com/google/uuid"
)

// Fee head categories and frequencies.
const (
	CategoryTuition   = "tuition"
	CategoryTransport = "transport"
	CategoryHostel    = "hostel"
	CategoryLab       = "lab"
	CategoryExam      = "exam"
	CategoryMisc      = "misc"

	FrequencyMonthly   = "monthly"
	FrequencyQuarterly = "quarterly"
	FrequencyTerm      = "term"
	FrequencyYearly    = "yearly"
)

// FeeHead is a fee category ("Tuition Fee").
type FeeHead struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID    uuid.UUID `gorm:"type:uuid;not null;index" json:"school_id"`
	Name        string    `gorm:"size:80;not null" json:"name"`
	Code        string    `gorm:"size:40;not null" json:"code"`
	Category    string    `gorm:"size:20;not null" json:"category"`
	IsRecurring bool      `gorm:"default:true" json:"is_recurring"`
	Frequency   string    `gorm:"size:20;default:monthly" json:"frequency"`
	TaxRate     float64   `gorm:"default:0" json:"tax_rate"`
	CreatedAt   time.Time `json:"created_at"`
}

// FeeStructure sets an amount for a head, optionally per class, for a session.
type FeeStructure struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"school_id"`
	SessionID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"session_id"`
	FeeHeadID       uuid.UUID  `gorm:"type:uuid;not null" json:"fee_head_id"`
	ClassDivisionID *uuid.UUID `gorm:"type:uuid" json:"class_division_id,omitempty"` // nil = whole school
	AmountINR       int64      `gorm:"not null" json:"amount_inr"`                   // integer paise? no: INR whole units stored as int64 rupees
	DueDay          int        `gorm:"default:10" json:"due_day"`
	ApplicableFrom  *time.Time `json:"applicable_from,omitempty"`
	ApplicableTo    *time.Time `json:"applicable_to,omitempty"`
	IsActive        bool       `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time  `json:"created_at"`
}

// Ledger status values.
const (
	LedgerDue    = "due"
	LedgerPartial = "partial"
	LedgerPaid   = "paid"
	LedgerWaived = "waived"
)

// FeeLedger is the per-student obligation for one fee head. Append-only:
// paid amounts change only through payments/allocations, never direct edits.
type FeeLedger struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID     uuid.UUID  `gorm:"type:uuid;not null;index:idx_fl_student" json:"school_id"`
	SessionID    uuid.UUID  `gorm:"type:uuid;not null;index:idx_fl_session_head" json:"session_id"`
	StudentID    uuid.UUID  `gorm:"type:uuid;not null;index:idx_fl_student" json:"student_id"`
	FeeHeadID    uuid.UUID  `gorm:"type:uuid;not null;index:idx_fl_session_head" json:"fee_head_id"`
	AmountINR    int64      `gorm:"not null" json:"amount_inr"`
	DueDate      time.Time  `gorm:"index:idx_fl_status_due" json:"due_date"`
	PaidAmountINR int64     `gorm:"default:0" json:"paid_amount_inr"`
	ConcessionINR int64     `gorm:"default:0" json:"concession_inr"`
	Status       string     `gorm:"size:20;not null;default:due;index:idx_fl_status_due" json:"status"`
	WaivedBy     *uuid.UUID `gorm:"type:uuid" json:"waived_by,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// FeeInstallment breaks a ledger item into a payment schedule.
type FeeInstallment struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"school_id"`
	FeeLedgerID uuid.UUID  `gorm:"type:uuid;not null;index" json:"fee_ledger_id"`
	InstallmentNo int      `json:"installment_no"`
	AmountINR   int64      `gorm:"not null" json:"amount_inr"`
	DueDate     time.Time  `json:"due_date"`
	PaidAt      *time.Time `json:"paid_at,omitempty"`
}

// Payment order status values (idempotency state machine).
const (
	OrderPending   = "pending"
	OrderProcessing = "processing"
	OrderCaptured  = "captured"
	OrderFailed    = "failed"
	OrderRefunded  = "refunded"
)

// FeePaymentOrder tracks an online payment attempt (Razorpay order).
type FeePaymentOrder struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_ord_idem,priority:1" json:"school_id"`
	StudentID      uuid.UUID `gorm:"type:uuid;not null;index" json:"student_id"`
	AmountINR      int64     `gorm:"not null" json:"amount_inr"`
	Currency       string    `gorm:"size:8;default:INR" json:"currency"`
	Gateway        string    `gorm:"size:20;default:razorpay" json:"gateway"`
	GatewayOrderID string    `gorm:"size:80" json:"gateway_order_id"`
	Status         string    `gorm:"size:20;not null;default:pending" json:"status"`
	IdempotencyKey string    `gorm:"size:80;uniqueIndex:idx_ord_idem,priority:2" json:"-"`
	CreatedBy      uuid.UUID `gorm:"type:uuid" json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Payment modes.
const (
	ModeCash     = "cash"
	ModeUPI      = "upi"
	ModeCard     = "card"
	ModeNetBanking = "netbanking"
	ModeCheque   = "cheque"
	ModeOnline   = "online"
)

// Payment is an append-only money-in record.
type Payment struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID    uuid.UUID  `gorm:"type:uuid;not null;index:idx_pay_school_time;uniqueIndex:idx_pay_receipt,priority:1" json:"school_id"`
	StudentID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"student_id"`
	OrderID     *uuid.UUID `gorm:"type:uuid;index" json:"order_id,omitempty"`
	ReceiptNo   string     `gorm:"size:40;uniqueIndex:idx_pay_receipt,priority:2" json:"receipt_no"`
	AmountINR   int64      `gorm:"not null" json:"amount_inr"`
	Mode        string     `gorm:"size:20;not null" json:"mode"`
	GatewayRef  string     `gorm:"size:120" json:"gateway_ref,omitempty"`
	PaidAt      time.Time  `gorm:"index:idx_pay_school_time" json:"paid_at"`
	RecordedBy  uuid.UUID  `gorm:"type:uuid;not null" json:"recorded_by"`
	Notes       string     `gorm:"size:300" json:"notes,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// PaymentAllocation splits a payment across ledger items.
type PaymentAllocation struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID    uuid.UUID `gorm:"type:uuid;not null;index" json:"school_id"`
	PaymentID   uuid.UUID `gorm:"type:uuid;not null;index" json:"payment_id"`
	FeeLedgerID uuid.UUID `gorm:"type:uuid;not null;index" json:"fee_ledger_id"`
	AmountINR   int64     `gorm:"not null" json:"amount_inr"`
}

// Receipt is the issued receipt for a payment (PDF stored in S3).
type Receipt struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID  uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_receipt_no,priority:1" json:"school_id"`
	PaymentID uuid.UUID `gorm:"type:uuid;not null;index" json:"payment_id"`
	ReceiptNo string    `gorm:"size:40;uniqueIndex:idx_receipt_no,priority:2" json:"receipt_no"`
	PDFURL    string    `gorm:"size:500" json:"pdf_url,omitempty"`
	IssuedAt  time.Time `json:"issued_at"`
}

// RefundStatus values.
const (
	RefundPending   = "pending"
	RefundProcessed = "processed"
)

// Refund reverses a payment (append-only reversing entries).
type Refund struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SchoolID    uuid.UUID `gorm:"type:uuid;not null;index" json:"school_id"`
	PaymentID   uuid.UUID `gorm:"type:uuid;not null;index" json:"payment_id"`
	AmountINR   int64     `gorm:"not null" json:"amount_inr"`
	Reason      string    `gorm:"size:300" json:"reason"`
	Status      string    `gorm:"size:20;not null;default:pending" json:"status"`
	ProcessedBy uuid.UUID `gorm:"type:uuid" json:"processed_by"`
	CreatedAt   time.Time `json:"created_at"`
}
