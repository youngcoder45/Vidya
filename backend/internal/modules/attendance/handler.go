package attendance

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/vidya/backend/internal/deps"
	"github.com/vidya/backend/internal/pkg/ctxuser"
	"github.com/vidya/backend/internal/pkg/httpx"
	"github.com/vidya/backend/internal/pkg/tenant"
	"github.com/vidya/backend/internal/server/middleware"
)

// Register wires attendance module routes.
func Register(rg *gin.RouterGroup, d *deps.Deps) {
	repo := NewRepository(d.DB)
	h := &handler{repo: repo}

	grp := rg.Group("/attendance")
	grp.GET("/classes/:classDivisionID/daily", middleware.RequirePermission("attendance.read"), h.daily)
	grp.POST("/daily", middleware.RequirePermission("attendance.mark"), h.mark)
	grp.GET("/students/:studentID", middleware.RequirePermission("attendance.read"), h.studentHistory)
	grp.GET("/analytics/classes/:classDivisionID", middleware.RequirePermission("attendance.read"), h.classAnalytics)
}

type handler struct{ repo Repository }

func (h *handler) daily(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	cdID, err := uuid.Parse(c.Param("classDivisionID"))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation)
		return
	}
	date, err := time.Parse("2006-01-02", c.DefaultQuery("date", time.Now().Format("2006-01-02")))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails("date must be YYYY-MM-DD"))
		return
	}
	records, err := h.repo.ListByClassDate(c.Request.Context(), tc.SchoolID, cdID, date)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, records)
}

type markRequest struct {
	ClassDivisionID uuid.UUID `json:"class_division_id" binding:"required"`
	Date            string    `json:"date" binding:"required"`
	SubjectID       *uuid.UUID `json:"subject_id"`
	Records         []struct {
		StudentID uuid.UUID `json:"student_id" binding:"required"`
		Status    string    `json:"status" binding:"required"`
	} `json:"records" binding:"required,min=1"`
}

func (h *handler) mark(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	markedBy := ctxuser.MustUserID(c.Request.Context())
	var req markRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails("date must be YYYY-MM-DD"))
		return
	}
	if date.After(time.Now().AddDate(0, 0, 1)) {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails("date cannot be in the future"))
		return
	}
	records := make([]AttendanceRecord, 0, len(req.Records))
	for _, r := range req.Records {
		if !ValidStatuses[r.Status] {
			httpx.WriteError(c, httpx.ErrValidation.WithDetails("invalid status: "+r.Status))
			return
		}
		records = append(records, AttendanceRecord{
			ID: uuid.New(), SchoolID: tc.SchoolID, ClassDivisionID: req.ClassDivisionID,
			SubjectID: req.SubjectID, StudentID: r.StudentID, Date: date,
			Status: r.Status, MarkedBy: markedBy,
		})
	}
	// Verify every student is actually enrolled in this class/school.
	ids := make([]uuid.UUID, 0, len(req.Records))
	seen := make(map[uuid.UUID]struct{}, len(req.Records))
	for _, r := range req.Records {
		if _, ok := seen[r.StudentID]; !ok {
			seen[r.StudentID] = struct{}{}
			ids = append(ids, r.StudentID)
		}
	}
	if err := h.repo.ValidateRoster(c.Request.Context(), tc.SchoolID, req.ClassDivisionID, ids); err != nil {
		if errors.Is(err, ErrStudentNotInClass) {
			httpx.WriteError(c, httpx.ErrValidation.WithDetails("one or more students are not enrolled in this class"))
			return
		}
		httpx.WriteError(c, err)
		return
	}
	if err := h.repo.UpsertDaily(c.Request.Context(), records); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, gin.H{"saved": len(records)})
}

func (h *handler) studentHistory(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	studentID, err := uuid.Parse(c.Param("studentID"))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation)
		return
	}
	from, err := time.Parse("2006-01-02", c.DefaultQuery("from", "2000-01-01"))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails("from must be YYYY-MM-DD"))
		return
	}
	to, err := time.Parse("2006-01-02", c.DefaultQuery("to", time.Now().Format("2006-01-02")))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails("to must be YYYY-MM-DD"))
		return
	}
	records, err := h.repo.ListByStudent(c.Request.Context(), tc.SchoolID, studentID, from, to)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, records)
}

// classAnalytics returns the daily attendance % series for a class.
func (h *handler) classAnalytics(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	cdID, err := uuid.Parse(c.Param("classDivisionID"))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation)
		return
	}
	from, err := time.Parse("2006-01-02", c.DefaultQuery("from", time.Now().AddDate(0, -1, 0).Format("2006-01-02")))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails("from must be YYYY-MM-DD"))
		return
	}
	to, err := time.Parse("2006-01-02", c.DefaultQuery("to", time.Now().Format("2006-01-02")))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails("to must be YYYY-MM-DD"))
		return
	}

	type dayPoint struct {
		Date       string  `json:"date"`
		Present    int64   `json:"present"`
		Total      int64   `json:"total"`
		Percentage float64 `json:"percentage"`
	}
	points := []dayPoint{}
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		present, total, err := h.repo.ClassDailyPercentage(c.Request.Context(), tc.SchoolID, cdID, d)
		if err != nil {
			httpx.WriteError(c, err)
			return
		}
		pct := 0.0
		if total > 0 {
			pct = float64(present) / float64(total) * 100
		}
		points = append(points, dayPoint{Date: d.Format("2006-01-02"), Present: present, Total: total, Percentage: pct})
	}
	httpx.WriteJSON(c, http.StatusOK, points)
}
