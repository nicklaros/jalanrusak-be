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

// PasswordServiceImpl implements the PasswordService use case
type PasswordServiceImpl struct {
	userRepo               external.UserRepository
	passwordResetTokenRepo external.PasswordResetTokenRepository
	passwordHasher         external.PasswordHasher
	tokenGenerator         external.TokenGenerator
	emailService           external.EmailService
	eventLogRepo           external.AuthEventLogRepository
}

// NewPasswordService creates a new PasswordService instance
func NewPasswordService(
	userRepo external.UserRepository,
	passwordResetTokenRepo external.PasswordResetTokenRepository,
	passwordHasher external.PasswordHasher,
	tokenGenerator external.TokenGenerator,
	emailService external.EmailService,
	eventLogRepo external.AuthEventLogRepository,
) usecases.PasswordService {
	return &PasswordServiceImpl{
		userRepo:               userRepo,
		passwordResetTokenRepo: passwordResetTokenRepo,
		passwordHasher:         passwordHasher,
		tokenGenerator:         tokenGenerator,
		emailService:           emailService,
		eventLogRepo:           eventLogRepo,
	}
}

// RequestPasswordReset creates a password reset token and sends reset email
func (s *PasswordServiceImpl) RequestPasswordReset(ctx context.Context, email, ipAddress, userAgent string) error {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Don't reveal if user exists or not (security best practice)
	// Always return success even if user doesn't exist
	if user == nil {
		// Log failed attempt but return success
		s.logAuthEvent(ctx, nil, entities.EventTypePasswordReset, ipAddress, userAgent, false)
		return nil
	}

	// Delete any existing password reset tokens for this user
	if err := s.passwordResetTokenRepo.DeleteByUserID(ctx, user.ID); err != nil {
		// Log but don't fail
		fmt.Printf("Warning: failed to delete old reset tokens: %v\n", err)
	}

	// Generate reset token
	resetToken, err := s.tokenGenerator.GenerateRefreshToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Hash token for storage
	tokenHash, err := s.tokenGenerator.HashToken(ctx, resetToken)
	if err != nil {
		return fmt.Errorf("failed to hash token: %w", err)
	}

	// Create password reset token entity (1 hour expiration)
	tokenEntity := entities.NewPasswordResetToken(user.ID, tokenHash)

	// Save to repository
	if err := s.passwordResetTokenRepo.Create(ctx, tokenEntity); err != nil {
		s.logAuthEvent(ctx, &user.ID, entities.EventTypePasswordReset, ipAddress, userAgent, false)
		return fmt.Errorf("failed to save reset token: %w", err)
	}

	// Send reset email with the unhashed token
	if err := s.emailService.SendPasswordResetEmail(ctx, user.Email, user.Name, resetToken); err != nil {
		s.logAuthEvent(ctx, &user.ID, entities.EventTypePasswordReset, ipAddress, userAgent, false)
		return fmt.Errorf("failed to send reset email: %w", err)
	}

	// Log successful password reset request
	s.logAuthEvent(ctx, &user.ID, entities.EventTypePasswordReset, ipAddress, userAgent, true)

	return nil
}

// ResetPassword resets a user's password using a valid reset token
func (s *PasswordServiceImpl) ResetPassword(ctx context.Context, token, newPassword, ipAddress, userAgent string) error {
	// Validate new password strength
	if !entities.ValidatePasswordStrength(newPassword) {
		return errors.ErrWeakPassword
	}

	// Hash the provided token
	tokenHash, err := s.tokenGenerator.HashToken(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to hash token: %w", err)
	}

	// Find reset token in repository
	tokenEntity, err := s.passwordResetTokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to find reset token: %w", err)
	}
	if tokenEntity == nil {
		return errors.ErrInvalidToken
	}

	// Validate token
	if !tokenEntity.IsValid() {
		s.logAuthEvent(ctx, &tokenEntity.UserID, entities.EventTypePasswordReset, ipAddress, userAgent, false)
		if tokenEntity.IsExpired() {
			return errors.ErrTokenExpired
		}
		return errors.ErrInvalidToken
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, tokenEntity.UserID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return errors.ErrUserNotFound
	}

	// Hash new password
	hashedPassword, err := s.passwordHasher.Hash(ctx, newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user's password
	user.UpdatePassword(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logAuthEvent(ctx, &user.ID, entities.EventTypePasswordReset, ipAddress, userAgent, false)
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used
	tokenEntity.MarkAsUsed()
	if err := s.passwordResetTokenRepo.Update(ctx, tokenEntity); err != nil {
		// Log but don't fail
		fmt.Printf("Warning: failed to mark reset token as used: %v\n", err)
	}

	// Send password changed notification email
	if err := s.emailService.SendPasswordChangedEmail(ctx, user.Email, user.Name); err != nil {
		// Log but don't fail
		fmt.Printf("Warning: failed to send password changed email: %v\n", err)
	}

	// Log successful password reset
	s.logAuthEvent(ctx, &user.ID, entities.EventTypePasswordReset, ipAddress, userAgent, true)

	return nil
}

// ChangePassword changes a user's password (requires current password)
func (s *PasswordServiceImpl) ChangePassword(ctx context.Context, userID, currentPassword, newPassword, ipAddress, userAgent string) error {
	// Validate new password strength
	if !entities.ValidatePasswordStrength(newPassword) {
		return errors.ErrWeakPassword
	}

	// Parse user ID
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, uid)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return errors.ErrUserNotFound
	}

	// Verify current password
	if err := s.passwordHasher.Compare(ctx, user.PasswordHash, currentPassword); err != nil {
		s.logAuthEvent(ctx, &user.ID, entities.EventTypePasswordChange, ipAddress, userAgent, false)
		return errors.ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := s.passwordHasher.Hash(ctx, newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user's password
	user.UpdatePassword(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logAuthEvent(ctx, &user.ID, entities.EventTypePasswordChange, ipAddress, userAgent, false)
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Send password changed notification email
	if err := s.emailService.SendPasswordChangedEmail(ctx, user.Email, user.Name); err != nil {
		// Log but don't fail
		fmt.Printf("Warning: failed to send password changed email: %v\n", err)
	}

	// Log successful password change
	s.logAuthEvent(ctx, &user.ID, entities.EventTypePasswordChange, ipAddress, userAgent, true)

	return nil
}

// logAuthEvent is a helper to log authentication events
func (s *PasswordServiceImpl) logAuthEvent(ctx context.Context, userID *uuid.UUID, eventType, ipAddress, userAgent string, success bool) {
	log := entities.NewAuthEventLog(userID, eventType, ipAddress, userAgent, success)
	// Ignore errors in logging to not fail the main operation
	_ = s.eventLogRepo.Create(ctx, log)
}
