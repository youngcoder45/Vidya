// Package seeds populates the platform catalog (roles/permissions) and a
// demo tenant for local development. Never run against production.
package seeds

import (
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/vidya/backend/internal/modules/auth"
	"github.com/vidya/backend/internal/modules/students"
	"github.com/vidya/backend/internal/modules/tenant"
	"github.com/vidya/backend/internal/pkg/passwd"
)

// Permission catalog (code → name/module). Handlers reference these codes;
// wildcard entries like "fees.*" grant every permission under that module.
var permissionCatalog = []struct{ Code, Name, Module string }{
	{"platform.schools.manage", "Manage schools (platform)", "platform"},
	{"tenant.*", "Manage sessions/classes/subjects/teachers", "tenant"},
	{"students.read", "View student profiles", "students"},
	{"students.write", "Create/edit students", "students"},
	{"students.*", "All student operations", "students"},
	{"attendance.read", "View attendance", "attendance"},
	{"attendance.mark", "Mark attendance", "attendance"},
	{"attendance.*", "All attendance operations", "attendance"},
	{"fees.read", "View fee data", "fees"},
	{"fees.manage", "Manage fee structures", "fees"},
	{"fees.payment.create", "Create fee payments", "fees"},
	{"fees.*", "All fee operations", "fees"},
	{"homework.read", "View homework", "homework"},
	{"homework.write", "Create/grade homework", "homework"},
	{"homework.*", "All homework operations", "homework"},
	{"exams.read", "View exams and results", "exams"},
	{"exams.manage", "Manage exams and marks", "exams"},
	{"exams.*", "All exam operations", "exams"},
	{"payroll.read", "View payroll", "payroll"},
	{"payroll.manage", "Run payroll", "payroll"},
	{"payroll.*", "All payroll operations", "payroll"},
	{"announcements.read", "View announcements", "announcements"},
	{"announcements.write", "Create/publish announcements", "announcements"},
	{"announcements.*", "All announcement operations", "announcements"},
	{"dashboard.read", "View dashboard", "dashboard"},
}

// Role definitions (code → name/description). Created idempotently.
var roleDefinitions = []struct{ Code, Name, Description string }{
	{"platform_admin", "Platform Admin", "Manages tenants and platform configuration"},
	{"school_admin", "School Admin", "Runs the school: students, fees, exams, payroll, announcements"},
	{"accountant", "Accountant", "Fee collection, receipts, financial reports"},
	{"teacher", "Teacher", "Attendance, homework, marks for assigned classes"},
	{"parent", "Parent", "Access to own children's data"},
	{"student", "Student", "Access to own data"},
}

// Run applies the seed data idempotently.
func Run(db *gorm.DB, log *slog.Logger) error {
	if err := seedPermissionsAndRoles(db); err != nil {
		return err
	}
	if err := seedPlatformAdmin(db); err != nil {
		return err
	}
	if err := seedDemoSchool(db); err != nil {
		return err
	}
	log.Info("seeds applied")
	return nil
}

func seedPermissionsAndRoles(db *gorm.DB) error {
	// 1. Roles (idempotent).
	for _, r := range roleDefinitions {
		var role auth.Role
		err := db.Where("code = ?", r.Code).First(&role).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			role = auth.Role{ID: uuid.New(), Code: r.Code, Name: r.Name, Description: r.Description}
			if err := db.Create(&role).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	// 2. Permissions (idempotent).
	permByCode := map[string]auth.Permission{}
	for _, p := range permissionCatalog {
		var perm auth.Permission
		err := db.Where("code = ?", p.Code).First(&perm).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			perm = auth.Permission{ID: uuid.New(), Code: p.Code, Name: p.Name, Module: p.Module}
			if err := db.Create(&perm).Error; err != nil {
			return err
			}
		} else if err != nil {
			return err
		}
		permByCode[p.Code] = perm
	}

	// 3. Role → permission grants (idempotent; composite PK acts as the key).
	rolePerms := map[string][]string{
		"school_admin": {"tenant.*", "students.*", "attendance.*", "fees.*", "homework.*", "exams.*", "payroll.*", "announcements.*", "dashboard.read"},
		"accountant":   {"students.read", "fees.*", "dashboard.read"},
		"teacher":      {"students.read", "attendance.*", "homework.*", "exams.*", "announcements.read", "dashboard.read"},
		"parent":       {"students.read", "fees.read", "fees.payment.create", "announcements.read", "dashboard.read"},
		"student":      {"students.read", "announcements.read", "dashboard.read"},
	}
	for code, perms := range rolePerms {
		var role auth.Role
		if err := db.Where("code = ?", code).First(&role).Error; err != nil {
			return err
		}
		for _, permCode := range perms {
			perm, ok := permByCode[permCode]
			if !ok {
				continue
			}
			var rp auth.RolePermission
			err := db.Where("role_id = ? AND permission_id = ?", role.ID, perm.ID).First(&rp).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				rp = auth.RolePermission{RoleID: role.ID, PermissionID: perm.ID}
				if err := db.Create(&rp).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
		}
	}
	return nil
}

