package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/schoolos/backend/internal/pkg/ctxuser"
	"github.com/schoolos/backend/internal/pkg/tenant"
)

// Logger emits one structured JSON log line per request with correlation
// fields (request_id, school_id, user_id) and latency.
func Logger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		rid := c.GetString(RequestIDHeader)
		path := c.Request.URL.Path

		c.Next()

		attrs := []any{
			"request_id", rid,
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		}
		if tc, ok := tenant.From(c.Request.Context()); ok {
			attrs = append(attrs, "school_id", tc.SchoolID)
		}
		if uid := ctxuser.MustUserID(c.Request.Context()); uid != uuid.Nil {
			attrs = append(attrs, "user_id", uid)
		}
		log.Info("http_request", attrs...)
	}
}
