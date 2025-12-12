package utils

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestGenerateRequestID(t *testing.T) {
	// Test that GenerateRequestID returns a valid hex string
	reqID := GenerateRequestID()

	// Should be 32 characters (16 bytes in hex)
	if len(reqID) != 32 && !strings.HasPrefix(reqID, "fallback-") {
		t.Errorf("Expected 32 character hex string or fallback, got length: %d, value: %s", len(reqID), reqID)
	}

	// Should be valid hex if not fallback
	if !strings.HasPrefix(reqID, "fallback-") {
		_, err := hex.DecodeString(reqID)
		if err != nil {
			t.Errorf("GenerateRequestID returned invalid hex string: %v", err)
		}
	}
}

func TestGenerateRequestIDUniqueness(t *testing.T) {
	// Generate multiple request IDs and ensure they're unique
	ids := make(map[string]bool)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		reqID := GenerateRequestID()
		if ids[reqID] {
			t.Errorf("Generated duplicate request ID: %s", reqID)
		}
		ids[reqID] = true
	}

	if len(ids) != iterations {
		t.Errorf("Expected %d unique IDs, got %d", iterations, len(ids))
	}
}

func TestGenerateRequestIDFormat(t *testing.T) {
	reqID := GenerateRequestID()

	// Should not be empty
	if reqID == "" {
		t.Error("GenerateRequestID returned empty string")
	}

	// Should be lowercase hex or fallback
	if !strings.HasPrefix(reqID, "fallback-") {
		for _, c := range reqID {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				t.Errorf("GenerateRequestID contains invalid character: %c in %s", c, reqID)
			}
		}
	}
}

func BenchmarkGenerateRequestID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateRequestID()
	}
}
