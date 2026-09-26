// Package stubs registers routes for modules that are designed and schema'd
// but not yet implemented in the scaffold (homework, exams, payroll). They
// return 501 so clients fail loudly rather than silently.
package stubs

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vidya/backend/internal/deps"
	"github.com/vidya/backend/internal/pkg/httpx"
)

// Register wires the not-implemented route groups.
func Register(rg *gin.RouterGroup, _ *deps.Deps) {
	for _, prefix := range []string{"/homework", "/assignments", "/exams", "/payroll"} {
		grp := rg.Group(prefix)
		grp.Any("/*any", notImplemented)
	}
}

func notImplemented(c *gin.Context) {
	httpx.WriteError(c, httpx.NewError(http.StatusNotImplemented, "NOT_IMPLEMENTED", "this module is in the roadmap; not available in this scaffold"))
}
