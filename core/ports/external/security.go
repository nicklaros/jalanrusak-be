package external

import "context"

// TokenGenerator defines the interface for JWT token generation and validation
type TokenGenerator interface {
	// GenerateAccessToken creates a new JWT access token for the given user ID
	GenerateAccessToken(ctx context.Context, userID string) (string, error)

	// GenerateRefreshToken creates a new refresh token
	GenerateRefreshToken(ctx context.Context) (string, error)

	// ValidateAccessToken validates an access token and returns the user ID
	ValidateAccessToken(ctx context.Context, token string) (userID string, err error)

	// HashToken creates a hash of the token for secure storage
	HashToken(ctx context.Context, token string) (string, error)
}

// PasswordHasher defines the interface for password hashing and verification
type PasswordHasher interface {
	// Hash creates a bcrypt hash from a plain text password
	Hash(ctx context.Context, password string) (string, error)

	// Compare compares a plain text password with a hash
	Compare(ctx context.Context, hashedPassword, password string) error
}

// EmailService defines the interface for sending emails
type EmailService interface {
	// SendPasswordResetEmail sends a password reset email with a token
	SendPasswordResetEmail(ctx context.Context, to, name, resetToken string) error

	// SendWelcomeEmail sends a welcome email to a newly registered user
	SendWelcomeEmail(ctx context.Context, to, name string) error

	// SendPasswordChangedEmail sends a notification email after password change
	SendPasswordChangedEmail(ctx context.Context, to, name string) error
}
