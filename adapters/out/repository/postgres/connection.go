package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/nicklaros/jalanrusak-be/pkg/logger"
)

// ConnectionConfig holds PostgreSQL connection configuration
type ConnectionConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// NewConnection creates a new PostgreSQL connection pool with PostGIS support
func NewConnection(config ConnectionConfig) (*sqlx.DB, error) {
	// Build connection string
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		config.SSLMode,
	)

	// Open database connection
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable PostGIS extension if not already enabled
	if err := ensurePostGIS(db.DB); err != nil {
		logger.Warn(fmt.Sprintf("PostGIS extension check failed (may already exist): %v", err))
	}

	logger.Info("Database connection established successfully")
	return db, nil
}

// ensurePostGIS ensures PostGIS extension is enabled
func ensurePostGIS(db *sql.DB) error {
	// Try to create PostGIS extension (will fail silently if already exists)
	_, err := db.Exec("CREATE EXTENSION IF NOT EXISTS postgis")
	if err != nil {
		return fmt.Errorf("failed to ensure PostGIS extension: %w", err)
	}

	// Verify PostGIS is available
	var version string
	err = db.QueryRow("SELECT PostGIS_version()").Scan(&version)
	if err != nil {
		return fmt.Errorf("failed to verify PostGIS: %w", err)
	}

	logger.Info(fmt.Sprintf("PostGIS extension verified: %s", version))
	return nil
}

// Close closes the database connection
func Close(db *sqlx.DB) error {
	if db == nil {
		return nil
	}

	logger.Info("Closing database connection")
	return db.Close()
}
