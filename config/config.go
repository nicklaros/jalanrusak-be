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
	URL string
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
			URL: viper.GetString("DATABASE_URL"),
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
	if config.Database.URL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if config.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return config, nil
}
