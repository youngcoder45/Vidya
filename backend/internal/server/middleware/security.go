package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders sets conservative response headers. TLS-only headers such as
// HSTS are harmless over plain HTTP in development and expected in production.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
	}
}
