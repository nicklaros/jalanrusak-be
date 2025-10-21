package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nicklaros/jalanrusak-be/pkg/logger"
)

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate or use existing request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store in context
		c.Set(string(logger.RequestIDKey), requestID)

		// Add to response headers
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// RequestLoggingMiddleware logs HTTP requests with structured logging
func RequestLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get request info
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Get error if any
		errorMsg := ""
		if len(c.Errors) > 0 {
			errorMsg = c.Errors.String()
		}

		// Log with structured fields
		logData := map[string]interface{}{
			"method":     method,
			"path":       path,
			"status":     statusCode,
			"latency_ms": latency.Milliseconds(),
			"client_ip":  clientIP,
			"user_agent": userAgent,
		}

		if errorMsg != "" {
			logData["errors"] = errorMsg
		}

		// Log based on status code
		if statusCode >= 500 {
			logger.ErrorContext(c.Request.Context(), "HTTP request failed", logData)
		} else if statusCode >= 400 {
			logger.WarnContext(c.Request.Context(), "HTTP request client error", logData)
		} else {
			logger.InfoContext(c.Request.Context(), "HTTP request completed", logData)
		}
	}
}
