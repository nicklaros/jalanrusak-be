package entities

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken represents a refresh token for session management
type RefreshToken struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	TokenHash  string
	ExpiresAt  time.Time
	Revoked    bool
	CreatedAt  time.Time
	LastUsedAt *time.Time
}

// NewRefreshToken creates a new RefreshToken entity
func NewRefreshToken(userID uuid.UUID, tokenHash string, ttlDays int) *RefreshToken {
	now := time.Now()
	return &RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: now.Add(time.Duration(ttlDays) * 24 * time.Hour),
		Revoked:   false,
		CreatedAt: now,
	}
}

// IsExpired checks if the token has expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsValid checks if the token is valid (not expired and not revoked)
func (rt *RefreshToken) IsValid() bool {
	return !rt.IsExpired() && !rt.Revoked
}

// Revoke marks the token as revoked
func (rt *RefreshToken) Revoke() {
	rt.Revoked = true
}

// UpdateLastUsed sets the LastUsedAt timestamp to current time
func (rt *RefreshToken) UpdateLastUsed() {
	now := time.Now()
	rt.LastUsedAt = &now
}

// ValidateTokenHash checks if the token hash is non-empty
func (rt *RefreshToken) ValidateTokenHash() bool {
	return len(rt.TokenHash) > 0
}
