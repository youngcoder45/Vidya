package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-ID"

// RequestID reads or generates a request ID and exposes it in the context
// and response header for end-to-end correlation.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(RequestIDHeader)
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Set(RequestIDHeader, rid)
		c.Header(RequestIDHeader, rid)
		c.Next()
	}
}
