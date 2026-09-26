// Package ctxuser carries the authenticated user's JWT claims through the
// request context.
package ctxuser

import (
	"context"

	"github.com/google/uuid"

	"github.com/schoolos/backend/internal/pkg/jwtutil"
)

type ctxKey struct{}

// WithClaims attaches the verified JWT claims to the context.
func WithClaims(ctx context.Context, c *jwtutil.Claims) context.Context {
	return context.WithValue(ctx, ctxKey{}, c)
}

// From returns the claims attached by the auth middleware.
func From(ctx context.Context) (*jwtutil.Claims, bool) {
	c, ok := ctx.Value(ctxKey{}).(*jwtutil.Claims)
	return c, ok
}

// MustUserID returns the authenticated user ID or the zero UUID.
func MustUserID(ctx context.Context) uuid.UUID {
	if c, ok := From(ctx); ok && c != nil {
		return c.UserID
	}
	return uuid.Nil
}

// MustSchoolID returns the tenant school ID or the zero UUID.
func MustSchoolID(ctx context.Context) uuid.UUID {
	if c, ok := From(ctx); ok && c != nil {
		return c.SchoolID
	}
	return uuid.Nil
}

// HasPermission checks a permission against the claims, honoring module
// wildcards like "students.*".
func HasPermission(ctx context.Context, required string) bool {
	c, ok := From(ctx)
	if !ok || c == nil {
		return false
	}
	for _, p := range c.Permissions {
		if p == "*" || p == required {
			return true
		}
		if len(p) > 2 && p[len(p)-2:] == ".*" && len(required) > len(p)-1 && required[:len(p)-1] == p[:len(p)-1] {
			return true
		}
	}
	return false
}

// HasRole checks whether the user holds a role.
func HasRole(ctx context.Context, role string) bool {
	c, ok := From(ctx)
	if !ok || c == nil {
		return false
	}
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}
