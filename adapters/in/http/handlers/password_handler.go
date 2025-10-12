package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nicklaros/jalanrusak-be/adapters/in/http/dto"
	"github.com/nicklaros/jalanrusak-be/core/domain/errors"
	"github.com/nicklaros/jalanrusak-be/core/ports/usecases"
)

// PasswordHandler handles password-related requests (reset, change)
type PasswordHandler struct {
	passwordService usecases.PasswordService
}

// NewPasswordHandler creates a new PasswordHandler
func NewPasswordHandler(passwordService usecases.PasswordService) *PasswordHandler {
	return &PasswordHandler{
		passwordService: passwordService,
	}
}

// RequestPasswordReset handles POST /api/auth/password/reset-request
func (h *PasswordHandler) RequestPasswordReset(c *gin.Context) {
	var req dto.PasswordResetRequestRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// Get client IP and User-Agent
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Call password service
	// Note: Always returns success to prevent email enumeration attacks
	if err := h.passwordService.RequestPasswordReset(c.Request.Context(), req.Email, ipAddress, userAgent); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to process password reset request",
		})
		return
	}

	// Return success response (even if email doesn't exist)
	c.JSON(http.StatusOK, dto.PasswordResetRequestResponse{
		Message: "If an account exists with this email, you will receive a password reset link",
	})
}

// ResetPassword handles POST /api/auth/password/reset-confirm
func (h *PasswordHandler) ResetPassword(c *gin.Context) {
	var req dto.PasswordResetConfirmRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// Get client IP and User-Agent
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Call password service
	if err := h.passwordService.ResetPassword(c.Request.Context(), req.Token, req.NewPassword, ipAddress, userAgent); err != nil {
		// Handle domain errors
		switch err {
		case errors.ErrInvalidToken:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "invalid_token",
				Message: "Invalid or already used reset token",
			})
		case errors.ErrTokenExpired:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "token_expired",
				Message: "Reset token has expired. Please request a new one",
			})
		case errors.ErrWeakPassword:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "weak_password",
				Message: "Password must be at least 8 characters and contain uppercase, lowercase, and digit",
			})
		case errors.ErrUserNotFound:
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "user_not_found",
				Message: "User not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to reset password",
			})
		}
		return
	}

	// Return success response
	c.JSON(http.StatusOK, dto.PasswordResetConfirmResponse{
		Message: "Password has been reset successfully",
	})
}

// ChangePassword handles POST /api/auth/password/change (requires authentication)
func (h *PasswordHandler) ChangePassword(c *gin.Context) {
	var req dto.PasswordChangeRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	// Get client IP and User-Agent
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Call password service
	if err := h.passwordService.ChangePassword(c.Request.Context(), userID.(string), req.CurrentPassword, req.NewPassword, ipAddress, userAgent); err != nil {
		// Handle domain errors
		switch err {
		case errors.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "invalid_password",
				Message: "Current password is incorrect",
			})
		case errors.ErrWeakPassword:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "weak_password",
				Message: "Password must be at least 8 characters and contain uppercase, lowercase, and digit",
			})
		case errors.ErrUserNotFound:
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "user_not_found",
				Message: "User not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to change password",
			})
		}
		return
	}

	// Return success response
	c.JSON(http.StatusOK, dto.PasswordChangeResponse{
		Message: "Password has been changed successfully",
	})
}
