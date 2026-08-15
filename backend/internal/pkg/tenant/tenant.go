// Package tenant carries the resolved tenant (school) scope through a request.
// Repositories MUST be constructed with a TenantContext and always filter by it.
package tenant

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// ErrNoTenant is returned when a request has no tenant scope.
var ErrNoTenant = errors.New("tenant: no tenant scope on request")

// Context carries the school scope for the current request.
type Context struct {
	SchoolID uuid.UUID
	PlanCode string
}

type ctxKey struct{}

// WithTenant attaches a tenant scope to the context.
func WithTenant(ctx context.Context, tc Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, tc)
}

// From returns the tenant scope from the context.
func From(ctx context.Context) (Context, bool) {
	tc, ok := ctx.Value(ctxKey{}).(Context)
	return tc, ok
}

// MustFrom returns the tenant scope or panics. Use in request-scoped code
// where the auth middleware has already guaranteed a tenant.
func MustFrom(ctx context.Context) Context {
	tc, ok := From(ctx)
	if !ok {
		panic(ErrNoTenant)
	}
	return tc
}
