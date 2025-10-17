package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luminosita/change-me/pkg/logger"
)

// Logger returns a middleware that logs HTTP requests using structured logging.
func Logger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Log request details after request completes
		duration := time.Since(start)
		log.Infow("http_request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", duration.Milliseconds(),
			"ip", c.ClientIP(),
		)
	}
}
