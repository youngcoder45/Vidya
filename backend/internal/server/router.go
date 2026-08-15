package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/schoolos/backend/internal/deps"
	"github.com/schoolos/backend/internal/modules/announcements"
	"github.com/schoolos/backend/internal/modules/attendance"
	"github.com/schoolos/backend/internal/modules/auth"
	"github.com/schoolos/backend/internal/modules/dashboard"
	"github.com/schoolos/backend/internal/modules/fees"
	"github.com/schoolos/backend/internal/modules/notifications"
	"github.com/schoolos/backend/internal/modules/students"
	"github.com/schoolos/backend/internal/modules/stubs"
	"github.com/schoolos/backend/internal/modules/tenant"
	"github.com/schoolos/backend/internal/server/middleware"
)

// newRouter builds the Gin engine with the full middleware chain and module
// routes. Deps are injected per module (composition root).
func newRouter(d *deps.Deps) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.RequestID())
	engine.Use(middleware.Logger(d.Log))
	engine.Use(middleware.CORS(d.Cfg.CORSOrigins))

	// Liveness/readiness/metrics (public, no auth).
	engine.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	engine.GET("/health/ready", func(c *gin.Context) {
		sqlDB, err := d.DB.DB()
		if err != nil || sqlDB.PingContext(c.Request.Context()) != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "db": "down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready", "db": "up"})
	})
	engine.GET("/metrics", func(c *gin.Context) {
		// Scaffold placeholder. Production: promhttp handler (see docs/07).
		c.String(http.StatusOK, "# schoolos metrics endpoint — wire prometheus client_golang in hardening phase\n")
	})

	// API v1.
	v1 := engine.Group("/api/v1")
	rateLimiter := middleware.NewRateLimiter(d.RDB, d.Cfg.RateLimitPerMin, time.Minute, d.Log)
	v1.Use(rateLimiter.Middleware())

	// Public webhooks (signed, no JWT).
	fees.RegisterWebhook(v1, d)

	// Authenticated + tenant-scoped module routes.
	api := v1.Group("", middleware.Auth(d.Issuer), middleware.Tenant())
	auth.Register(api, d)
	tenant.Register(api, d)
	students.Register(api, d)
	attendance.Register(api, d)
	fees.Register(api, d)
	announcements.Register(api, d)
	notifications.Register(api, d)
	dashboard.Register(api, d)
	stubs.Register(api, d)

	// Platform admin (superadmin) routes.
	platform := v1.Group("/platform", middleware.Auth(d.Issuer), middleware.Tenant())
	tenant.RegisterPlatform(platform, d)

	return engine
}
