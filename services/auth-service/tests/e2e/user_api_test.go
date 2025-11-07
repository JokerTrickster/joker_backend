package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// TestUserAPI_CreateUser tests user creation endpoint
func TestUserAPI_CreateUser(t *testing.T) {
	// Clean up before test
	cleanupTestData()

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedError  string
		validate       func(t *testing.T, body map[string]interface{})
	}{
		{
			name:           "Success - Valid user creation",
			requestBody:    `{"name":"John Doe","email":"john@example.com"}`,
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, body map[string]interface{}) {
				if !body["success"].(bool) {
					t.Error("Expected success to be true")
				}

				data := body["data"].(map[string]interface{})
				if data["name"] != "John Doe" {
					t.Errorf("Expected name to be 'John Doe', got %v", data["name"])
				}
				if data["email"] != "john@example.com" {
					t.Errorf("Expected email to be 'john@example.com', got %v", data["email"])
				}
				if data["id"] == nil {
					t.Error("Expected ID to be set")
				}
			},
		},
		{
			name:           "Fail - Missing email",
			requestBody:    `{"name":"John Doe"}`,
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, body map[string]interface{}) {
				if body["success"].(bool) {
					t.Error("Expected success to be false")
				}
			},
		},
		{
			name:           "Fail - Missing name",
			requestBody:    `{"email":"john@example.com"}`,
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, body map[string]interface{}) {
				if body["success"].(bool) {
					t.Error("Expected success to be false")
				}
			},
		},
		{
			name:           "Fail - Empty email",
			requestBody:    `{"name":"John Doe","email":""}`,
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, body map[string]interface{}) {
				if body["success"].(bool) {
					t.Error("Expected success to be false")
				}
			},
		},
		{
			name:           "Fail - Invalid JSON",
			requestBody:    `{"name":"John Doe","email":}`,
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, body map[string]interface{}) {
				if body["success"].(bool) {
					t.Error("Expected success to be false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			testServer.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			var responseBody map[string]interface{}
			if err := json.Unmarshal(rec.Body.Bytes(), &responseBody); err != nil {
				t.Fatalf("Failed to parse response body: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, responseBody)
			}
		})
	}
}

// TestUserAPI_GetUser tests user retrieval endpoint
func TestUserAPI_GetUser(t *testing.T) {
	// Clean up and setup test data
	cleanupTestData()

	// Create test user
	testUser, err := createTestUser("Jane Doe", "jane@example.com")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	defer deleteTestUser(testUser.ID)

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		validate       func(t *testing.T, body map[string]interface{})
	}{
		{
			name:           "Success - Get existing user",
			userID:         fmt.Sprintf("%d", testUser.ID),
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, body map[string]interface{}) {
				if !body["success"].(bool) {
					t.Error("Expected success to be true")
				}

				data := body["data"].(map[string]interface{})
				if data["name"] != "Jane Doe" {
					t.Errorf("Expected name to be 'Jane Doe', got %v", data["name"])
				}
				if data["email"] != "jane@example.com" {
					t.Errorf("Expected email to be 'jane@example.com', got %v", data["email"])
				}
			},
		},
		{
			name:           "Fail - User not found",
			userID:         "999999",
			expectedStatus: http.StatusNotFound,
			validate: func(t *testing.T, body map[string]interface{}) {
				if body["success"].(bool) {
					t.Error("Expected success to be false")
				}
			},
		},
		{
			name:           "Fail - Invalid user ID format",
			userID:         "invalid",
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, body map[string]interface{}) {
				if body["success"].(bool) {
					t.Error("Expected success to be false")
				}
			},
		},
		{
			name:           "Fail - Negative user ID",
			userID:         "-1",
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, body map[string]interface{}) {
				if body["success"].(bool) {
					t.Error("Expected success to be false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+tt.userID, nil)
			rec := httptest.NewRecorder()

			testServer.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			var responseBody map[string]interface{}
			if err := json.Unmarshal(rec.Body.Bytes(), &responseBody); err != nil {
				t.Fatalf("Failed to parse response body: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, responseBody)
			}
		})
	}
}

// TestUserAPI_HealthCheck tests health endpoint
func TestUserAPI_HealthCheck(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	testServer.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var responseBody map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	if !responseBody["success"].(bool) {
		t.Error("Expected success to be true")
	}
}

// TestUserAPI_DuplicateEmail tests duplicate email handling
func TestUserAPI_DuplicateEmail(t *testing.T) {
	cleanupTestData()

	// Create first user
	requestBody := `{"name":"User One","email":"duplicate@example.com"}`
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(requestBody))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	testServer.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusCreated {
		t.Fatalf("First user creation failed with status %d", rec1.Code)
	}

	// Try to create second user with same email
	requestBody2 := `{"name":"User Two","email":"duplicate@example.com"}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(requestBody2))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	testServer.ServeHTTP(rec2, req2)

	// Should fail due to unique constraint
	if rec2.Code == http.StatusCreated {
		t.Error("Expected duplicate email to fail, but it succeeded")
	}
}

// TestUserAPI_ConcurrentCreation tests concurrent user creation
func TestUserAPI_ConcurrentCreation(t *testing.T) {
	cleanupTestData()

	numRequests := 10
	done := make(chan bool, numRequests)
	var successCount int32 // Use atomic type to prevent race condition

	for i := 0; i < numRequests; i++ {
		go func(index int) {
			requestBody := fmt.Sprintf(`{"name":"Concurrent User %d","email":"concurrent%d@example.com"}`, index, index)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			testServer.ServeHTTP(rec, req)

			if rec.Code == http.StatusCreated {
				atomic.AddInt32(&successCount, 1) // Atomic increment
			}

			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		<-done
	}

	// Verify all users were created
	count, err := countUsers()
	if err != nil {
		t.Fatalf("Failed to count users: %v", err)
	}

	if count != numRequests {
		t.Errorf("Expected %d users to be created, got %d", numRequests, count)
	}
}
