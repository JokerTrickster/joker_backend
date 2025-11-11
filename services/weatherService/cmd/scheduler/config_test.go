package main

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		wantErr  bool
		validate func(*testing.T, *Config)
	}{
		{
			name: "default values",
			envVars: map[string]string{
				"DB_HOST":             "testhost",
				"DB_USER":             "testuser",
				"DB_NAME":             "testdb",
				"REDIS_HOST":          "redishost",
				"FCM_CREDENTIALS_PATH": "/test/path",
			},
			wantErr: false,
			validate: func(t *testing.T, c *Config) {
				if c.DBPort != "3306" {
					t.Errorf("Expected default DB_PORT=3306, got %s", c.DBPort)
				}
				if c.SchedulerInterval != 1*time.Minute {
					t.Errorf("Expected default interval=1m, got %v", c.SchedulerInterval)
				}
			},
		},
		{
			name: "custom values",
			envVars: map[string]string{
				"DB_HOST":             "customhost",
				"DB_PORT":             "3307",
				"DB_USER":             "customuser",
				"DB_PASSWORD":         "custompass",
				"DB_NAME":             "customdb",
				"REDIS_HOST":          "customredis",
				"REDIS_PORT":          "6380",
				"REDIS_PASSWORD":      "redispass",
				"FCM_CREDENTIALS_PATH": "/custom/fcm.json",
				"SCHEDULER_INTERVAL":  "5m",
				"LOG_LEVEL":           "debug",
			},
			wantErr: false,
			validate: func(t *testing.T, c *Config) {
				if c.DBHost != "customhost" {
					t.Errorf("Expected DB_HOST=customhost, got %s", c.DBHost)
				}
				if c.DBPort != "3307" {
					t.Errorf("Expected DB_PORT=3307, got %s", c.DBPort)
				}
				if c.SchedulerInterval != 5*time.Minute {
					t.Errorf("Expected interval=5m, got %v", c.SchedulerInterval)
				}
				if c.LogLevel != "debug" {
					t.Errorf("Expected LOG_LEVEL=debug, got %s", c.LogLevel)
				}
			},
		},
		{
			name: "invalid scheduler interval",
			envVars: map[string]string{
				"DB_HOST":             "testhost",
				"DB_USER":             "testuser",
				"DB_NAME":             "testdb",
				"REDIS_HOST":          "redishost",
				"FCM_CREDENTIALS_PATH": "/test/path",
				"SCHEDULER_INTERVAL":  "invalid",
			},
			wantErr: false,
			validate: func(t *testing.T, c *Config) {
				// Should fall back to default 1 minute
				if c.SchedulerInterval != 1*time.Minute {
					t.Errorf("Expected fallback to 1m, got %v", c.SchedulerInterval)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer func() {
				// Clean up
				for k := range tt.envVars {
					os.Unsetenv(k)
				}
			}()

			config := loadConfig()

			if tt.validate != nil {
				tt.validate(t, config)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				DBHost:             "localhost",
				DBUser:             "root",
				DBName:             "joker",
				RedisHost:          "localhost",
				FCMCredentialsPath: "/path/to/fcm.json",
			},
			wantErr: false,
		},
		{
			name: "missing DB_HOST",
			config: &Config{
				DBHost:             "",
				DBUser:             "root",
				DBName:             "joker",
				RedisHost:          "localhost",
				FCMCredentialsPath: "/path/to/fcm.json",
			},
			wantErr: true,
			errMsg:  "DB_HOST is required",
		},
		{
			name: "missing DB_USER",
			config: &Config{
				DBHost:             "localhost",
				DBUser:             "",
				DBName:             "joker",
				RedisHost:          "localhost",
				FCMCredentialsPath: "/path/to/fcm.json",
			},
			wantErr: true,
			errMsg:  "DB_USER is required",
		},
		{
			name: "missing DB_NAME",
			config: &Config{
				DBHost:             "localhost",
				DBUser:             "root",
				DBName:             "",
				RedisHost:          "localhost",
				FCMCredentialsPath: "/path/to/fcm.json",
			},
			wantErr: true,
			errMsg:  "DB_NAME is required",
		},
		{
			name: "missing REDIS_HOST",
			config: &Config{
				DBHost:             "localhost",
				DBUser:             "root",
				DBName:             "joker",
				RedisHost:          "",
				FCMCredentialsPath: "/path/to/fcm.json",
			},
			wantErr: true,
			errMsg:  "REDIS_HOST is required",
		},
		{
			name: "missing FCM_CREDENTIALS_PATH",
			config: &Config{
				DBHost:             "localhost",
				DBUser:             "root",
				DBName:             "joker",
				RedisHost:          "localhost",
				FCMCredentialsPath: "",
			},
			wantErr: true,
			errMsg:  "FCM_CREDENTIALS_PATH is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("Expected error message '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "env var set",
			key:          "TEST_VAR",
			defaultValue: "default",
			envValue:     "custom",
			expected:     "custom",
		},
		{
			name:         "env var not set",
			key:          "TEST_VAR_MISSING",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
		{
			name:         "env var empty string",
			key:          "TEST_VAR_EMPTY",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env var if provided
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.defaultValue)

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		envValue     string
		expected     int
	}{
		{
			name:         "valid int",
			key:          "TEST_INT",
			defaultValue: 100,
			envValue:     "200",
			expected:     200,
		},
		{
			name:         "invalid int",
			key:          "TEST_INT_INVALID",
			defaultValue: 100,
			envValue:     "invalid",
			expected:     100,
		},
		{
			name:         "env var not set",
			key:          "TEST_INT_MISSING",
			defaultValue: 100,
			envValue:     "",
			expected:     100,
		},
		{
			name:         "zero value",
			key:          "TEST_INT_ZERO",
			defaultValue: 100,
			envValue:     "0",
			expected:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env var if provided
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnvInt(tt.key, tt.defaultValue)

			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}
