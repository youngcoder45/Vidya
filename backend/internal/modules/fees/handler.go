package fees

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/vidya/backend/internal/deps"
	"github.com/vidya/backend/internal/pkg/ctxuser"
	"github.com/vidya/backend/internal/pkg/httpx"
	"github.com/vidya/backend/internal/pkg/tenant"
	"github.com/vidya/backend/internal/server/middleware"
)

// Register wires fees module routes.
func Register(rg *gin.RouterGroup, d *deps.Deps) {
	repo := NewRepository(d.DB)
	gateway := NewRazorpayGateway(d.Cfg.RazorpayKeyID, d.Cfg.RazorpayKeySecret)
	svc := NewService(repo, gateway, d.Cfg.RazorpayWebhookSecret, d.Audit, d.Bus, d.Log)
	h := &handler{svc: svc, repo: repo}

	grp := rg.Group("/fees")
	grp.GET("/heads", middleware.RequirePermission("fees.read"), h.listHeads)
	grp.POST("/heads", middleware.RequirePermission("fees.manage"), h.createHead)
	grp.GET("/structures", middleware.RequirePermission("fees.read"), h.listStructures)
	grp.POST("/structures", middleware.RequirePermission("fees.manage"), h.createStructure)
	grp.POST("/structures/generate", middleware.RequirePermission("fees.manage"), h.generateLedgers)
	grp.GET("/students/:studentID/ledger", middleware.RequirePermission("fees.read"), h.studentLedger)
	grp.GET("/dues", middleware.RequirePermission("fees.read"), h.dues)
	grp.POST("/orders", middleware.RequirePermission("fees.payment.create"), h.createOrder)
	grp.POST("/payments/offline", middleware.RequirePermission("fees.payment.create"), h.offlinePayment)
	grp.GET("/students/:studentID/payments", middleware.RequirePermission("fees.read"), h.studentPayments)
}

// RegisterWebhook wires the public Razorpay webhook endpoint (signed, no JWT).
func RegisterWebhook(rg *gin.RouterGroup, d *deps.Deps) {
	repo := NewRepository(d.DB)
	gateway := NewRazorpayGateway(d.Cfg.RazorpayKeyID, d.Cfg.RazorpayKeySecret)
	svc := NewService(repo, gateway, d.Cfg.RazorpayWebhookSecret, d.Audit, d.Bus, d.Log)
	h := &handler{svc: svc, repo: repo}
	rg.POST("/webhooks/razorpay", h.razorpayWebhook)
}

type handler struct {
	svc  *Service
	repo Repository
}

func (h *handler) createHead(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	var req struct {
		Name       string `json:"name" binding:"required"`
		Code       string `json:"code" binding:"required"`
		Category   string `json:"category" binding:"required,oneof=tuition transport hostel lab exam misc"`
		IsRecurring bool  `json:"is_recurring"`
		Frequency  string `json:"frequency" binding:"oneof=monthly quarterly term yearly"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	head := &FeeHead{
		ID: uuid.New(), SchoolID: tc.SchoolID, Name: req.Name, Code: req.Code,
		Category: req.Category, IsRecurring: req.IsRecurring, Frequency: req.Frequency,
	}
	if err := h.repo.CreateFeeHead(c.Request.Context(), head); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusCreated, head)
}

func (h *handler) listHeads(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	heads, err := h.repo.ListFeeHeads(c.Request.Context(), tc.SchoolID)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, heads)
}

func (h *handler) createStructure(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	var req struct {
		SessionID       uuid.UUID  `json:"session_id" binding:"required"`
		FeeHeadID       uuid.UUID  `json:"fee_head_id" binding:"required"`
		ClassDivisionID *uuid.UUID `json:"class_division_id"`
		AmountINR       int64      `json:"amount_inr" binding:"required,min=1"`
		DueDay          int        `json:"due_day"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	st := &FeeStructure{
		ID: uuid.New(), SchoolID: tc.SchoolID, SessionID: req.SessionID,
		FeeHeadID: req.FeeHeadID, ClassDivisionID: req.ClassDivisionID,
		AmountINR: req.AmountINR, DueDay: req.DueDay, IsActive: true,
	}
	if err := h.repo.CreateFeeStructure(c.Request.Context(), st); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusCreated, st)
}

func (h *handler) listStructures(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	sessionID, err := uuid.Parse(c.DefaultQuery("session_id", ""))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails("session_id query param required"))
		return
	}
	list, err := h.repo.ListFeeStructures(c.Request.Context(), tc.SchoolID, sessionID)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, list)
}

