package config

import (
	"errors"
	"fmt"

	"github.com/JokerTrickster/joker_backend/shared/utils"
	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	CORS     CORSConfig
	LogLevel string
	Env      string
	IsLocal  bool
}

type CORSConfig struct {
	AllowedOrigins string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

func Load() (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()

	env := utils.GetEnv("ENV", "development")

	// Set secure CORS defaults for production
	defaultCORSOrigins := "http://localhost:3000,http://localhost:3001"
	if env == "production" {
		defaultCORSOrigins = "" // Force explicit configuration in production
	}

	cfg := &Config{
		Database: DatabaseConfig{
			Host:     utils.GetEnv("DB_HOST", "localhost"),
			Port:     utils.GetEnv("DB_PORT", "3306"),
			User:     utils.GetEnv("DB_USER", "root"),
			Password: utils.GetEnv("DB_PASSWORD", ""),
			Database: utils.GetEnv("DB_NAME", "backend_dev"),
		},
		CORS: CORSConfig{
			AllowedOrigins: utils.GetEnv("CORS_ALLOWED_ORIGINS", defaultCORSOrigins),
		},
		LogLevel: utils.GetEnv("LOG_LEVEL", "info"),
		Env:      env,
		IsLocal:  utils.GetEnv("IS_LOCAL", "false") == "true",
	}

	// Validate critical configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Validate checks that critical configuration values are set correctly
func (c *Config) Validate() error {
	var errs []error

	// Validate database configuration
	if c.Database.Host == "" {
		errs = append(errs, errors.New("DB_HOST is required"))
	}
	if c.Database.Port == "" {
		errs = append(errs, errors.New("DB_PORT is required"))
	}
	if c.Database.User == "" {
		errs = append(errs, errors.New("DB_USER is required"))
	}
	if c.Database.Database == "" {
		errs = append(errs, errors.New("DB_NAME is required"))
	}

	// Production-specific validations
	if c.Env == "production" {
		if c.Database.Password == "" {
			errs = append(errs, errors.New("DB_PASSWORD is required in production"))
		}
		if c.CORS.AllowedOrigins == "" {
			errs = append(errs, errors.New("CORS_ALLOWED_ORIGINS must be explicitly set in production"))
		}
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.LogLevel] {
		errs = append(errs, fmt.Errorf("invalid LOG_LEVEL: %s (must be debug, info, warn, or error)", c.LogLevel))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
