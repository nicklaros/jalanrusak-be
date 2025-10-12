package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nicklaros/jalanrusak-be/adapters/in/http/dto"
	"github.com/nicklaros/jalanrusak-be/core/ports/usecases"
)

// AuthMiddleware creates a middleware for JWT authentication
func AuthMiddleware(authService usecases.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "missing_token",
				Message: "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Check Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "invalid_token_format",
				Message: "Authorization header must be in format: Bearer <token>",
			})
			c.Abort()
			return
		}

		accessToken := parts[1]

		// Verify access token
		userID, err := authService.VerifyAccessToken(c.Request.Context(), accessToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "invalid_token",
				Message: "Invalid or expired access token",
			})
			c.Abort()
			return
		}

		// Set user ID in context for handlers to use
		c.Set("userID", userID)

		// Continue to next handler
		c.Next()
	}
}
