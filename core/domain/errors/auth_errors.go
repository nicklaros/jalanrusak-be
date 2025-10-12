package errors

import "errors"

// Authentication and authorization errors
var (
	// ErrInvalidCredentials is returned when login credentials are incorrect
	ErrInvalidCredentials = errors.New("invalid email or password")

	// ErrUserAlreadyExists is returned when attempting to register with an existing email
	ErrUserAlreadyExists = errors.New("user with this email already exists")

	// ErrInvalidToken is returned when a token is malformed or invalid
	ErrInvalidToken = errors.New("invalid token")

	// ErrTokenExpired is returned when a token has expired
	ErrTokenExpired = errors.New("token has expired")

	// ErrTokenRevoked is returned when a refresh token has been revoked
	ErrTokenRevoked = errors.New("token has been revoked")

	// ErrWeakPassword is returned when a password doesn't meet strength requirements
	ErrWeakPassword = errors.New("password must be at least 8 characters and contain uppercase, lowercase, and digit")

	// ErrInvalidEmail is returned when email format is invalid
	ErrInvalidEmail = errors.New("invalid email format")

	// ErrUserNotFound is returned when a user is not found in the system
	ErrUserNotFound = errors.New("user not found")

	// ErrInvalidName is returned when name validation fails
	ErrInvalidName = errors.New("invalid name: must be non-empty and max 100 characters")

	// ErrPasswordResetTokenUsed is returned when trying to use an already-used password reset token
	ErrPasswordResetTokenUsed = errors.New("password reset token has already been used")

	// ErrUnauthorized is returned when user lacks permission for an action
	ErrUnauthorized = errors.New("unauthorized access")

	// ErrInvalidTokenHash is returned when token hash is empty or invalid
	ErrInvalidTokenHash = errors.New("invalid token hash")
)
