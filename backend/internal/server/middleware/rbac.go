package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/schoolos/backend/internal/pkg/ctxuser"
	"github.com/schoolos/backend/internal/pkg/httpx"
)

// RequirePermission denies requests unless the user holds the permission
// (or a module wildcard like "students.*"). Must run after Auth + Tenant.
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ctxuser.HasPermission(c.Request.Context(), permission) {
			httpx.WriteError(c, httpx.ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireRole denies requests unless the user holds one of the roles.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, r := range roles {
			if ctxuser.HasRole(c.Request.Context(), r) {
				c.Next()
				return
			}
		}
		httpx.WriteError(c, httpx.ErrForbidden)
		c.Abort()
	}
}
