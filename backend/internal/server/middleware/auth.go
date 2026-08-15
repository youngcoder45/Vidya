package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/schoolos/backend/internal/pkg/ctxuser"
	"github.com/schoolos/backend/internal/pkg/httpx"
	"github.com/schoolos/backend/internal/pkg/jwtutil"
)

// Auth verifies the Bearer access token and attaches its claims to the
// request context.
func Auth(issuer *jwtutil.Issuer) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			httpx.WriteError(c, httpx.ErrUnauthenticated)
			c.Abort()
			return
		}
		raw := strings.TrimPrefix(header, "Bearer ")
		claims, err := issuer.Parse(raw)
		if err != nil {
			httpx.WriteError(c, httpx.ErrUnauthenticated)
			c.Abort()
			return
		}
		c.Request = c.Request.WithContext(ctxuser.WithClaims(c.Request.Context(), claims))
		c.Next()
	}
}
