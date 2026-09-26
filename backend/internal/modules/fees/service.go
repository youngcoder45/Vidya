package fees

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/schoolos/backend/internal/events"
	"github.com/schoolos/backend/internal/pkg/audit"
)

// ErrOutstandingMismatch is returned when allocations don't match the
// available outstanding amount.
var ErrOutstandingMismatch = errors.New("fees: payment amount exceeds outstanding dues")

// ErrStudentNotFound is returned when the student is not in the tenant.
var ErrStudentNotFound = errors.New("fees: student not found")

// PaymentGateway abstracts an online payment provider (Razorpay today).
type PaymentGateway interface {
	CreateOrder(ctx context.Context, in CreateOrderInput) (*GatewayOrder, error)
}

// CreateOrderInput is a gateway order request.
type CreateOrderInput struct {
	AmountINR int64
	Currency  string
	Receipt   string
}

// GatewayOrder is the gateway's response.
type GatewayOrder struct {
	GatewayOrderID string
	KeyID          string
}

// RazorpayGateway implements PaymentGateway against the Razorpay Orders API.
// When credentials are absent (dev), it returns a synthetic order id so the
// whole flow is exercisable locally.
type RazorpayGateway struct {
	keyID     string
	keySecret string
	http      *http.Client
}

// NewRazorpayGateway creates the gateway.
func NewRazorpayGateway(keyID, keySecret string) *RazorpayGateway {
	return &RazorpayGateway{keyID: keyID, keySecret: keySecret, http: &http.Client{Timeout: 10 * time.Second}}
}

func (g *RazorpayGateway) CreateOrder(ctx context.Context, in CreateOrderInput) (*GatewayOrder, error) {
	if g.keyID == "" || g.keySecret == "" {
		// Dev mode: synthetic order so flows can be tested end-to-end.
		return &GatewayOrder{GatewayOrderID: "rzp_test_" + uuid.NewString()[:16], KeyID: "rzp_test_dev_key"}, nil
	}
	body, _ := json.Marshal(map[string]any{
		"amount":   in.AmountINR * 100, // paise
		"currency": in.Currency,
		"receipt":  in.Receipt,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.razorpay.com/v1/orders", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(g.keyID, g.keySecret)
	req.Header.Set("Content-Type", "application/json")
	resp, err := g.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("razorpay: order failed (%d): %s", resp.StatusCode, string(raw))
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &GatewayOrder{GatewayOrderID: out.ID, KeyID: g.keyID}, nil
}

// Service implements fees use cases.
type Service struct {
	repo          Repository
	gateway       PaymentGateway
	webhookSecret string
	audit         *audit.Writer
	bus           *events.Bus
	log           *slog.Logger
}

// NewService wires the fees service.
func NewService(repo Repository, gateway PaymentGateway, webhookSecret string, auditWriter *audit.Writer, bus *events.Bus, log *slog.Logger) *Service {
	return &Service{repo: repo, gateway: gateway, webhookSecret: webhookSecret, audit: auditWriter, bus: bus, log: log}
}

// notifyPaid emits fees.paid to each registered guardian so the notification
// subscriber can fan out. Best-effort: a failure never fails the payment.
func (s *Service) notifyPaid(ctx context.Context, schoolID, studentID uuid.UUID, p *Payment) {
	if s.bus == nil {
		return
	}
	ids, err := s.repo.GuardianUserIDs(ctx, schoolID, studentID)
	if err != nil {
		s.log.Warn("guardian lookup failed for fee notification", "err", err)
		return
	}
	for _, uid := range ids {
		s.bus.Publish(ctx, events.NewEvent(EvFeePaid, schoolID, uid, map[string]any{
			"payment_id": p.ID.String(), "amount_inr": p.AmountINR,
		}))
	}
}

// OfflinePaymentInput is a cash/UPI/cheque payment captured by office staff.
type OfflinePaymentInput struct {
	StudentID uuid.UUID
	Mode      string
	AmountINR int64
	Notes     string
	RecordedBy uuid.UUID
}

// CaptureOfflinePayment validates the amount against outstanding dues and
// records the payment (append-only) with allocations.
func (s *Service) CaptureOfflinePayment(ctx context.Context, schoolID uuid.UUID, in OfflinePaymentInput) (*Payment, error) {
	if ok, err := s.repo.StudentExists(ctx, schoolID, in.StudentID); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrStudentNotFound
	}
	ledgers, err := s.repo.ListLedgersByStudent(ctx, schoolID, in.StudentID)
	if err != nil {
		return nil, err
	}
	allocations, updates, err := computeAllocations(ledgers, in.AmountINR)
	if err != nil {
		return nil, err
	}
	p := &Payment{
		ID: uuid.New(), SchoolID: schoolID, StudentID: in.StudentID,
		ReceiptNo: s.nextReceiptNo(ctx, schoolID), AmountINR: in.AmountINR,
		Mode: in.Mode, PaidAt: time.Now(), RecordedBy: in.RecordedBy, Notes: in.Notes,
	}
	if err := s.captureWithReceiptRetry(ctx, schoolID, p, allocations, updates); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, audit.Entry{
		SchoolID: &schoolID, UserID: &in.RecordedBy, Action: "fees.offline_payment",
		EntityType: "payment", EntityID: p.ID.String(),
	})
	s.notifyPaid(ctx, schoolID, in.StudentID, p)
	return p, nil
}

