// Package db opens the Postgres connection and (in dev) runs AutoMigrate for
// the scaffold's implemented modules. Production schema is versioned in
// migrations/001_init.sql and applied by golang-migrate in CI/CD — AutoMigrate
// is a dev convenience that never drops data.
package db

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/schoolos/backend/internal/config"
	"github.com/schoolos/backend/internal/modules/announcements"
	"github.com/schoolos/backend/internal/modules/attendance"
	"github.com/schoolos/backend/internal/modules/auth"
	"github.com/schoolos/backend/internal/modules/fees"
	"github.com/schoolos/backend/internal/modules/notifications"
	"github.com/schoolos/backend/internal/modules/students"
	"github.com/schoolos/backend/internal/modules/tenant"
	"github.com/schoolos/backend/internal/pkg/audit"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open connects to Postgres and returns a configured *gorm.DB.
func Open(cfg *config.Config, log *slog.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	gormLogLevel := logger.Warn
	if cfg.AppEnv == "local" {
		gormLogLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("db: connect: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	log.Info("db connected", "host", cfg.DBHost, "db", cfg.DBName)
	return db, nil
}

// MigrateDev runs AutoMigrate for the implemented modules (dev only).
func MigrateDev(db *gorm.DB, log *slog.Logger) error {
	models := []any{
		&auth.User{}, &auth.Role{}, &auth.Permission{}, &auth.UserRole{},
		&auth.UserDevice{}, &auth.AuthSession{}, &auth.OtpCode{},

		&tenant.Plan{}, &tenant.School{}, &tenant.AcademicSession{},
		&tenant.ClassGroup{}, &tenant.Division{}, &tenant.ClassDivision{},
		&tenant.Subject{}, &tenant.ClassSubject{}, &tenant.Teacher{},

		&students.Student{}, &students.Guardian{}, &students.StudentGuardian{},
		&students.StudentEnrollment{}, &students.StudentDocument{},

		&attendance.AttendanceRecord{},

		&fees.FeeHead{}, &fees.FeeStructure{}, &fees.FeeLedger{},
		&fees.FeeInstallment{}, &fees.FeePaymentOrder{}, &fees.Payment{},
		&fees.PaymentAllocation{}, &fees.Receipt{}, &fees.Refund{},

		&announcements.Announcement{}, &announcements.AnnouncementRead{},
		&announcements.HolidayEvent{},

		&notifications.Notification{}, &notifications.NotificationDelivery{},
		&notifications.NotificationPreference{}, &notifications.NotificationTemplate{},

		&audit.Entry{},
	}
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("db: automigrate: %w", err)
	}
	log.Info("db migrated (dev automigrate)")
	return nil
}
