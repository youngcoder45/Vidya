// Command migrate applies versioned SQL migrations to the configured database
// and exits. Run with: go run ./cmd/migrate
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/schoolos/backend/internal/config"
	"github.com/schoolos/backend/internal/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "err", err)
		os.Exit(1)
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	conn, err := db.Open(cfg, log)
	if err != nil {
		slog.Error("db connect failed", "err", err)
		os.Exit(1)
	}
	if err := db.MigrateSQL(context.Background(), conn, cfg.MigrationsDir, log); err != nil {
		slog.Error("migrate failed", "err", err)
		os.Exit(1)
	}
	log.Info("migrations complete")
}
