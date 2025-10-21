package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Email    EmailConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
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

type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type EmailConfig struct {
	ServiceType string
	SMTPHost    string
	SMTPPort    int
	SMTPUser    string
	SMTPPass    string
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("ACCESS_TOKEN_TTL_HOURS", 24)
	viper.SetDefault("REFRESH_TOKEN_TTL_DAYS", 30)
	viper.SetDefault("EMAIL_SERVICE_TYPE", "console")
	viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("DB_SSL_MODE", "disable")
	viper.SetDefault("DB_MAX_OPEN_CONNS", 25)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 5)
	viper.SetDefault("DB_CONN_MAX_LIFETIME_MINUTES", 5)

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		// Config file not found is acceptable, we'll use env vars and defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	config := &Config{
		Server: ServerConfig{
			Port: viper.GetString("SERVER_PORT"),
		},
		Database: DatabaseConfig{
			Host:            viper.GetString("DB_HOST"),
			Port:            viper.GetInt("DB_PORT"),
			User:            viper.GetString("DB_USER"),
			Password:        viper.GetString("DB_PASSWORD"),
			DBName:          viper.GetString("DB_NAME"),
			SSLMode:         viper.GetString("DB_SSL_MODE"),
			MaxOpenConns:    viper.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:    viper.GetInt("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetime: time.Duration(viper.GetInt("DB_CONN_MAX_LIFETIME_MINUTES")) * time.Minute,
		},
		JWT: JWTConfig{
			Secret:          viper.GetString("JWT_SECRET"),
			AccessTokenTTL:  time.Duration(viper.GetInt("ACCESS_TOKEN_TTL_HOURS")) * time.Hour,
			RefreshTokenTTL: time.Duration(viper.GetInt("REFRESH_TOKEN_TTL_DAYS")) * 24 * time.Hour,
		},
		Email: EmailConfig{
			ServiceType: viper.GetString("EMAIL_SERVICE_TYPE"),
			SMTPHost:    viper.GetString("SMTP_HOST"),
			SMTPPort:    viper.GetInt("SMTP_PORT"),
			SMTPUser:    viper.GetString("SMTP_USER"),
			SMTPPass:    viper.GetString("SMTP_PASS"),
		},
	}

	// Validate required fields
	if config.Database.Host == "" || config.Database.User == "" || config.Database.DBName == "" {
		return nil, fmt.Errorf("DB_HOST, DB_USER, and DB_NAME are required")
	}
	if config.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return config, nil
}
