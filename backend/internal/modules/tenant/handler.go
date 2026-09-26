package tenant

import (
	"errors"
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

// Register wires tenant module routes. Platform routes are registered on a
// separate group by the router; here we handle school-scoped setup routes.
func Register(rg *gin.RouterGroup, d *deps.Deps) {
	repo := NewRepository(d.DB)
	h := &handler{repo: repo}

	setup := rg.Group("", middleware.RequirePermission("tenant.*"))
	setup.GET("/sessions", h.listSessions)
	setup.POST("/sessions", h.createSession)
	setup.POST("/sessions/:id/activate", h.activateSession)
	setup.GET("/classes", h.listClassGroups)
	setup.POST("/classes", h.createClassGroup)
	setup.GET("/divisions", h.listDivisions)
	setup.POST("/divisions", h.createDivision)
	setup.GET("/class-divisions", h.listClassDivisions)
	setup.POST("/class-divisions", h.createClassDivision)
	setup.POST("/class-divisions/:id/class-teacher", h.setClassTeacher)
	setup.GET("/subjects", h.listSubjects)
	setup.POST("/subjects", h.createSubject)
	setup.GET("/class-subjects", h.listClassSubjects)
	setup.POST("/class-subjects", h.createClassSubject)
	setup.GET("/teachers", h.listTeachers)
	setup.GET("/teachers/:id", h.getTeacher)
	setup.POST("/teachers", h.createTeacher)
}

// RegisterPlatform wires platform-admin school management routes.
func RegisterPlatform(rg *gin.RouterGroup, d *deps.Deps) {
	repo := NewRepository(d.DB)
	h := &handler{repo: repo}
	platform := rg.Group("/schools", middleware.RequirePermission("platform.schools.manage"))
	platform.GET("", h.listSchools)
	platform.POST("", h.createSchool)
	platform.GET("/:id", h.getSchool)
}

type handler struct{ repo Repository }

func (h *handler) createSchool(c *gin.Context) {
	var req struct {
		Name  string `json:"name" binding:"required"`
		Code  string `json:"code" binding:"required,min=3,max=40"`
		Board string `json:"board"`
		Phone string `json:"phone"`
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	school := &School{
		ID: uuid.New(), Name: req.Name, Code: req.Code, Board: req.Board,
		Phone: req.Phone, Email: req.Email, Status: SchoolActive,
		Branding: `{"primary_color":"#1B5E20"}`,
	}
	if err := h.repo.CreateSchool(c.Request.Context(), school); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusCreated, school)
}

func (h *handler) listSchools(c *gin.Context) {
	schools, err := h.repo.ListSchools(c.Request.Context())
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, schools)
}

func (h *handler) getSchool(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation)
		return
	}
	school, err := h.repo.GetSchool(c.Request.Context(), id)
	if err != nil {
		httpx.WriteError(c, httpx.ErrNotFound)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, school)
}

func (h *handler) createSession(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	var req struct {
		Name     string    `json:"name" binding:"required"`
		StartsOn time.Time `json:"starts_on" binding:"required"`
		EndsOn   time.Time `json:"ends_on" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	session := &AcademicSession{
		ID: uuid.New(), SchoolID: tc.SchoolID, Name: req.Name,
		StartsOn: req.StartsOn, EndsOn: req.EndsOn,
	}
	if err := h.repo.CreateSession(c.Request.Context(), session); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusCreated, session)
}

func (h *handler) listSessions(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	sessions, err := h.repo.ListSessions(c.Request.Context(), tc.SchoolID)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, sessions)
}

func (h *handler) activateSession(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation)
		return
	}
	if err := h.repo.ActivateSession(c.Request.Context(), tc.SchoolID, id); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, gin.H{"activated": true})
}

func (h *handler) createClassGroup(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	var req struct {
		Name  string `json:"name" binding:"required"`
		Level int    `json:"level"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	group := &ClassGroup{ID: uuid.New(), SchoolID: tc.SchoolID, Name: req.Name, Level: req.Level}
	if err := h.repo.CreateClassGroup(c.Request.Context(), group); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusCreated, group)
}

func (h *handler) listClassGroups(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	groups, err := h.repo.ListClassGroups(c.Request.Context(), tc.SchoolID)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, groups)
}

