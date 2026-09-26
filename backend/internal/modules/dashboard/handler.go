package dashboard

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/vidya/backend/internal/deps"
	"github.com/vidya/backend/internal/pkg/httpx"
	"github.com/vidya/backend/internal/pkg/tenant"
	"github.com/vidya/backend/internal/server/middleware"
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

// schoolToday resolves "today" in the tenant timezone as a UTC-midnight time
// so comparisons against DATE columns stay correct across offsets.
func (h *handler) schoolToday(ctx context.Context, schoolID uuid.UUID) time.Time {
	loc := time.UTC
	var row struct{ Timezone string }
	if err := h.db.WithContext(ctx).Table("schools").Select("timezone").
		Where("id = ?", schoolID).Scan(&row).Error; err == nil && row.Timezone != "" {
		if l, e := time.LoadLocation(row.Timezone); e == nil {
			loc = l
		}
	}
	n := time.Now().In(loc)
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
}

func (h *handler) build(ctx context.Context, schoolID uuid.UUID) (*Summary, error) {
	s := &Summary{UpcomingExams: []any{}}

	// Student count (active records).
	if err := h.db.WithContext(ctx).Model(&struct{}{}).Table("students").
		Where("school_id = ? AND status = ?", schoolID, "active").Count(&s.Students).Error; err != nil {
		return nil, err
	}

	// Teacher count.
	if err := h.db.WithContext(ctx).Model(&struct{}{}).Table("teachers").
		Where("school_id = ? AND status = ?", schoolID, "active").Count(&s.Teachers).Error; err != nil {
		return nil, err
	}

	// Attendance today, in the school's local calendar (columns store UTC dates).
	today := h.schoolToday(ctx, schoolID)
	var present, total int64
	if err := h.db.WithContext(ctx).Model(&struct{}{}).Table("attendance_records").
		Where("school_id = ? AND date = ?", schoolID, today).Count(&total).Error; err != nil {
		return nil, err
	}
	if err := h.db.WithContext(ctx).Model(&struct{}{}).Table("attendance_records").
		Where("school_id = ? AND date = ? AND status IN ?", schoolID, today,
			[]string{"present", "late", "half_day"}).Count(&present).Error; err != nil {
		return nil, err
	}
	if total > 0 {
		pct := float64(present) / float64(total) * 100
		s.AttendanceToday = &pct
	}

	// Pending fees.
	var dueCount int64
	var dueAmount struct{ Total int64 }
	if err := h.db.WithContext(ctx).Model(&struct{}{}).Table("fee_ledgers").
		Where("school_id = ? AND status IN ?", schoolID, []string{"due", "partial"}).
		Count(&dueCount).Error; err != nil {
		return nil, err
	}
	if err := h.db.WithContext(ctx).Model(&struct{}{}).Table("fee_ledgers").
		Select("COALESCE(SUM(amount_inr - paid_amount_inr - concession_inr), 0) AS total").
		Where("school_id = ? AND status IN ?", schoolID, []string{"due", "partial"}).
		Scan(&dueAmount).Error; err != nil {
		return nil, err
	}
	s.PendingFees = &FeeTotals{Count: dueCount, AmountINR: dueAmount.Total}

	// Recent announcements.
	if err := h.db.WithContext(ctx).Model(&struct{}{}).Table("announcements").
		Select("id, title, publish_at").
		Where("school_id = ? AND status = ?", schoolID, "published").
		Order("publish_at DESC").Limit(5).
		Scan(&s.RecentAnnouncements).Error; err != nil {
		return nil, err
	}

	// Revenue this month.
	monthStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	var collected struct{ Total int64 }
	if err := h.db.WithContext(ctx).Model(&struct{}{}).Table("payments").
		Select("COALESCE(SUM(amount_inr), 0) AS total").
		Where("school_id = ? AND paid_at >= ?", schoolID, monthStart).
		Scan(&collected).Error; err != nil {
		return nil, err
	}
	rate := 0.0
	if dueAmount.Total+collected.Total > 0 {
		rate = float64(collected.Total) / float64(dueAmount.Total+collected.Total) * 100
	}
	s.Revenue = &Revenue{CollectedINR: collected.Total, CollectionRate: rate}

	return s, nil
}
