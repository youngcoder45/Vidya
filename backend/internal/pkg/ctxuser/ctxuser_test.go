package ctxuser

import (
	"context"
	"testing"

	"github.com/schoolos/backend/internal/pkg/jwtutil"
)

func withPerms(perms ...string) context.Context {
	return WithClaims(context.Background(), &jwtutil.Claims{Permissions: perms})
}

func TestHasPermission(t *testing.T) {
	cases := []struct {
		name     string
		perms    []string
		required string
		want     bool
	}{
		{"exact match", []string{"students.read"}, "students.read", true},
		{"missing", []string{"students.read"}, "fees.read", false},
		{"module wildcard", []string{"students.*"}, "students.read", true},
		{"module wildcard wrong module", []string{"students.*"}, "fees.read", false},
		{"global wildcard grants anything", []string{"*"}, "fees.payment.create", true},
		{"empty", nil, "students.read", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := HasPermission(withPerms(tc.perms...), tc.required)
			if got != tc.want {
				t.Fatalf("HasPermission(%v, %q) = %v, want %v", tc.perms, tc.required, got, tc.want)
			}
		})
	}
}
