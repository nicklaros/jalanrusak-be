package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nicklaros/jalanrusak-be/adapters/in/http/dto"
	"github.com/nicklaros/jalanrusak-be/core/domain/errors"
	"github.com/nicklaros/jalanrusak-be/core/ports/usecases"
)

// RegistrationHandler handles user registration requests
type RegistrationHandler struct {
	userService usecases.UserService
}

// NewRegistrationHandler creates a new RegistrationHandler
func NewRegistrationHandler(userService usecases.UserService) *RegistrationHandler {
	return &RegistrationHandler{
		userService: userService,
	}
}

// Register handles POST /api/auth/register
// @Summary Register a new user
// @Description Create a new user with name, email, and password.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegistrationRequest true "Registration payload"
// @Success 201 {object} dto.RegistrationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/register [post]
func (h *RegistrationHandler) Register(c *gin.Context) {
	var req dto.RegistrationRequest

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

	// Call user service
	user, err := h.userService.Register(c.Request.Context(), req.Name, req.Email, req.Password, ipAddress, userAgent)
	if err != nil {
		// Handle domain errors
		switch err {
		case errors.ErrInvalidEmail:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "invalid_email",
				Message: "Email format is invalid",
			})
		case errors.ErrWeakPassword:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "weak_password",
				Message: "Password must be at least 8 characters and contain uppercase, lowercase, and digit",
			})
		case errors.ErrUserAlreadyExists:
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error:   "user_already_exists",
				Message: "A user with this email already exists",
			})
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to register user",
			})
		}
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, dto.RegistrationResponse{
		ID:        user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	})
}