// captureWithReceiptRetry records the payment, regenerating the receipt number
// if a concurrent write claimed it first (count()+1 is not atomic).
func (s *Service) captureWithReceiptRetry(ctx context.Context, schoolID uuid.UUID, p *Payment, allocations []PaymentAllocation, updates []FeeLedger) error {
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		err = s.repo.CapturePayment(ctx, p, allocations, updates)
		if err == nil || !strings.Contains(err.Error(), "duplicate key") {
			return err
		}
		p.ReceiptNo = s.nextReceiptNo(ctx, schoolID)
	}
	return err
}

// CreateOrderParams is a request to start an online payment.
type CreateOrderParams struct {
	StudentID uuid.UUID
	AmountINR int64
	CreatedBy uuid.UUID
}

// CreateOrder validates the amount and creates a gateway order.
func (s *Service) CreateOrder(ctx context.Context, schoolID uuid.UUID, in CreateOrderParams) (*FeePaymentOrder, *GatewayOrder, error) {
	if ok, err := s.repo.StudentExists(ctx, schoolID, in.StudentID); err != nil {
		return nil, nil, err
	} else if !ok {
		return nil, nil, ErrStudentNotFound
	}
	ledgers, err := s.repo.ListLedgersByStudent(ctx, schoolID, in.StudentID)
	if err != nil {
		return nil, nil, err
	}
	totalOutstanding := int64(0)
	for _, l := range ledgers {
		if l.Status == LedgerDue || l.Status == LedgerPartial {
			totalOutstanding += l.AmountINR - l.PaidAmountINR - l.ConcessionINR
		}
	}
	if totalOutstanding < in.AmountINR {
		return nil, nil, ErrOutstandingMismatch
	}
	order := &FeePaymentOrder{
		ID: uuid.New(), SchoolID: schoolID, StudentID: in.StudentID,
		AmountINR: in.AmountINR, Currency: "INR", Gateway: "razorpay",
		Status: OrderPending, IdempotencyKey: uuid.NewString(), CreatedBy: in.CreatedBy,
	}
	if err := s.repo.CreateOrder(ctx, order); err != nil {
		return nil, nil, err
	}
	gwOrder, err := s.gateway.CreateOrder(ctx, CreateOrderInput{
		AmountINR: in.AmountINR, Currency: "INR", Receipt: order.ID.String(),
	})
	if err != nil {
		return nil, nil, err
	}
	order.GatewayOrderID = gwOrder.GatewayOrderID
	if err := s.repo.SetOrderGatewayID(ctx, order.ID, gwOrder.GatewayOrderID); err != nil {
		return nil, nil, err
	}
	return order, gwOrder, nil
}

// RazorpayWebhookEvent is the subset of the Razorpay event payload we consume.
type RazorpayWebhookEvent struct {
	Event string `json:"event"`
	Payload struct {
		Payment struct {
			Entity struct {
				ID       string `json:"id"`
				OrderID  string `json:"order_id"`
				Amount   int64  `json:"amount"` // paise
				Currency string `json:"currency"`
				Status   string `json:"status"`
			} `json:"entity"`
		} `json:"payment"`
	} `json:"payload"`
}

