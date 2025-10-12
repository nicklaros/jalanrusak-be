package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nicklaros/jalanrusak-be/core/domain/entities"
	"github.com/nicklaros/jalanrusak-be/core/ports/external"
)

// AuthEventLogRepository implements the AuthEventLogRepository interface using PostgreSQL
type AuthEventLogRepository struct {
	db *sql.DB
}

// NewAuthEventLogRepository creates a new PostgreSQL AuthEventLogRepository
func NewAuthEventLogRepository(db *sql.DB) external.AuthEventLogRepository {
	return &AuthEventLogRepository{
		db: db,
	}
}

// Create creates a new auth event log entry
func (r *AuthEventLogRepository) Create(ctx context.Context, log *entities.AuthEventLog) error {
	query := `
		INSERT INTO auth_event_logs (id, user_id, event_type, ip_address, user_agent, success, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		log.ID,
		log.UserID,
		log.EventType,
		log.IPAddress,
		log.UserAgent,
		log.Success,
		log.CreatedAt,
	)
	return err
}

// FindByUserID retrieves auth event logs for a user
func (r *AuthEventLogRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*entities.AuthEventLog, error) {
	query := `
		SELECT id, user_id, event_type, ip_address, user_agent, success, created_at
		FROM auth_event_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*entities.AuthEventLog
	for rows.Next() {
		log := &entities.AuthEventLog{}
		var userIDNull sql.NullString

		err := rows.Scan(
			&log.ID,
			&userIDNull,
			&log.EventType,
			&log.IPAddress,
			&log.UserAgent,
			&log.Success,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if userIDNull.Valid {
			uid, _ := uuid.Parse(userIDNull.String)
			log.UserID = &uid
		}

		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// FindFailedLoginAttempts retrieves recent failed login attempts by IP address
func (r *AuthEventLogRepository) FindFailedLoginAttempts(ctx context.Context, ipAddress string, limit int) ([]*entities.AuthEventLog, error) {
	query := `
		SELECT id, user_id, event_type, ip_address, user_agent, success, created_at
		FROM auth_event_logs
		WHERE ip_address = $1 
		  AND event_type = $2 
		  AND success = false
		ORDER BY created_at DESC
		LIMIT $3
	`
	rows, err := r.db.QueryContext(ctx, query, ipAddress, entities.EventTypeLogin, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*entities.AuthEventLog
	for rows.Next() {
		log := &entities.AuthEventLog{}
		var userIDNull sql.NullString

		err := rows.Scan(
			&log.ID,
			&userIDNull,
			&log.EventType,
			&log.IPAddress,
			&log.UserAgent,
			&log.Success,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if userIDNull.Valid {
			uid, _ := uuid.Parse(userIDNull.String)
			log.UserID = &uid
		}

		logs = append(logs, log)
	}

	return logs, rows.Err()
}
