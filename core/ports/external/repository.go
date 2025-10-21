package external

import (
	"context"

	"github.com/google/uuid"
	"github.com/nicklaros/jalanrusak-be/core/domain/entities"
)

// UserRepository defines the interface for user persistence
type UserRepository interface {
	// Create creates a new user in the database
	Create(ctx context.Context, user *entities.User) error

	// FindByID retrieves a user by ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error)

	// FindByEmail retrieves a user by email
	FindByEmail(ctx context.Context, email string) (*entities.User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *entities.User) error

	// Delete deletes a user by ID
	Delete(ctx context.Context, id uuid.UUID) error

	// ExistsByEmail checks if a user with the given email exists
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

// RefreshTokenRepository defines the interface for refresh token persistence
type RefreshTokenRepository interface {
	// Create creates a new refresh token
	Create(ctx context.Context, token *entities.RefreshToken) error

	// FindByTokenHash retrieves a refresh token by its hash
	FindByTokenHash(ctx context.Context, tokenHash string) (*entities.RefreshToken, error)

	// FindByUserID retrieves all refresh tokens for a user
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.RefreshToken, error)

	// Update updates an existing refresh token
	Update(ctx context.Context, token *entities.RefreshToken) error

	// RevokeByUserID revokes all refresh tokens for a user
	RevokeByUserID(ctx context.Context, userID uuid.UUID) error

	// RevokeByTokenHash revokes a specific refresh token
	RevokeByTokenHash(ctx context.Context, tokenHash string) error

	// DeleteExpired deletes all expired refresh tokens
	DeleteExpired(ctx context.Context) error
}

// PasswordResetTokenRepository defines the interface for password reset token persistence
type PasswordResetTokenRepository interface {
	// Create creates a new password reset token
	Create(ctx context.Context, token *entities.PasswordResetToken) error

	// FindByTokenHash retrieves a password reset token by its hash
	FindByTokenHash(ctx context.Context, tokenHash string) (*entities.PasswordResetToken, error)

	// Update updates an existing password reset token
	Update(ctx context.Context, token *entities.PasswordResetToken) error

	// DeleteByUserID deletes all password reset tokens for a user
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error

	// DeleteExpired deletes all expired password reset tokens
	DeleteExpired(ctx context.Context) error
}

// AuthEventLogRepository defines the interface for auth event log persistence
type AuthEventLogRepository interface {
	// Create creates a new auth event log entry
	Create(ctx context.Context, log *entities.AuthEventLog) error

	// FindByUserID retrieves auth event logs for a user
	FindByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*entities.AuthEventLog, error)

	// FindFailedLoginAttempts retrieves recent failed login attempts by IP or email
	FindFailedLoginAttempts(ctx context.Context, ipAddress string, limit int) ([]*entities.AuthEventLog, error)
}

// DamagedRoadRepository defines the interface for damaged road report persistence
type DamagedRoadRepository interface {
	// Create creates a new damaged road report
	Create(ctx context.Context, road *entities.DamagedRoad) error

	// FindByID retrieves a damaged road report by ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.DamagedRoad, error)

	// FindByAuthor retrieves damaged road reports by author with pagination
	FindByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*entities.DamagedRoad, int, error)

	// List retrieves damaged road reports with filters and pagination
	List(ctx context.Context, filters *entities.DamagedRoadFilters) ([]*entities.DamagedRoad, int, error)

	// UpdateStatus updates the status of a damaged road report
	UpdateStatus(ctx context.Context, id uuid.UUID, status entities.Status) error

	// Update updates an existing damaged road report
	Update(ctx context.Context, road *entities.DamagedRoad) error

	// Delete deletes a damaged road report by ID
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByGeometry finds damaged road reports within a geographic boundary
	FindByGeometry(ctx context.Context, bounds entities.Geometry) ([]*entities.DamagedRoad, error)
}

// BoundaryRepository defines the interface for administrative boundary and centroid data.
// Used for validating that reported coordinates align with the selected subdistrict.
type BoundaryRepository interface {
	// GetCentroid retrieves the geographic centroid for a given subdistrict code.
	// Returns error if subdistrict code is not found in the boundary dataset.
	GetCentroid(subDistrictCode entities.SubDistrictCode) (entities.Point, error)

	// CheckSubDistrictExists verifies if a subdistrict code exists in the official dataset.
	CheckSubDistrictExists(subDistrictCode entities.SubDistrictCode) (bool, error)

	// StoreCentroid stores centroid data for a subdistrict (for data seeding/updates).
	StoreCentroid(subDistrictCode entities.SubDistrictCode, centroid entities.Point) error
}
