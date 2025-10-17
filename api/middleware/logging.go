package middleware

import (
	"astron-xmod-shim/pkg/log"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logging HTTP request logging middleware based on zap
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start time
		startTime := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// End time and processing duration
		duration := time.Since(startTime)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		// Select log level based on status code
		logger := log.Info
		if statusCode >= 400 && statusCode < 500 {
			logger = log.Warn
		} else if statusCode >= 500 {
			logger = log.Error
		}

		// Log request information
		logger("HTTP request processing completed",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.String("client_ip", clientIP),
			zap.Duration("duration", duration),
		)
	}
}
