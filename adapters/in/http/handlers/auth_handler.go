package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nicklaros/jalanrusak-be/adapters/in/http/dto"
	"github.com/nicklaros/jalanrusak-be/core/domain/errors"
	"github.com/nicklaros/jalanrusak-be/core/ports/usecases"
)

// AuthHandler handles authentication requests (login, logout, refresh)
type AuthHandler struct {
	authService    usecases.AuthService
	userService    usecases.UserService
	accessTokenTTL int // in hours
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService usecases.AuthService, userService usecases.UserService, accessTokenTTL int) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		userService:    userService,
		accessTokenTTL: accessTokenTTL,
	}
}

// Login handles POST /api/v1/auth/login
// @Summary Authenticate user credentials
// @Description Login with email and password to receive access and refresh tokens.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login payload"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

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

	// Call auth service
	accessToken, refreshToken, err := h.authService.Login(c.Request.Context(), req.Email, req.Password, ipAddress, userAgent)
	if err != nil {
		// Handle domain errors
		switch err {
		case errors.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "invalid_credentials",
				Message: "Invalid email or password",
			})
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to login",
			})
		}
		return
	}

	// Get user info
	user, err := h.userService.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve user info",
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    h.accessTokenTTL * 3600, // convert hours to seconds
		User: dto.UserInfo{
			ID:        user.ID.String(),
			Name:      user.Name,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			LastLogin: user.LastLoginAt,
		},
	})
}

// RefreshToken handles POST /api/v1/auth/refresh
// @Summary Refresh access token
// @Description Exchange a valid refresh token for a new access token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token payload"
// @Success 200 {object} dto.RefreshTokenResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest

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

	// Call auth service
	accessToken, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken, ipAddress, userAgent)
	if err != nil {
		// Handle domain errors
		switch err {
		case errors.ErrInvalidToken:
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "invalid_token",
				Message: "Invalid or revoked refresh token",
			})
		case errors.ErrTokenExpired:
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "token_expired",
				Message: "Refresh token has expired",
			})
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to refresh token",
			})
		}
		return
	}

	// Return success response
	c.JSON(http.StatusOK, dto.RefreshTokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   h.accessTokenTTL * 3600, // convert hours to seconds
	})
}

// Logout handles POST /api/v1/auth/logout
// @Summary Logout and revoke tokens
// @Description Revoke the active session and optional refresh token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LogoutRequest false "Optional refresh token to revoke"
// @Success 200 {object} map[string]string
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	// Get refresh token from request body (optional)
	var req dto.LogoutRequest
	_ = c.ShouldBindJSON(&req)

	// Call auth service to revoke token(s)
	if err := h.authService.Logout(c.Request.Context(), userID.(string), req.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to logout",
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}
