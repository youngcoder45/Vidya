package dashboard

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/schoolos/backend/internal/deps"
	"github.com/schoolos/backend/internal/pkg/httpx"
	"github.com/schoolos/backend/internal/pkg/tenant"
	"github.com/schoolos/backend/internal/server/middleware"
)

// Register wires the dashboard route.
func Register(rg *gin.RouterGroup, d *deps.Deps) {
	h := &handler{db: d.DB}
	rg.GET("/dashboard/summary", middleware.RequirePermission("dashboard.read"), h.summary)
}

type handler struct{ db *gorm.DB }

// Summary is the admin dashboard payload.
type Summary struct {
	Students        int64     `json:"students"`
	Teachers        int64     `json:"teachers"`
	AttendanceToday *float64  `json:"attendance_today_percent"`
	PendingFees     *FeeTotals `json:"pending_fees"`
	UpcomingExams   []any     `json:"upcoming_exams"` // empty until exams module lands
	RecentAnnouncements []RecentAnnouncement `json:"recent_announcements"`
	Revenue         *Revenue  `json:"revenue"`
}

// FeeTotals aggregates outstanding dues.
type FeeTotals struct {
	Count     int64 `json:"count"`
	AmountINR int64 `json:"amount_inr"`
}

// RecentAnnouncement is a lightweight announcement row.
type RecentAnnouncement struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	PublishedAt *time.Time `json:"published_at"`
}

// Revenue is the collection summary for the current month.
type Revenue struct {
	CollectedINR  int64   `json:"collected_inr"`
	CollectionRate float64 `json:"collection_rate"`
}

func (h *handler) summary(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	s, err := h.build(c.Request.Context(), tc.SchoolID)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, s)
}

func (h *handler) build(ctx context.Context, schoolID uuid.UUID) (*Summary, error) {
	s := &Summary{UpcomingExams: []any{}}

	// Student count (active records).
	h.db.WithContext(ctx).Model(&struct{}{}).Table("students").
		Where("school_id = ? AND status = ?", schoolID, "active").Count(&s.Students)

	// Teacher count.
	h.db.WithContext(ctx).Model(&struct{}{}).Table("teachers").
		Where("school_id = ? AND status = ?", schoolID, "active").Count(&s.Teachers)

	// Attendance today.
	today := time.Now().UTC().Truncate(24 * time.Hour)
	var present, total int64
	h.db.WithContext(ctx).Model(&struct{}{}).Table("attendance_records").
		Where("school_id = ? AND date = ?", schoolID, today).Count(&total)
	h.db.WithContext(ctx).Model(&struct{}{}).Table("attendance_records").
		Where("school_id = ? AND date = ? AND status IN ?", schoolID, today,
			[]string{"present", "late", "half_day"}).Count(&present)
	if total > 0 {
		pct := float64(present) / float64(total) * 100
		s.AttendanceToday = &pct
	}

	// Pending fees.
	var dueCount int64
	var dueAmount struct{ Total int64 }
	h.db.WithContext(ctx).Model(&struct{}{}).Table("fee_ledgers").
		Where("school_id = ? AND status IN ?", schoolID, []string{"due", "partial"}).
		Count(&dueCount)
	h.db.WithContext(ctx).Model(&struct{}{}).Table("fee_ledgers").
		Select("COALESCE(SUM(amount_inr - paid_amount_inr - concession_inr), 0) AS total").
		Where("school_id = ? AND status IN ?", schoolID, []string{"due", "partial"}).
		Scan(&dueAmount)
	s.PendingFees = &FeeTotals{Count: dueCount, AmountINR: dueAmount.Total}

	// Recent announcements.
	h.db.WithContext(ctx).Model(&struct{}{}).Table("announcements").
		Select("id, title, publish_at").
		Where("school_id = ? AND status = ?", schoolID, "published").
		Order("publish_at DESC").Limit(5).
		Scan(&s.RecentAnnouncements)

	// Revenue this month.
	monthStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	var collected struct{ Total int64 }
	h.db.WithContext(ctx).Model(&struct{}{}).Table("payments").
		Select("COALESCE(SUM(amount_inr), 0) AS total").
		Where("school_id = ? AND paid_at >= ?", schoolID, monthStart).
		Scan(&collected)
	rate := 0.0
	if dueAmount.Total+collected.Total > 0 {
		rate = float64(collected.Total) / float64(dueAmount.Total+collected.Total) * 100
	}
	s.Revenue = &Revenue{CollectedINR: collected.Total, CollectionRate: rate}

	return s, nil
}
