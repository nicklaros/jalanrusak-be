package entities

import (
	"time"

	"github.com/google/uuid"
)

// AuthEventLog represents an audit log entry for authentication events
type AuthEventLog struct {
	ID        uuid.UUID
	UserID    *uuid.UUID // Nullable for failed login attempts where user doesn't exist
	EventType string
	IPAddress string
	UserAgent string
	Success   bool
	CreatedAt time.Time
}

// Event type constants
const (
	EventTypeRegistration      = "registration"
	EventTypeLogin             = "login"
	EventTypeLogout            = "logout"
	EventTypePasswordReset     = "password_reset"
	EventTypePasswordChange    = "password_change"
	EventTypeTokenRefresh      = "token_refresh"
	EventTypeEmailVerification = "email_verification"
)

// NewAuthEventLog creates a new AuthEventLog entity
func NewAuthEventLog(userID *uuid.UUID, eventType, ipAddress, userAgent string, success bool) *AuthEventLog {
	return &AuthEventLog{
		ID:        uuid.New(),
		UserID:    userID,
		EventType: eventType,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   success,
		CreatedAt: time.Now(),
	}
}

// ValidateEventType checks if the event type is one of the defined constants
func (ael *AuthEventLog) ValidateEventType() bool {
	validTypes := map[string]bool{
		EventTypeRegistration:      true,
		EventTypeLogin:             true,
		EventTypeLogout:            true,
		EventTypePasswordReset:     true,
		EventTypePasswordChange:    true,
		EventTypeTokenRefresh:      true,
		EventTypeEmailVerification: true,
	}
	return validTypes[ael.EventType]
}

// IsSecurityEvent checks if this is a security-relevant event (failed login, etc.)
func (ael *AuthEventLog) IsSecurityEvent() bool {
	return !ael.Success && (ael.EventType == EventTypeLogin || ael.EventType == EventTypePasswordReset)
}