func seedPlatformAdmin(db *gorm.DB) error {
	var admin auth.User
	err := db.Where("email = ?", "admin@vidya.app").First(&admin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		hash, _ := passwd.Hash("admin12345")
		admin = auth.User{
			ID: uuid.New(), FullName: "Platform Admin", Email: "admin@vidya.app",
			PasswordHash: hash, Status: auth.UserActive, IsSuperadmin: true,
		}
		if err := db.Create(&admin).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedDemoSchool(db *gorm.DB) error {
	var school tenant.School
	err := db.Where("code = ?", "DEMO").First(&school).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		school = tenant.School{
			ID: uuid.New(), Name: "Greenwood Public School", Code: "DEMO",
			Board: "CBSE", Timezone: "Asia/Kolkata", Currency: "INR",
			Status: tenant.SchoolActive, Branding: `{"primary_color":"#1B5E20","logo_url":""}`,
		}
		if err := db.Create(&school).Error; err != nil {
			return err
		}

		// Demo admin user with school_admin role.
		hash, _ := passwd.Hash("admin12345")
		admin := auth.User{
			ID: uuid.New(), SchoolID: &school.ID, FullName: "Meera Iyer",
			Email: "principal@greenwood.edu", Phone: "+919800000001",
			PasswordHash: hash, Status: auth.UserActive,
		}
		if err := db.Create(&admin).Error; err != nil {
			return err
		}
		var role auth.Role
		if err := db.Where("code = ?", "school_admin").First(&role).Error; err != nil {
			return err
		}
		if err := db.Create(&auth.UserRole{ID: uuid.New(), SchoolID: school.ID, UserID: admin.ID, RoleID: role.ID}).Error; err != nil {
			return err
		}

		// Demo session + class + division + subject.
		session := tenant.AcademicSession{
			ID: uuid.New(), SchoolID: school.ID, Name: "2026-27",
			StartsOn: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			EndsOn:   time.Date(2027, 3, 31, 0, 0, 0, 0, time.UTC),
			IsActive: true,
		}
		if err := db.Create(&session).Error; err != nil {
			return err
		}
		class := tenant.ClassGroup{ID: uuid.New(), SchoolID: school.ID, Name: "Class 6", Level: 6}
		div := tenant.Division{ID: uuid.New(), SchoolID: school.ID, Name: "A"}
		if err := db.Create(&class).Error; err != nil {
			return err
		}
		if err := db.Create(&div).Error; err != nil {
			return err
		}
		classDiv := tenant.ClassDivision{
			ID: uuid.New(), SchoolID: school.ID, SessionID: session.ID,
			ClassGroupID: class.ID, DivisionID: div.ID,
		}
		if err := db.Create(&classDiv).Error; err != nil {
			return err
		}
		subject := tenant.Subject{ID: uuid.New(), SchoolID: school.ID, Name: "Mathematics", Code: "MATH"}
		if err := db.Create(&subject).Error; err != nil {
			return err
		}

		// A couple of demo students with enrollments + guardians.
		studentsToSeed := []struct {
			FirstName string
			LastName  string
			Admission string
			Roll      int
		}{
			{"Aarav", "Sharma", "ADM-2026-001", 1},
			{"Diya", "Patel", "ADM-2026-002", 2},
		}
		for _, s := range studentsToSeed {
			student := students.Student{
				ID: uuid.New(), SchoolID: school.ID, AdmissionNo: s.Admission,
				FirstName: s.FirstName, LastName: s.LastName,
				DOB: time.Date(2015, 4, 12, 0, 0, 0, 0, time.UTC), Gender: "female",
				Status: students.StudentActive,
			}
			if err := db.Create(&student).Error; err != nil {
				return err
			}
			en := students.StudentEnrollment{
				ID: uuid.New(), SchoolID: school.ID, SessionID: session.ID,
				StudentID: student.ID, ClassDivisionID: classDiv.ID,
				RollNo: s.Roll, AdmissionDate: time.Now(), Status: students.EnrollmentActive,
			}
			if err := db.Create(&en).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
