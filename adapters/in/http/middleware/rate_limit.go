package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// RateLimitMiddleware creates a rate limiting middleware with the specified rate
// Rate format: "requests-per-period" (e.g., "10-M" = 10 per minute, "100-H" = 100 per hour)
func RateLimitMiddleware(rate limiter.Rate) gin.HandlerFunc {
	// Create in-memory store
	store := memory.NewStore()

	// Create rate limiter instance
	instance := limiter.New(store, rate)

	// Create Gin middleware
	middleware := mgin.NewMiddleware(instance)

	// Wrap with custom error handling
	return func(c *gin.Context) {
		// Get limiter context
		limiterCtx, err := instance.Get(c.Request.Context(), c.ClientIP())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Rate limiter error",
				"message": "Failed to check rate limit",
			})
			c.Abort()
			return
		}

		// Check if limit exceeded
		if limiterCtx.Reached {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":               "Rate limit exceeded",
				"message":             "Too many requests. Please try again later.",
				"retry_after_seconds": limiterCtx.Reset,
			})
			c.Abort()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limiterCtx.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limiterCtx.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Unix(limiterCtx.Reset, 0).Unix()))

		middleware(c)
	}
}
