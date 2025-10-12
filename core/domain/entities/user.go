package entities

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID
	Name         string
	Email        string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLoginAt  *time.Time
}

// NewUser creates a new User entity with generated UUID and timestamps
func NewUser(name, email, passwordHash string) *User {
	now := time.Now()
	return &User{
		ID:           uuid.New(),
		Name:         name,
		Email:        strings.ToLower(strings.TrimSpace(email)),
		PasswordHash: passwordHash,
		Role:         "user", // default role
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// ValidateEmail checks if the email format is valid
func (u *User) ValidateEmail() bool {
	if u.Email == "" {
		return false
	}
	// RFC 5322 compliant email regex (simplified)
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(u.Email)
}

// ValidatePasswordStrength checks if a password meets minimum requirements
// Returns true if password is at least 8 characters and contains:
// - At least one uppercase letter
// - At least one lowercase letter
// - At least one digit
func ValidatePasswordStrength(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)

	return hasUpper && hasLower && hasDigit
}

// ValidateName checks if the name is valid (non-empty and reasonable length)
func (u *User) ValidateName() bool {
	name := strings.TrimSpace(u.Name)
	return len(name) > 0 && len(name) <= 100
}

// UpdateLastLogin sets the LastLoginAt timestamp to current time
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}

// UpdatePassword updates the password hash and UpdatedAt timestamp
func (u *User) UpdatePassword(newPasswordHash string) {
	u.PasswordHash = newPasswordHash
	u.UpdatedAt = time.Now()
}

// IsAdmin checks if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}
