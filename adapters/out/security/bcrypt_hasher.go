package security

import (
	"context"

	"github.com/nicklaros/jalanrusak-be/core/ports/external"
	"golang.org/x/crypto/bcrypt"
)

// BcryptHasher implements the PasswordHasher interface using bcrypt
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher creates a new BcryptHasher with the specified cost
// Default cost is bcrypt.DefaultCost (10)
func NewBcryptHasher(cost int) external.PasswordHasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	return &BcryptHasher{
		cost: cost,
	}
}

// Hash creates a bcrypt hash from a plain text password
func (h *BcryptHasher) Hash(ctx context.Context, password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// Compare compares a plain text password with a bcrypt hash
func (h *BcryptHasher) Compare(ctx context.Context, hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
