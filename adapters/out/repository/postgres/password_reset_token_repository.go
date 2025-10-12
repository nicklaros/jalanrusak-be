package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nicklaros/jalanrusak-be/core/domain/entities"
	"github.com/nicklaros/jalanrusak-be/core/ports/external"
)

// PasswordResetTokenRepository implements the PasswordResetTokenRepository interface using PostgreSQL
type PasswordResetTokenRepository struct {
	db *sql.DB
}

// NewPasswordResetTokenRepository creates a new PostgreSQL PasswordResetTokenRepository
func NewPasswordResetTokenRepository(db *sql.DB) external.PasswordResetTokenRepository {
	return &PasswordResetTokenRepository{
		db: db,
	}
}

// Create creates a new password reset token
func (r *PasswordResetTokenRepository) Create(ctx context.Context, token *entities.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at, used, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.Used,
		token.CreatedAt,
	)
	return err
}

// FindByTokenHash retrieves a password reset token by its hash
func (r *PasswordResetTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*entities.PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, used, created_at
		FROM password_reset_tokens
		WHERE token_hash = $1
	`
	token := &entities.PasswordResetToken{}

	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.Used,
		&token.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return token, nil
}

// Update updates an existing password reset token
func (r *PasswordResetTokenRepository) Update(ctx context.Context, token *entities.PasswordResetToken) error {
	query := `
		UPDATE password_reset_tokens
		SET used = $2
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, token.ID, token.Used)
	return err
}

// DeleteByUserID deletes all password reset tokens for a user
func (r *PasswordResetTokenRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM password_reset_tokens WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// DeleteExpired deletes all expired password reset tokens
func (r *PasswordResetTokenRepository) DeleteExpired(ctx context.Context) error {
	query := `
		DELETE FROM password_reset_tokens
		WHERE expires_at < NOW()
	`
	_, err := r.db.ExecContext(ctx, query)
	return err
}