func (h *handler) createDivision(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	div := &Division{ID: uuid.New(), SchoolID: tc.SchoolID, Name: req.Name}
	if err := h.repo.CreateDivision(c.Request.Context(), div); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusCreated, div)
}

func (h *handler) listDivisions(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	divs, err := h.repo.ListDivisions(c.Request.Context(), tc.SchoolID)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, divs)
}

func (h *handler) createClassDivision(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	var req struct {
		SessionID    uuid.UUID `json:"session_id" binding:"required"`
		ClassGroupID uuid.UUID `json:"class_group_id" binding:"required"`
		DivisionID   uuid.UUID `json:"division_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	cd := &ClassDivision{
		ID: uuid.New(), SchoolID: tc.SchoolID, SessionID: req.SessionID,
		ClassGroupID: req.ClassGroupID, DivisionID: req.DivisionID,
	}
	if err := h.repo.CreateClassDivision(c.Request.Context(), cd); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusCreated, cd)
}

func (h *handler) listClassDivisions(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	sessionID, err := uuid.Parse(c.DefaultQuery("session_id", ""))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails("session_id query param required"))
		return
	}
	list, err := h.repo.ListClassDivisions(c.Request.Context(), tc.SchoolID, sessionID)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, list)
}

func (h *handler) setClassTeacher(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation)
		return
	}
	var req struct {
		TeacherID uuid.UUID `json:"teacher_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	if err := h.repo.SetClassTeacher(c.Request.Context(), tc.SchoolID, id, req.TeacherID); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, gin.H{"updated": true})
}

func (h *handler) createSubject(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	var req struct {
		Name       string `json:"name" binding:"required"`
		Code       string `json:"code"`
		IsLanguage bool   `json:"is_language"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	sub := &Subject{ID: uuid.New(), SchoolID: tc.SchoolID, Name: req.Name, Code: req.Code, IsLanguage: req.IsLanguage}
	if err := h.repo.CreateSubject(c.Request.Context(), sub); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusCreated, sub)
}

func (h *handler) listSubjects(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	subs, err := h.repo.ListSubjects(c.Request.Context(), tc.SchoolID)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, subs)
}

func (h *handler) createClassSubject(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	var req struct {
		SessionID       uuid.UUID  `json:"session_id" binding:"required"`
		ClassDivisionID uuid.UUID  `json:"class_division_id" binding:"required"`
		SubjectID       uuid.UUID  `json:"subject_id" binding:"required"`
		TeacherID       *uuid.UUID `json:"teacher_id"`
		IsElective      bool       `json:"is_elective"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	cs := &ClassSubject{
		ID: uuid.New(), SchoolID: tc.SchoolID, SessionID: req.SessionID,
		ClassDivisionID: req.ClassDivisionID, SubjectID: req.SubjectID,
		TeacherID: req.TeacherID, IsElective: req.IsElective,
	}
	if err := h.repo.CreateClassSubject(c.Request.Context(), cs); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusCreated, cs)
}

func (h *handler) listClassSubjects(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	list, err := h.repo.ListClassSubjects(c.Request.Context(), tc.SchoolID)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, list)
}

func (h *handler) createTeacher(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	var req struct {
		UserID        uuid.UUID `json:"user_id" binding:"required"`
		EmployeeCode  string    `json:"employee_code" binding:"required"`
		Designation   string    `json:"designation"`
		Qualification string    `json:"qualification"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	t := &Teacher{
		ID: uuid.New(), SchoolID: tc.SchoolID, UserID: req.UserID,
		EmployeeCode: req.EmployeeCode, Designation: req.Designation,
		Qualification: req.Qualification, Status: TeacherActive,
	}
	if err := h.repo.CreateTeacher(c.Request.Context(), t); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusCreated, t)
}

func (h *handler) listTeachers(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	list, err := h.repo.ListTeachers(c.Request.Context(), tc.SchoolID)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, list)
}

func (h *handler) getTeacher(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation)
		return
	}
	t, err := h.repo.GetTeacher(c.Request.Context(), tc.SchoolID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpx.WriteError(c, httpx.ErrNotFound)
			return
		}
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, t)
}
