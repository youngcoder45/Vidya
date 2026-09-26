// Package deps is the composition-root dependency container. Each module's
// Register function receives the deps it needs — explicit constructor
// injection, no service locator.
package deps

import (
	"log/slog"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/vidya/backend/internal/config"
	"github.com/vidya/backend/internal/events"
	"github.com/vidya/backend/internal/pkg/audit"
	"github.com/vidya/backend/internal/pkg/jwtutil"
)

// Deps carries all shared dependencies into module registration.
type Deps struct {
	Cfg    *config.Config
	DB     *gorm.DB
	RDB    *redis.Client
	Bus    *events.Bus
	Audit  *audit.Writer
	Issuer *jwtutil.Issuer
	Log    *slog.Logger
}
