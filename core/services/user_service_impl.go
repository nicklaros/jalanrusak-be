package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nicklaros/jalanrusak-be/core/domain/entities"
	"github.com/nicklaros/jalanrusak-be/core/domain/errors"
	"github.com/nicklaros/jalanrusak-be/core/ports/external"
	"github.com/nicklaros/jalanrusak-be/core/ports/usecases"
)

// UserServiceImpl implements the UserService use case
type UserServiceImpl struct {
	userRepo       external.UserRepository
	passwordHasher external.PasswordHasher
	eventLogRepo   external.AuthEventLogRepository
}

// NewUserService creates a new UserService instance
func NewUserService(
	userRepo external.UserRepository,
	passwordHasher external.PasswordHasher,
	eventLogRepo external.AuthEventLogRepository,
) usecases.UserService {
	return &UserServiceImpl{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
		eventLogRepo:   eventLogRepo,
	}
}

// Register creates a new user account
func (s *UserServiceImpl) Register(ctx context.Context, name, email, password, ipAddress, userAgent string) (*entities.User, error) {
	// Validate email format
	tempUser := entities.NewUser(name, email, "")
	if !tempUser.ValidateEmail() {
		// Log failed registration attempt
		s.logAuthEvent(ctx, nil, entities.EventTypeRegistration, ipAddress, userAgent, false)
		return nil, errors.ErrInvalidEmail
	}

	// Validate name
	if !tempUser.ValidateName() {
		s.logAuthEvent(ctx, nil, entities.EventTypeRegistration, ipAddress, userAgent, false)
		return nil, fmt.Errorf("name is required and must be less than 100 characters")
	}

	// Validate password strength
	if !entities.ValidatePasswordStrength(password) {
		s.logAuthEvent(ctx, nil, entities.EventTypeRegistration, ipAddress, userAgent, false)
		return nil, errors.ErrWeakPassword
	}

	// Check if user already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		s.logAuthEvent(ctx, nil, entities.EventTypeRegistration, ipAddress, userAgent, false)
		return nil, errors.ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := s.passwordHasher.Hash(ctx, password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user entity
	user := entities.NewUser(name, email, hashedPassword)

	// Save to repository
	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logAuthEvent(ctx, nil, entities.EventTypeRegistration, ipAddress, userAgent, false)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Log successful registration
	s.logAuthEvent(ctx, &user.ID, entities.EventTypeRegistration, ipAddress, userAgent, true)

	return user, nil
}

// GetUserByID retrieves a user by their ID
func (s *UserServiceImpl) GetUserByID(ctx context.Context, userID string) (*entities.User, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, errors.ErrUserNotFound
	}

	return user, nil
}

// GetUserByEmail retrieves a user by their email
func (s *UserServiceImpl) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, errors.ErrUserNotFound
	}

	return user, nil
}

// UpdateUser updates user information
func (s *UserServiceImpl) UpdateUser(ctx context.Context, user *entities.User) error {
	// Validate user data
	if !user.ValidateEmail() {
		return errors.ErrInvalidEmail
	}
	if !user.ValidateName() {
		return fmt.Errorf("invalid user name")
	}

	// Update in repository
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// logAuthEvent is a helper to log authentication events
func (s *UserServiceImpl) logAuthEvent(ctx context.Context, userID *uuid.UUID, eventType, ipAddress, userAgent string, success bool) {
	log := entities.NewAuthEventLog(userID, eventType, ipAddress, userAgent, success)
	// Ignore errors in logging to not fail the main operation
	_ = s.eventLogRepo.Create(ctx, log)
}
