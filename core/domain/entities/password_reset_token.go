package entities

import (
	"time"

	"github.com/google/uuid"
)

// PasswordResetToken represents a temporary token for password reset flow
type PasswordResetToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}

// NewPasswordResetToken creates a new PasswordResetToken entity
// Default TTL is 1 hour
func NewPasswordResetToken(userID uuid.UUID, tokenHash string) *PasswordResetToken {
	now := time.Now()
	return &PasswordResetToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: now.Add(1 * time.Hour), // 1 hour expiration
		Used:      false,
		CreatedAt: now,
	}
}

// IsExpired checks if the token has expired
func (prt *PasswordResetToken) IsExpired() bool {
	return time.Now().After(prt.ExpiresAt)
}

// IsValid checks if the token is valid (not expired and not used)
func (prt *PasswordResetToken) IsValid() bool {
	return !prt.IsExpired() && !prt.Used
}

// MarkAsUsed marks the token as used
func (prt *PasswordResetToken) MarkAsUsed() {
	prt.Used = true
}

// ValidateTokenHash checks if the token hash is non-empty
func (prt *PasswordResetToken) ValidateTokenHash() bool {
	return len(prt.TokenHash) > 0
}
