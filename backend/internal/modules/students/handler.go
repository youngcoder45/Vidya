package students

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/schoolos/backend/internal/deps"
	"github.com/schoolos/backend/internal/pkg/httpx"
	"github.com/schoolos/backend/internal/pkg/tenant"
	"github.com/schoolos/backend/internal/server/middleware"
)

// Register wires students module routes.
func Register(rg *gin.RouterGroup, d *deps.Deps) {
	repo := NewRepository(d.DB)
	h := &handler{repo: repo}

	grp := rg.Group("/students")
	grp.GET("", middleware.RequirePermission("students.read"), h.list)
	grp.POST("", middleware.RequirePermission("students.write"), h.create)
	grp.GET("/:id", middleware.RequirePermission("students.read"), h.get)
	grp.PUT("/:id", middleware.RequirePermission("students.write"), h.update)
	grp.DELETE("/:id", middleware.RequirePermission("students.write"), h.delete)
	grp.POST("/:id/enroll", middleware.RequirePermission("students.write"), h.enroll)
	grp.POST("/:id/promote", middleware.RequirePermission("students.write"), h.promote)
	grp.GET("/:id/documents", middleware.RequirePermission("students.read"), h.documents)
	grp.POST("/:id/documents", middleware.RequirePermission("students.write"), h.addDocument)
	grp.GET("/:id/guardians", middleware.RequirePermission("students.read"), h.guardians)
	grp.POST("/:id/guardians", middleware.RequirePermission("students.write"), h.addGuardian)
	grp.GET("/:id/history", middleware.RequirePermission("students.read"), h.history)
}

type handler struct{ repo Repository }

type guardianInput struct {
	Name      string `json:"name" binding:"required"`
	Relation  string `json:"relation"`
	Phone     string `json:"phone" binding:"required"`
	Email     string `json:"email"`
	IsPrimary bool   `json:"is_primary"`
}

type createStudentRequest struct {
	FirstName   string          `json:"first_name" binding:"required"`
	LastName    string          `json:"last_name"`
	DOB         time.Time       `json:"dob"`
	Gender      string          `json:"gender"`
	BloodGroup  string          `json:"blood_group"`
	Address     string          `json:"address"`
	MedicalNotes string         `json:"medical_notes"`
	AdmissionNo string          `json:"admission_no" binding:"required"`
	SessionID   uuid.UUID       `json:"session_id"`
	ClassDivisionID uuid.UUID   `json:"class_division_id"`
	RollNo      int             `json:"roll_no"`
	Guardians   []guardianInput `json:"guardians"`
}

func (h *handler) create(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	var req createStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	student := &Student{
		ID: uuid.New(), SchoolID: tc.SchoolID, AdmissionNo: req.AdmissionNo,
		FirstName: req.FirstName, LastName: req.LastName, DOB: req.DOB,
		Gender: req.Gender, BloodGroup: req.BloodGroup, Address: req.Address,
		MedicalNotes: req.MedicalNotes, Status: StudentActive,
	}
	var en *StudentEnrollment
	if req.ClassDivisionID != uuid.Nil {
		en = &StudentEnrollment{
			ID: uuid.New(), SchoolID: tc.SchoolID, SessionID: req.SessionID,
			StudentID: student.ID, ClassDivisionID: req.ClassDivisionID,
			RollNo: req.RollNo, AdmissionDate: time.Now(), Status: EnrollmentActive,
		}
	}
	guardians := make([]Guardian, 0, len(req.Guardians))
	links := make([]StudentGuardian, 0, len(req.Guardians))
	for i, g := range req.Guardians {
		guardians = append(guardians, Guardian{
			ID: uuid.New(), SchoolID: tc.SchoolID, Name: g.Name, Relation: g.Relation,
			Phone: g.Phone, Email: g.Email, IsPrimary: g.IsPrimary,
		})
		links = append(links, StudentGuardian{
			ID: uuid.New(), SchoolID: tc.SchoolID, StudentID: student.ID,
			Relation: g.Relation, Priority: i,
		})
	}
	if err := h.repo.CreateStudentFull(c.Request.Context(), student, en, guardians, links); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusCreated, student)
}

func (h *handler) list(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	p := httpx.ParsePagination(c)
	sessionID, _ := uuid.Parse(c.DefaultQuery("session_id", ""))
	classDivisionID, _ := uuid.Parse(c.DefaultQuery("class_division_id", ""))
	students, total, err := h.repo.ListStudents(c.Request.Context(), tc.SchoolID, sessionID, classDivisionID, c.Query("search"), p.Page, p.Limit)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteList(c, students, p, total)
}

