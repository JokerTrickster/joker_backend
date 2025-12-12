package config

import (
	"os"
	"strings"
	"testing"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name          string
		env           map[string]string
		expectError   bool
		errorContains string
	}{
		{
			name: "valid development config",
			env: map[string]string{
				"ENV":     "development",
				"DB_HOST": "localhost",
				"DB_PORT": "3306",
				"DB_USER": "root",
				"DB_NAME": "test_db",
			},
			expectError: false,
		},
		{
			name: "valid production config",
			env: map[string]string{
				"ENV":                  "production",
				"DB_HOST":              "prod-db.example.com",
				"DB_PORT":              "3306",
				"DB_USER":              "prod_user",
				"DB_PASSWORD":          "secure_password",
				"DB_NAME":              "prod_db",
				"CORS_ALLOWED_ORIGINS": "https://example.com",
			},
			expectError: false,
		},
		{
			name: "development with defaults is valid",
			env: map[string]string{
				"ENV": "development",
			},
			expectError: false,
		},
		{
			name: "production missing password",
			env: map[string]string{
				"ENV":                  "production",
				"DB_HOST":              "prod-db.example.com",
				"DB_PORT":              "3306",
				"DB_USER":              "prod_user",
				"DB_NAME":              "prod_db",
				"CORS_ALLOWED_ORIGINS": "https://example.com",
			},
			expectError:   true,
			errorContains: "DB_PASSWORD is required in production",
		},
		{
			name: "production missing CORS",
			env: map[string]string{
				"ENV":         "production",
				"DB_HOST":     "prod-db.example.com",
				"DB_PORT":     "3306",
				"DB_USER":     "prod_user",
				"DB_PASSWORD": "secure_password",
				"DB_NAME":     "prod_db",
			},
			expectError:   true,
			errorContains: "CORS_ALLOWED_ORIGINS must be explicitly set in production",
		},
		{
			name: "invalid log level",
			env: map[string]string{
				"ENV":       "development",
				"DB_HOST":   "localhost",
				"DB_PORT":   "3306",
				"DB_USER":   "root",
				"DB_NAME":   "test_db",
				"LOG_LEVEL": "invalid",
			},
			expectError:   true,
			errorContains: "invalid LOG_LEVEL",
		},
		{
			name: "valid log levels",
			env: map[string]string{
				"ENV":       "development",
				"DB_HOST":   "localhost",
				"DB_PORT":   "3306",
				"DB_USER":   "root",
				"DB_NAME":   "test_db",
				"LOG_LEVEL": "debug",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all environment variables
			os.Clearenv()

			// Set test environment variables
			for k, v := range tt.env {
				os.Setenv(k, v)
			}

			// Load configuration
			cfg, err := Load()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if cfg == nil {
					t.Errorf("Expected config but got nil")
				}
			}

			// Clean up
			os.Clearenv()
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	// Clear all environment variables
	os.Clearenv()

	// Set only required variables
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "3306")
	os.Setenv("DB_USER", "root")
	os.Setenv("DB_NAME", "test_db")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check defaults
	if cfg.Env != "development" {
		t.Errorf("Expected default ENV=development, got: %s", cfg.Env)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("Expected default LOG_LEVEL=info, got: %s", cfg.LogLevel)
	}
	if cfg.IsLocal != false {
		t.Errorf("Expected default IS_LOCAL=false, got: %v", cfg.IsLocal)
	}
	if cfg.CORS.AllowedOrigins != "http://localhost:3000,http://localhost:3001" {
		t.Errorf("Expected default CORS origins, got: %s", cfg.CORS.AllowedOrigins)
	}

	os.Clearenv()
}

func TestConfigProductionDefaults(t *testing.T) {
	// Clear all environment variables
	os.Clearenv()

	// Set required variables for production
	os.Setenv("ENV", "production")
	os.Setenv("DB_HOST", "prod-db.example.com")
	os.Setenv("DB_PORT", "3306")
	os.Setenv("DB_USER", "prod_user")
	os.Setenv("DB_PASSWORD", "secure_password")
	os.Setenv("DB_NAME", "prod_db")
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// In production, CORS should be explicitly set (not default)
	if cfg.CORS.AllowedOrigins != "https://example.com" {
		t.Errorf("Expected explicit CORS origins in production, got: %s", cfg.CORS.AllowedOrigins)
	}

	os.Clearenv()
}
