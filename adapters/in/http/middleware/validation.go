package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidationError represents a validation error response
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Details []ValidationError `json:"details,omitempty"`
}

// ValidateStruct validates a struct and returns formatted error response
func ValidateStruct(obj interface{}) []ValidationError {
	var validationErrors []ValidationError

	err := validate.Struct(obj)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element ValidationError
			element.Field = err.Field()
			element.Message = msgForTag(err.Tag(), err.Param())
			validationErrors = append(validationErrors, element)
		}
	}

	return validationErrors
}

// msgForTag returns a human-readable message for validation tags
func msgForTag(tag string, param string) string {
	switch tag {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Minimum length is " + param
	case "max":
		return "Maximum length is " + param
	case "gte":
		return "Must be greater than or equal to " + param
	case "lte":
		return "Must be less than or equal to " + param
	case "url":
		return "Invalid URL format"
	default:
		return "Invalid value"
	}
}

// BindAndValidate binds JSON request and validates it
func BindAndValidate(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return false
	}

	if validationErrors := ValidateStruct(obj); len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "Request validation failed",
			Details: validationErrors,
		})
		return false
	}

	return true
}