func (h *handler) get(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation)
		return
	}
	student, err := h.repo.GetStudent(c.Request.Context(), tc.SchoolID, id)
	if err != nil {
		httpx.WriteError(c, httpx.ErrNotFound)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, student)
}

func (h *handler) update(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation)
		return
	}
	existing, err := h.repo.GetStudent(c.Request.Context(), tc.SchoolID, id)
	if err != nil {
		httpx.WriteError(c, httpx.ErrNotFound)
		return
	}
	if err := c.ShouldBindJSON(existing); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	if err := h.repo.UpdateStudent(c.Request.Context(), existing); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, existing)
}

func (h *handler) delete(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation)
		return
	}
	if err := h.repo.SoftDeleteStudent(c.Request.Context(), tc.SchoolID, id); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, gin.H{"deleted": true})
}

type enrollRequest struct {
	SessionID       uuid.UUID `json:"session_id" binding:"required"`
	ClassDivisionID uuid.UUID `json:"class_division_id" binding:"required"`
	RollNo          int       `json:"roll_no"`
}

func (h *handler) enroll(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	studentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation)
		return
	}
	var req enrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	en := &StudentEnrollment{
		ID: uuid.New(), SchoolID: tc.SchoolID, SessionID: req.SessionID,
		StudentID: studentID, ClassDivisionID: req.ClassDivisionID,
		RollNo: req.RollNo, AdmissionDate: time.Now(), Status: EnrollmentActive,
	}
	if err := h.repo.Enroll(c.Request.Context(), en); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusCreated, en)
}

type promoteRequest struct {
	FromSessionID   uuid.UUID `json:"from_session_id" binding:"required"`
	ToSessionID     uuid.UUID `json:"to_session_id" binding:"required"`
	ToClassDivisionID uuid.UUID `json:"to_class_division_id" binding:"required"`
	RollNo          int       `json:"roll_no"`
}

func (h *handler) promote(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	studentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation)
		return
	}
	var req promoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	if err := h.repo.Promote(c.Request.Context(), tc.SchoolID, studentID, req.FromSessionID, req.ToSessionID, req.ToClassDivisionID, req.RollNo); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, gin.H{"promoted": true})
}

func (h *handler) documents(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	studentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation)
		return
	}
	docs, err := h.repo.ListDocuments(c.Request.Context(), tc.SchoolID, studentID)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, docs)
}

type addDocumentRequest struct {
	DocType string `json:"doc_type" binding:"required"`
	FileURL string `json:"file_url" binding:"required"`
}

func (h *handler) addDocument(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	studentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation)
		return
	}
	var req addDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	doc := &StudentDocument{
		ID: uuid.New(), SchoolID: tc.SchoolID, StudentID: studentID,
		DocType: req.DocType, FileURL: req.FileURL,
	}
	if err := h.repo.CreateDocument(c.Request.Context(), doc); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusCreated, doc)
}

func (h *handler) guardians(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	studentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation)
		return
	}
	list, err := h.repo.ListGuardiansByStudent(c.Request.Context(), tc.SchoolID, studentID)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, list)
}

func (h *handler) addGuardian(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	studentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation)
		return
	}
	var req guardianInput
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, httpx.ErrValidation.WithDetails(err.Error()))
		return
	}
	g := &Guardian{
		ID: uuid.New(), SchoolID: tc.SchoolID, Name: req.Name, Relation: req.Relation,
		Phone: req.Phone, Email: req.Email, IsPrimary: req.IsPrimary,
	}
	if err := h.repo.CreateGuardian(c.Request.Context(), g); err != nil {
		httpx.WriteError(c, err)
		return
	}
	if err := h.repo.LinkGuardian(c.Request.Context(), tc.SchoolID, studentID, g.ID, req.Relation, 0); err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusCreated, g)
}

func (h *handler) history(c *gin.Context) {
	tc := tenant.MustFrom(c.Request.Context())
	studentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.WriteError(c, httpx.ErrValidation)
		return
	}
	history, err := h.repo.ListEnrollmentHistory(c.Request.Context(), tc.SchoolID, studentID)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	httpx.WriteJSON(c, http.StatusOK, history)
}
