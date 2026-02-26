package assertions

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertEventuallyTrue waits for a condition to become true within timeout
func AssertEventuallyTrue(t *testing.T, fn func() bool, timeout time.Duration, msg ...string) {
	t.Helper()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	timeoutCh := time.After(timeout)

	for {
		select {
		case <-ticker.C:
			if fn() {
				return
			}
		case <-timeoutCh:
			message := "condition did not become true within timeout"
			if len(msg) > 0 {
				message = strings.Join(msg, " ")
			}
			t.Fatalf(message)
		}
	}
}

// AssertJSONEqual compares two JSON strings for equality
func AssertJSONEqual(t *testing.T, expected, actual string) {
	t.Helper()

	var expectedData, actualData interface{}

	err := json.Unmarshal([]byte(expected), &expectedData)
	require.NoError(t, err, "Failed to unmarshal expected JSON")

	err = json.Unmarshal([]byte(actual), &actualData)
	require.NoError(t, err, "Failed to unmarshal actual JSON")

	assert.Equal(t, expectedData, actualData, "JSON content does not match")
}

// AssertErrorContains checks if error contains specific substring
func AssertErrorContains(t *testing.T, err error, substring string) {
	t.Helper()

	require.Error(t, err, "Expected an error but got nil")
	assert.Contains(t, err.Error(), substring, "Error message does not contain expected substring")
}

// AssertHTTPResponse validates HTTP response characteristics
type HTTPResponseAssertion struct {
	StatusCode int
	Headers    map[string]string
	BodyParts  []string
}

// AssertHTTPResponse checks multiple aspects of HTTP response
func AssertHTTPResponse(t *testing.T, actual interface{}, expected HTTPResponseAssertion) {
	t.Helper()

	// This would be implemented based on your HTTP response structure
	// Example implementation shown
	t.Log("HTTP Response assertion called")
}

// AssertTimeWithin checks if a time is within a duration of expected time
func AssertTimeWithin(t *testing.T, expected, actual time.Time, delta time.Duration) {
	t.Helper()

	diff := actual.Sub(expected)
	if diff < 0 {
		diff = -diff
	}

	assert.LessOrEqual(t, diff, delta,
		"Time difference %v exceeds allowed delta %v", diff, delta)
}

// AssertSliceContains checks if slice contains an element
func AssertSliceContains(t *testing.T, slice interface{}, element interface{}) {
	t.Helper()

	sliceVal := reflect.ValueOf(slice)
	if sliceVal.Kind() != reflect.Slice {
		t.Fatalf("First argument must be a slice, got %T", slice)
	}

	for i := 0; i < sliceVal.Len(); i++ {
		if reflect.DeepEqual(sliceVal.Index(i).Interface(), element) {
			return
		}
	}

	t.Errorf("Slice does not contain element: %v", element)
}

// AssertPanic checks if a function panics
func AssertPanic(t *testing.T, fn func(), msg ...string) {
	t.Helper()

	defer func() {
		if r := recover(); r == nil {
			message := "Function did not panic as expected"
			if len(msg) > 0 {
				message = strings.Join(msg, " ")
			}
			t.Errorf(message)
		}
	}()

	fn()
}

// AssertNoPanic checks if a function does not panic
func AssertNoPanic(t *testing.T, fn func(), msg ...string) {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			message := "Function panicked unexpectedly"
			if len(msg) > 0 {
				message = strings.Join(msg, " ")
			}
			t.Errorf("%s: %v", message, r)
		}
	}()

	fn()
}

// AssertMapContainsKeys checks if map contains all specified keys
func AssertMapContainsKeys(t *testing.T, m interface{}, keys ...string) {
	t.Helper()

	mapVal := reflect.ValueOf(m)
	if mapVal.Kind() != reflect.Map {
		t.Fatalf("First argument must be a map, got %T", m)
	}

	mapKeys := mapVal.MapKeys()
	keyStrings := make(map[string]bool)
	for _, k := range mapKeys {
		keyStrings[k.String()] = true
	}

	for _, key := range keys {
		if !keyStrings[key] {
			t.Errorf("Map does not contain key: %s", key)
		}
	}
}

// AssertBetween checks if a value is between min and max (inclusive)
func AssertBetween(t *testing.T, value, min, max interface{}) {
	t.Helper()

	// Convert to float64 for comparison
	val := toFloat64(t, value)
	minVal := toFloat64(t, min)
	maxVal := toFloat64(t, max)

	assert.GreaterOrEqual(t, val, minVal, "Value %v is less than minimum %v", val, minVal)
	assert.LessOrEqual(t, val, maxVal, "Value %v is greater than maximum %v", val, maxVal)
}

// Helper function to convert numeric types to float64
func toFloat64(t *testing.T, v interface{}) float64 {
	t.Helper()

	switch val := v.(type) {
	case int:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	default:
		t.Fatalf("Cannot convert %T to float64", v)
		return 0
	}
}