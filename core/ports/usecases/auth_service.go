package usecases

import (
	"context"

	"github.com/nicklaros/jalanrusak-be/core/domain/entities"
)

// AuthService defines the authentication use case interface
type AuthService interface {
	// Login authenticates a user with email and password
	// Returns access token, refresh token, and error
	Login(ctx context.Context, email, password, ipAddress, userAgent string) (accessToken, refreshToken string, err error)

	// RefreshToken generates a new access token using a valid refresh token
	// Returns new access token and error
	RefreshToken(ctx context.Context, refreshToken, ipAddress, userAgent string) (accessToken string, err error)

	// Logout invalidates the user's refresh token
	Logout(ctx context.Context, userID string, refreshToken string) error

	// VerifyAccessToken validates an access token and returns the user ID
	VerifyAccessToken(ctx context.Context, accessToken string) (userID string, err error)
}

// UserService defines the user management use case interface
type UserService interface {
	// Register creates a new user account
	// Returns the created user and error
	Register(ctx context.Context, name, email, password, ipAddress, userAgent string) (*entities.User, error)

	// GetUserByID retrieves a user by their ID
	GetUserByID(ctx context.Context, userID string) (*entities.User, error)

	// GetUserByEmail retrieves a user by their email
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)

	// UpdateUser updates user information
	UpdateUser(ctx context.Context, user *entities.User) error
}

// PasswordService defines the password management use case interface
type PasswordService interface {
	// RequestPasswordReset creates a password reset token and sends reset email
	// Returns error
	RequestPasswordReset(ctx context.Context, email, ipAddress, userAgent string) error

	// ResetPassword resets a user's password using a valid reset token
	// Returns error
	ResetPassword(ctx context.Context, token, newPassword, ipAddress, userAgent string) error

	// ChangePassword changes a user's password (requires current password)
	// Returns error
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword, ipAddress, userAgent string) error
}
