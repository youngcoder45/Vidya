package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/vidya/backend/internal/pkg/ctxuser"
	"github.com/vidya/backend/internal/pkg/httpx"
	"github.com/vidya/backend/internal/pkg/tenant"
)

// Tenant resolves the tenant scope from the authenticated user's claims and
// attaches a tenant.Context to the request. Superadmins (platform) may carry
// no school; guarded endpoints that need a school must require one.
func Tenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := ctxuser.From(c.Request.Context())
		if !ok || claims == nil {
			httpx.WriteError(c, httpx.ErrUnauthenticated)
			c.Abort()
			return
		}
		tc := tenant.Context{SchoolID: claims.SchoolID}
		c.Request = c.Request.WithContext(tenant.WithTenant(c.Request.Context(), tc))
		c.Next()
	}
}
