package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateRequestID generates a cryptographically secure request ID
// Returns a 16-byte (32 hex characters) unique identifier
func GenerateRequestID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails
		// This should be extremely rare in practice
		return fmt.Sprintf("fallback-%d", TimeToEpochMillis(time.Now()))
	}
	return hex.EncodeToString(b)
}
