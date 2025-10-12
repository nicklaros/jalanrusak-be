package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nicklaros/jalanrusak-be/core/domain/entities"
	"github.com/nicklaros/jalanrusak-be/core/ports/external"
)

// RefreshTokenRepository implements the RefreshTokenRepository interface using PostgreSQL
type RefreshTokenRepository struct {
	db *sql.DB
}

// NewRefreshTokenRepository creates a new PostgreSQL RefreshTokenRepository
func NewRefreshTokenRepository(db *sql.DB) external.RefreshTokenRepository {
	return &RefreshTokenRepository{
		db: db,
	}
}

// Create creates a new refresh token
func (r *RefreshTokenRepository) Create(ctx context.Context, token *entities.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, revoked, created_at, last_used_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.Revoked,
		token.CreatedAt,
		token.LastUsedAt,
	)
	return err
}

// FindByTokenHash retrieves a refresh token by its hash
func (r *RefreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*entities.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, revoked, created_at, last_used_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`
	token := &entities.RefreshToken{}
	var lastUsedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.Revoked,
		&token.CreatedAt,
		&lastUsedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if lastUsedAt.Valid {
		token.LastUsedAt = &lastUsedAt.Time
	}

	return token, nil
}

// FindByUserID retrieves all refresh tokens for a user
func (r *RefreshTokenRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, revoked, created_at, last_used_at
		FROM refresh_tokens
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*entities.RefreshToken
	for rows.Next() {
		token := &entities.RefreshToken{}
		var lastUsedAt sql.NullTime

		err := rows.Scan(
			&token.ID,
			&token.UserID,
			&token.TokenHash,
			&token.ExpiresAt,
			&token.Revoked,
			&token.CreatedAt,
			&lastUsedAt,
		)
		if err != nil {
			return nil, err
		}

		if lastUsedAt.Valid {
			token.LastUsedAt = &lastUsedAt.Time
		}

		tokens = append(tokens, token)
	}

	return tokens, rows.Err()
}

// Update updates an existing refresh token
func (r *RefreshTokenRepository) Update(ctx context.Context, token *entities.RefreshToken) error {
	query := `
		UPDATE refresh_tokens
		SET revoked = $2, last_used_at = $3
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query,
		token.ID,
		token.Revoked,
		token.LastUsedAt,
	)
	return err
}

// RevokeByUserID revokes all refresh tokens for a user
func (r *RefreshTokenRepository) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens
		SET revoked = true
		WHERE user_id = $1 AND revoked = false
	`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// RevokeByTokenHash revokes a specific refresh token
func (r *RefreshTokenRepository) RevokeByTokenHash(ctx context.Context, tokenHash string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked = true
		WHERE token_hash = $1
	`
	_, err := r.db.ExecContext(ctx, query, tokenHash)
	return err
}

// DeleteExpired deletes all expired refresh tokens
func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < NOW()
	`
	_, err := r.db.ExecContext(ctx, query)
	return err
}
