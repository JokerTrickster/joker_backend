package utils

import "os"

// GetEnv retrieves environment variable with fallback to default value
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