// VerifyWebhookSignature checks the X-Razorpay-Signature HMAC.
func VerifyWebhookSignature(rawBody []byte, signature, secret string) bool {
	if secret == "" {
		return true // dev mode: accept unsigned webhooks
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(rawBody)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// HandleWebhook processes payment.captured events idempotently. The tenant is
// derived from the matched order, never from the (unsigned) payload.
func (s *Service) HandleWebhook(ctx context.Context, ev RazorpayWebhookEvent) error {
	if ev.Event != "payment.captured" {
		return nil // other events (order.paid etc.) are no-ops here
	}
	p := ev.Payload.Payment.Entity
	order, err := s.repo.GetOrderByGatewayID(ctx, p.OrderID)
	if err != nil {
		// Ack unknown orders so the gateway stops retrying; alert via logs.
		s.log.Warn("webhook for unknown order", "gateway_order_id", p.OrderID)
		return nil
	}
	if order.Status == OrderCaptured {
		return nil // idempotent: already captured
	}
	if p.Status != "captured" {
		_ = s.repo.UpdateOrderStatus(ctx, order.ID, OrderFailed)
		return nil
	}
	amountINR := p.Amount / 100
	if amountINR != order.AmountINR {
		s.log.Error("webhook amount mismatch — ignoring",
			"order_id", order.ID, "expected_inr", order.AmountINR, "received_inr", amountINR)
		return nil
	}
	ledgers, err := s.repo.ListLedgersByStudent(ctx, order.SchoolID, order.StudentID)
	if err != nil {
		return err
	}
	allocations, updates, err := computeAllocations(ledgers, amountINR)
	if err != nil {
		// Over-allocated or inconsistent; ack and reconcile offline.
		s.log.Error("webhook allocation failed", "order_id", order.ID, "err", err)
		return nil
	}
	payment := &Payment{
		ID: uuid.New(), SchoolID: order.SchoolID, StudentID: order.StudentID,
		OrderID: &order.ID, ReceiptNo: s.nextReceiptNo(ctx, order.SchoolID),
		AmountINR: amountINR, Mode: ModeOnline, GatewayRef: p.ID,
		PaidAt: time.Now(), RecordedBy: order.CreatedBy,
	}
	if err := s.captureWithReceiptRetry(ctx, order.SchoolID, payment, allocations, updates); err != nil {
		return err
	}
	s.audit.Record(ctx, audit.Entry{
		SchoolID: &order.SchoolID, Action: "fees.online_payment",
		EntityType: "payment", EntityID: payment.ID.String(),
	})
	s.notifyPaid(ctx, order.SchoolID, order.StudentID, payment)
	return s.repo.UpdateOrderStatus(ctx, order.ID, OrderCaptured)
}

// computeAllocations splits amount across outstanding ledgers (oldest first)
// and computes the resulting ledger states.
func computeAllocations(ledgers []FeeLedger, amountINR int64) ([]PaymentAllocation, []FeeLedger, error) {
	outstanding := make([]FeeLedger, 0, len(ledgers))
	for _, l := range ledgers {
		if l.Status == LedgerDue || l.Status == LedgerPartial {
			outstanding = append(outstanding, l)
		}
	}
	sort.Slice(outstanding, func(i, j int) bool { return outstanding[i].DueDate.Before(outstanding[j].DueDate) })

	var remaining = amountINR
	var allocations []PaymentAllocation
	var updates []FeeLedger
	for _, l := range outstanding {
		if remaining <= 0 {
			break
		}
		due := l.AmountINR - l.PaidAmountINR - l.ConcessionINR
		if due <= 0 {
			continue
		}
		alloc := due
		if alloc > remaining {
			alloc = remaining
		}
		allocations = append(allocations, PaymentAllocation{
			ID: uuid.New(), SchoolID: l.SchoolID, FeeLedgerID: l.ID, AmountINR: alloc,
		})
		newPaid := l.PaidAmountINR + alloc
		status := LedgerPartial
		if newPaid >= l.AmountINR-l.ConcessionINR {
			status = LedgerPaid
		}
		updates = append(updates, FeeLedger{ID: l.ID, PaidAmountINR: newPaid, Status: status})
		remaining -= alloc
	}
	if remaining > 0 {
		return nil, nil, ErrOutstandingMismatch
	}
	return allocations, updates, nil
}

func (s *Service) nextReceiptNo(ctx context.Context, schoolID uuid.UUID) string {
	year := time.Now().Year()
	n, err := s.repo.CountReceiptsForYear(ctx, schoolID, year)
	if err != nil {
		n = 0
	}
	return fmt.Sprintf("RCP-%d-%05d", year, n+1)
}
