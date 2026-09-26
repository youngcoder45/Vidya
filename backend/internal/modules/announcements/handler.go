package announcements

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/vidya/backend/internal/deps"
	"github.com/vidya/backend/internal/pkg/ctxuser"
	"github.com/vidya/backend/internal/pkg/httpx"
	"github.com/vidya/backend/internal/pkg/tenant"
	"github.com/vidya/backend/internal/server/middleware"
)

// Register wires announcements + calendar routes.
func Register(rg *gin.RouterGroup, d *deps.Deps) {
	repo := NewRepository(d.DB)
	h := &handler{repo: repo}

	ann := rg.Group("/announcements")
	ann.GET("", middleware.RequirePermission("announcements.read"), h.list)
	ann.POST("", middleware.RequirePermission("announcements.write"), h.create)
	ann.GET("/:id", middleware.RequirePermission("announcements.read"), h.get)
	ann.POST("/:id/publish", middleware.RequirePermission("announcements.write"), h.publish)
	ann.POST("/:id/read", middleware.Auth(d.Issuer), middleware.Tenant(), h.markRead)
	ann.GET("/:id/read-receipts", middleware.RequirePermission("announcements.read"), h.readReceipts)

	cal := rg.Group("/calendar", middleware.RequirePermission("announcements.read"))
	cal.GET("/events", h.listEvents)
	cal.POST("/events", middleware.RequirePermission("announcements.write"), h.createEvent)
}

type handler struct{ repo Repository }

type createAnnouncementRequest struct {
	Title         string     `json:"title" binding:"required"`
	Body          string     `json:"body" binding:"required"`
	AudienceScope string     `json:"audience_scope" binding:"required,oneof=school class division student staff"`
	AudienceClassDivisionID *uuid.UUID `json:"audience_class_division_id"`
	AudienceStudentID       *uuid.UUID `json:"audience_student_id"`
	PublishAt     *time.Time `json:"publish_at"`
	ExpiresAt     *time.Time `json:"expires_at"`
}

func (h *handler) create(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	createdBy := ctxuser.MustUserID(c.Request.Context())
	var req createAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	status := StatusDraft
	if req.PublishAt != nil && req.PublishAt.After(time.Now()) {
		status = StatusScheduled
	}
	a := &Announcement{
		ID: uuid.New(), SchoolID: tc.SchoolID, Title: req.Title, Body: req.Body,
		AudienceScope: req.AudienceScope, AudienceClassDivisionID: req.AudienceClassDivisionID,
		AudienceStudentID: req.AudienceStudentID, PublishAt: req.PublishAt,
		ExpiresAt: req.ExpiresAt, Status: status, CreatedBy: createdBy,
	}
	if err := h.repo.Create(c.Request.Context(), a); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusCreated, a)
}

func (h *handler) list(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	// Only writers may inspect drafts; everyone else sees published only.
	canManage := ctxuser.HasPermission(c.Request.Context(), "announcements.write")
	status := c.Query("status")
	if !canManage {
		status = ""
	}
	list, err := h.repo.List(c.Request.Context(), tc.SchoolID, status, !canManage)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, list)
}

func (h *handler) get(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation)
		return
	}
	a, err := h.repo.Get(c.Request.Context(), tc.SchoolID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpx.WriteError(c, httpx.ErrNotFound)
			return
		}
		httpx.WriteError(c, err)
		return
	}
	if !ctxuser.HasPermission(c.Request.Context(), "announcements.write") {
		dueToRead := a.Status == StatusScheduled && a.PublishAt != nil && !a.PublishAt.After(time.Now())
		if a.Status != StatusPublished && !dueToRead {
			httpx.WriteError(c, httpx.ErrNotFound)
			return
		}
	}
	httpx.WriteJSON(c, http.StatusOK, a)
}

func (h *handler) publish(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation)
		return
	}
	a, err := h.repo.Get(c.Request.Context(), tc.SchoolID, id)
	if err != nil {
		httpx.WriteError(c, httpx.ErrNotFound)
		return
	}
	if err := h.repo.SetStatus(c.Request.Context(), tc.SchoolID, id, StatusPublished); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, gin.H{"published": true, "audience_scope": a.AudienceScope})
}

func (h *handler) markRead(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	uid := ctxuser.MustUserID(c.Request.Context())
	annID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation)
		return
	}
	if err := h.repo.MarkRead(c.Request.Context(), tc.SchoolID, annID, uid); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, gin.H{"read": true})
}

func (h *handler) readReceipts(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	annID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation)
		return
	}
	count, err := h.repo.ReadCount(c.Request.Context(), tc.SchoolID, annID)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, gin.H{"read_count": count})
}

func (h *handler) listEvents(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	sessionID, _ := uuid.Parse(c.DefaultQuery("session_id", ""))
	list, err := h.repo.ListEvents(c.Request.Context(), tc.SchoolID, sessionID)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, list)
}

func (h *handler) createEvent(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	createdBy := ctxuser.MustUserID(c.Request.Context())
	var req struct {
		SessionID     uuid.UUID `json:"session_id" binding:"required"`
		Title         string    `json:"title" binding:"required"`
		Type          string    `json:"type" binding:"required,oneof=holiday event activity"`
		StartsOn      time.Time `json:"starts_on" binding:"required"`
		EndsOn        time.Time `json:"ends_on" binding:"required"`
		AudienceScope string    `json:"audience_scope"`
		Description   string    `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	ev := &HolidayEvent{
		ID: uuid.New(), SchoolID: tc.SchoolID, SessionID: req.SessionID,
		Title: req.Title, Type: req.Type, StartsOn: req.StartsOn, EndsOn: req.EndsOn,
		AudienceScope: req.AudienceScope, Description: req.Description, CreatedBy: createdBy,
	}
	if err := h.repo.CreateEvent(c.Request.Context(), ev); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusCreated, ev)
}
