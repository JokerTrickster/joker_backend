package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	CORS     CORSConfig
	LogLevel string
	Env      string
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

	env := getEnv("ENV", "development")

	// Set secure CORS defaults for production
	defaultCORSOrigins := "http://localhost:3000,http://localhost:3001"
	if env == "production" {
		defaultCORSOrigins = "" // Force explicit configuration in production
	}

	cfg := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Database: getEnv("DB_NAME", "backend_dev"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", defaultCORSOrigins),
		},
		LogLevel: getEnv("LOG_LEVEL", "info"),
		Env:      env,
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
