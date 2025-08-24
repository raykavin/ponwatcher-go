package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
)

// Logger interface that matches your existing log.Smart interface
type Logger interface {
	API(method, path string, statusCode int, duration time.Duration)
}

// RequestLog creates a Gin middleware for request logging
func RequestLog(log Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Process request
		c.Next()
		
		// Calculate duration
		duration := time.Since(start)
		
		// Log the request
		log.API(c.Request.Method, c.Request.URL.Path, c.Writer.Status(), duration)
	}
}