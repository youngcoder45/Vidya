package db

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"

	"gorm.io/gorm"
)

// MigrateSQL applies the numbered .sql files in dir once each, recording
// applied files in schema_migrations. It is the production counterpart to the
// dev-only AutoMigrate and keeps deployment from running on an empty schema.
func MigrateSQL(ctx context.Context, db *gorm.DB, dir string, log *slog.Logger) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return fmt.Errorf("db: list migrations: %w", err)
	}
	if len(files) == 0 {
		return fmt.Errorf("db: no migrations found in %s", dir)
	}
	sort.Strings(files)

	if err := db.WithContext(ctx).Exec(
		`CREATE TABLE IF NOT EXISTS schema_migrations (
			filename   TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`).Error; err != nil {
		return fmt.Errorf("db: create schema_migrations: %w", err)
	}

	for _, path := range files {
		name := filepath.Base(path)
		var applied int64
		if err := db.WithContext(ctx).Table("schema_migrations").
			Where("filename = ?", name).Count(&applied).Error; err != nil {
			return err
		}
		if applied > 0 {
			continue
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("db: read %s: %w", name, err)
		}
		err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec(string(body)).Error; err != nil {
				return err
			}
			return tx.Exec("INSERT INTO schema_migrations (filename) VALUES (?)", name).Error
		})
		if err != nil {
			return fmt.Errorf("db: apply %s: %w", name, err)
		}
		log.Info("migration applied", "file", name)
	}
	return nil
}
