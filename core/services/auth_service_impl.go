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

// AuthServiceImpl implements the AuthService use case
type AuthServiceImpl struct {
	userRepo        external.UserRepository
	tokenRepo       external.RefreshTokenRepository
	passwordHasher  external.PasswordHasher
	tokenGenerator  external.TokenGenerator
	eventLogRepo    external.AuthEventLogRepository
	refreshTokenTTL int // TTL in days
}

// NewAuthService creates a new AuthService instance
func NewAuthService(
	userRepo external.UserRepository,
	tokenRepo external.RefreshTokenRepository,
	passwordHasher external.PasswordHasher,
	tokenGenerator external.TokenGenerator,
	eventLogRepo external.AuthEventLogRepository,
	refreshTokenTTL int,
) usecases.AuthService {
	return &AuthServiceImpl{
		userRepo:        userRepo,
		tokenRepo:       tokenRepo,
		passwordHasher:  passwordHasher,
		tokenGenerator:  tokenGenerator,
		eventLogRepo:    eventLogRepo,
		refreshTokenTTL: refreshTokenTTL,
	}
}

// Login authenticates a user with email and password
func (s *AuthServiceImpl) Login(ctx context.Context, email, password, ipAddress, userAgent string) (accessToken, refreshToken string, err error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", "", fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		// Log failed login attempt
		s.logAuthEvent(ctx, nil, entities.EventTypeLogin, ipAddress, userAgent, false)
		return "", "", errors.ErrInvalidCredentials
	}

	// Verify password
	if err := s.passwordHasher.Compare(ctx, user.PasswordHash, password); err != nil {
		// Log failed login attempt
		s.logAuthEvent(ctx, &user.ID, entities.EventTypeLogin, ipAddress, userAgent, false)
		return "", "", errors.ErrInvalidCredentials
	}

	// Generate access token
	accessToken, err = s.tokenGenerator.GenerateAccessToken(ctx, user.ID.String())
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshTokenRaw, err := s.tokenGenerator.GenerateRefreshToken(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Hash refresh token for storage
	refreshTokenHash, err := s.tokenGenerator.HashToken(ctx, refreshTokenRaw)
	if err != nil {
		return "", "", fmt.Errorf("failed to hash refresh token: %w", err)
	}

	// Save refresh token to repository
	tokenEntity := entities.NewRefreshToken(user.ID, refreshTokenHash, s.refreshTokenTTL)
	if err := s.tokenRepo.Create(ctx, tokenEntity); err != nil {
		return "", "", fmt.Errorf("failed to save refresh token: %w", err)
	}

	// Update user's last login time
	user.UpdateLastLogin()
	if err := s.userRepo.Update(ctx, user); err != nil {
		// Log error but don't fail the login
		fmt.Printf("Warning: failed to update last login time: %v\n", err)
	}

	// Log successful login
	s.logAuthEvent(ctx, &user.ID, entities.EventTypeLogin, ipAddress, userAgent, true)

	return accessToken, refreshTokenRaw, nil
}

// RefreshToken generates a new access token using a valid refresh token
func (s *AuthServiceImpl) RefreshToken(ctx context.Context, refreshToken, ipAddress, userAgent string) (accessToken string, err error) {
	// Hash the provided refresh token
	tokenHash, err := s.tokenGenerator.HashToken(ctx, refreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to hash token: %w", err)
	}

	// Find refresh token in repository
	tokenEntity, err := s.tokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return "", fmt.Errorf("failed to find refresh token: %w", err)
	}
	if tokenEntity == nil {
		return "", errors.ErrInvalidToken
	}

	// Validate token
	if !tokenEntity.IsValid() {
		s.logAuthEvent(ctx, &tokenEntity.UserID, entities.EventTypeTokenRefresh, ipAddress, userAgent, false)
		if tokenEntity.IsExpired() {
			return "", errors.ErrTokenExpired
		}
		return "", errors.ErrInvalidToken
	}

	// Generate new access token
	accessToken, err = s.tokenGenerator.GenerateAccessToken(ctx, tokenEntity.UserID.String())
	if err != nil {
		return "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// Update last used time
	tokenEntity.UpdateLastUsed()
	if err := s.tokenRepo.Update(ctx, tokenEntity); err != nil {
		// Log error but don't fail the refresh
		fmt.Printf("Warning: failed to update token last used time: %v\n", err)
	}

	// Log successful token refresh
	s.logAuthEvent(ctx, &tokenEntity.UserID, entities.EventTypeTokenRefresh, ipAddress, userAgent, true)

	return accessToken, nil
}

// Logout invalidates the user's refresh token
func (s *AuthServiceImpl) Logout(ctx context.Context, userID string, refreshToken string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// If refresh token provided, revoke specific token
	if refreshToken != "" {
		tokenHash, err := s.tokenGenerator.HashToken(ctx, refreshToken)
		if err != nil {
			return fmt.Errorf("failed to hash token: %w", err)
		}
		if err := s.tokenRepo.RevokeByTokenHash(ctx, tokenHash); err != nil {
			return fmt.Errorf("failed to revoke token: %w", err)
		}
	} else {
		// Otherwise, revoke all user tokens
		if err := s.tokenRepo.RevokeByUserID(ctx, uid); err != nil {
			return fmt.Errorf("failed to revoke user tokens: %w", err)
		}
	}

	// Log logout event
	s.logAuthEvent(ctx, &uid, entities.EventTypeLogout, "", "", true)

	return nil
}

// VerifyAccessToken validates an access token and returns the user ID
func (s *AuthServiceImpl) VerifyAccessToken(ctx context.Context, accessToken string) (userID string, err error) {
	userID, err = s.tokenGenerator.ValidateAccessToken(ctx, accessToken)
	if err != nil {
		return "", errors.ErrInvalidToken
	}
	return userID, nil
}

// logAuthEvent is a helper to log authentication events
func (s *AuthServiceImpl) logAuthEvent(ctx context.Context, userID *uuid.UUID, eventType, ipAddress, userAgent string, success bool) {
	log := entities.NewAuthEventLog(userID, eventType, ipAddress, userAgent, success)
	// Ignore errors in logging to not fail the main operation
	_ = s.eventLogRepo.Create(ctx, log)
}