func (h *handler) generateLedgers(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	var req struct {
		SessionID uuid.UUID `json:"session_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	created, err := h.repo.GenerateLedgers(c.Request.Context(), tc.SchoolID, req.SessionID)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, gin.H{"ledgers_created": created})
}

func (h *handler) studentLedger(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	studentID, err := uuid.Parse(c.Param("studentID"))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation)
		return
	}
	ledgers, err := h.repo.ListLedgersByStudent(c.Request.Context(), tc.SchoolID, studentID)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, ledgers)
}

func (h *handler) dues(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	sessionID, err := uuid.Parse(c.DefaultQuery("session_id", ""))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails("session_id query param required"))
		return
	}
	classDivisionID, _ := uuid.Parse(c.DefaultQuery("class_division_id", ""))
	list, err := h.repo.ListDues(c.Request.Context(), tc.SchoolID, sessionID, classDivisionID, c.Query("status"))
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, list)
}

func (h *handler) createOrder(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	createdBy := ctxuser.MustUserID(c.Request.Context())
	var req struct {
		StudentID uuid.UUID `json:"student_id" binding:"required"`
		AmountINR int64     `json:"amount_inr" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	order, gw, err := h.svc.CreateOrder(c.Request.Context(), tc.SchoolID, CreateOrderParams{
		StudentID: req.StudentID, AmountINR: req.AmountINR, CreatedBy: createdBy,
	})
	if err != nil {
		if errors.Is(err, ErrOutstandingMismatch) {
			httpx.WriteError(c, httpx.NewError(http.StatusUnprocessableEntity, "OUTSTANDING_MISMATCH", "amount exceeds outstanding dues"))
			return
		}
		if errors.Is(err, ErrStudentNotFound) {
			httpx.WriteError(c, httpx.ErrNotFound)
			return
		}
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusCreated, gin.H{
		"order_id": order.ID, "amount_inr": order.AmountINR, "currency": order.Currency,
		"razorpay_order_id": gw.GatewayOrderID, "key_id": gw.KeyID,
	})
}

func (h *handler) offlinePayment(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	recordedBy := ctxuser.MustUserID(c.Request.Context())
	var req struct {
		StudentID uuid.UUID `json:"student_id" binding:"required"`
		Mode      string    `json:"mode" binding:"required,oneof=cash upi card netbanking cheque"`
		AmountINR int64     `json:"amount_inr" binding:"required,min=1"`
		Notes     string    `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	payment, err := h.svc.CaptureOfflinePayment(c.Request.Context(), tc.SchoolID, OfflinePaymentInput{
		StudentID: req.StudentID, Mode: req.Mode, AmountINR: req.AmountINR,
		Notes: req.Notes, RecordedBy: recordedBy,
	})
	if err != nil {
		if errors.Is(err, ErrOutstandingMismatch) {
			httpx.WriteError(c, httpx.NewError(http.StatusUnprocessableEntity, "OUTSTANDING_MISMATCH", "amount exceeds outstanding dues"))
			return
		}
		if errors.Is(err, ErrStudentNotFound) {
			httpx.WriteError(c, httpx.ErrNotFound)
			return
		}
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusCreated, payment)
}

func (h *handler) studentPayments(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	studentID, err := uuid.Parse(c.Param("studentID"))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation)
		return
	}
	payments, err := h.repo.ListPaymentsByStudent(c.Request.Context(), tc.SchoolID, studentID)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, payments)
}

func (h *handler) razorpayWebhook(c *gin.Context) {
	// Bound the body: webhooks are public and must not be able to exhaust memory.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 256*1024)
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation)
		return
	}
	if !VerifyWebhookSignature(raw, c.GetHeader("X-Razorpay-Signature"), h.svc.webhookSecret) {
		httpx.WriteError(c, httpx.NewError(http.StatusUnauthorized, "INVALID_SIGNATURE", "webhook signature mismatch"))
		return
	}
	var ev RazorpayWebhookEvent
	if err := json.Unmarshal(raw, &ev); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails("malformed webhook payload"))
		return
	}
	if err := h.svc.HandleWebhook(c.Request.Context(), ev); err != nil {
		httpx.WriteError(c, err)
		return
	}
	// Fast-ack: Razorpay retries on non-2xx.
	httpx.WriteJSON(c, http.StatusOK, gin.H{"received": true})
}
