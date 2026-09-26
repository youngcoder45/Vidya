// Vidya API — composition root. Wires configuration, persistence,
// caching, the event bus, and the HTTP server.
package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/vidya/backend/internal/config"
	"github.com/vidya/backend/internal/db"
	"github.com/vidya/backend/internal/deps"
	"github.com/vidya/backend/internal/events"
	"github.com/vidya/backend/internal/modules/notifications"
	"github.com/vidya/backend/internal/pkg/audit"
	"github.com/vidya/backend/internal/pkg/jwtutil"
	"github.com/vidya/backend/internal/seeds"
	"github.com/vidya/backend/internal/server"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	// Postgres.
	dbConn, err := db.Open(cfg, log)
	if err != nil {
		return err
	}
	if cfg.AppEnv != "production" {
		if err := db.MigrateDev(dbConn, log); err != nil {
			return err
		}
	} else if cfg.MigrateOnStart {
		// Production schema is versioned SQL, applied once per file.
		if err := db.MigrateSQL(context.Background(), dbConn, cfg.MigrationsDir, log); err != nil {
			return err
		}
	}

	// Redis (optional at boot — in-memory fallbacks cover its absence).
	rdb := connectRedis(cfg, log)

	bus := events.New(log)
	auditWriter := audit.NewWriter(dbConn)
	issuer := jwtutil.New(cfg.JWTSecret, cfg.JWTAccessTTL)

	d := &deps.Deps{
		Cfg: cfg, DB: dbConn, RDB: rdb, Bus: bus,
		Audit: auditWriter, Issuer: issuer, Log: log,
	}

	// Domain events → in-app notifications (the microservice seam).
	notifications.SubscribeToEvents(bus, notifications.NewRepository(dbConn), log)

	if cfg.SeedOnStart {
		if err := seeds.Run(dbConn, log); err != nil {
			return err
		}
	}

	return server.New(d).Run()
}

func connectRedis(cfg *config.Config, log *slog.Logger) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Warn("redis unavailable — using in-memory fallbacks", "addr", cfg.RedisAddr, "err", err)
		return nil
	}
	log.Info("redis connected", "addr", cfg.RedisAddr)
	return rdb
}
