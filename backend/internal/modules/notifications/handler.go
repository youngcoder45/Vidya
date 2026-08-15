package notifications

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/schoolos/backend/internal/deps"
	"github.com/schoolos/backend/internal/events"
	"github.com/schoolos/backend/internal/pkg/ctxuser"
	"github.com/schoolos/backend/internal/pkg/httpx"
	"github.com/schoolos/backend/internal/pkg/tenant"
	"github.com/schoolos/backend/internal/server/middleware"
)

// Register wires notifications module routes.
func Register(rg *gin.RouterGroup, d *deps.Deps) {
	repo := NewRepository(d.DB)
	h := &handler{repo: repo}

	grp := rg.Group("/notifications", middleware.Auth(d.Issuer), middleware.Tenant())
	grp.GET("", h.list)
	grp.POST("/:id/read", h.markRead)
	grp.POST("/read-all", h.markAllRead)
	grp.GET("/preferences", h.preferences)
	grp.PUT("/preferences", h.savePreference)
}

// SubscribeToEvents demonstrates the domain-event seam: domain events become
// in-app notifications. The worker (separate process later) replaces this
// subscriber for push/email fan-out.
func SubscribeToEvents(bus *events.Bus, dbRepo Repository, log *slog.Logger) {
	bus.Subscribe(func(ctx context.Context, ev events.Event) {
		if ev.UserID == uuid.Nil {
			return
		}
		n := &Notification{
			ID: uuid.New(), SchoolID: ev.SchoolID, UserID: ev.UserID,
			Type: ev.Type, Title: humanTitle(ev.Type),
			Channel: ChannelInApp, CreatedAt: time.Now(),
		}
		if err := dbRepo.Create(ctx, n); err != nil {
			log.Warn("notification create failed", "err", err, "type", ev.Type)
		}
	})
}

func humanTitle(evType string) string {
	titles := map[string]string{
		EvFeePaid:            "Payment received",
		EvResultsPublished:   "Results published",
		EvHomeworkAssigned:   "New homework assigned",
		EvAnnouncement:       "New announcement",
		EvExamSchedule:       "Exam schedule published",
		EvFeeDue:             "Fee due reminder",
		EvHoliday:            "Holiday announced",
		EvSalaryPaid:         "Salary credited",
	}
	if t, ok := titles[evType]; ok {
		return t
	}
	return "New update"
}

type handler struct{ repo Repository }

func (h *handler) list(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	uid := ctxuser.MustUserID(c.Request.Context())
	unreadOnly := c.Query("unread") == "true"
	list, err := h.repo.ListForUser(c.Request.Context(), tc.SchoolID, uid, unreadOnly, 50)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, list)
}

func (h *handler) markRead(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	uid := ctxuser.MustUserID(c.Request.Context())
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation)
		return
	}
	if err := h.repo.MarkRead(c.Request.Context(), tc.SchoolID, uid, id); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, gin.H{"read": true})
}

func (h *handler) markAllRead(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	uid := ctxuser.MustUserID(c.Request.Context())
	if err := h.repo.MarkAllRead(c.Request.Context(), tc.SchoolID, uid); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, gin.H{"read": true})
}

func (h *handler) preferences(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	uid := ctxuser.MustUserID(c.Request.Context())
	prefs, err := h.repo.ListPreferences(c.Request.Context(), tc.SchoolID, uid)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, prefs)
}

func (h *handler) savePreference(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	uid := ctxuser.MustUserID(c.Request.Context())
	var req struct {
		EventType string `json:"event_type" binding:"required"`
		Channel   string `json:"channel" binding:"required,oneof=inapp push email sms"`
		Enabled   bool   `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	pref := &NotificationPreference{
		ID: uuid.New(), SchoolID: tc.SchoolID, UserID: uid,
		EventType: req.EventType, Channel: req.Channel, Enabled: req.Enabled,
	}
	if err := h.repo.UpsertPreference(c.Request.Context(), pref); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, pref)
}
