package dto

// PasswordResetRequestRequest represents the request to initiate password reset
type PasswordResetRequestRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// PasswordResetRequestResponse represents the response after password reset request
type PasswordResetRequestResponse struct {
	Message string `json:"message"`
}

// PasswordResetConfirmRequest represents the request to confirm password reset with token
type PasswordResetConfirmRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// PasswordResetConfirmResponse represents the response after successful password reset
type PasswordResetConfirmResponse struct {
	Message string `json:"message"`
}

// PasswordChangeRequest represents the request to change password (authenticated)
type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// PasswordChangeResponse represents the response after successful password change
type PasswordChangeResponse struct {
	Message string `json:"message"`
}
