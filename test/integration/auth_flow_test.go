package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompleteAuthFlow tests the complete authentication flow
func TestCompleteAuthFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test server
	router := setupTestRouter()
	ts := httptest.NewServer(router)
	defer ts.Close()

	// Test data
	email := "integration@test.com"
	password := "SecurePass123!"
	nickname := "IntegrationUser"

	t.Run("Complete auth flow", func(t *testing.T) {
		// Step 1: Sign up
		signupReq := map[string]string{
			"email":    email,
			"password": password,
			"nickname": nickname,
		}

		signupResp := makeRequest(t, ts, "POST", "/api/v1/auth/signup", signupReq)
		assert.Equal(t, http.StatusCreated, signupResp.StatusCode)

		var signupResult map[string]interface{}
		err := json.NewDecoder(signupResp.Body).Decode(&signupResult)
		require.NoError(t, err)
		assert.Contains(t, signupResult, "user_id")
		assert.Contains(t, signupResult, "message")

		// Step 2: Sign in
		signinReq := map[string]string{
			"email":    email,
			"password": password,
		}

		signinResp := makeRequest(t, ts, "POST", "/api/v1/auth/signin", signinReq)
		assert.Equal(t, http.StatusOK, signinResp.StatusCode)

		var signinResult map[string]interface{}
		err = json.NewDecoder(signinResp.Body).Decode(&signinResult)
		require.NoError(t, err)
		assert.Contains(t, signinResult, "access_token")
		assert.Contains(t, signinResult, "refresh_token")

		accessToken := signinResult["access_token"].(string)
		refreshToken := signinResult["refresh_token"].(string)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)

		// Step 3: Access protected endpoint
		protectedResp := makeAuthRequest(t, ts, "GET", "/api/v1/user/profile", nil, accessToken)
		assert.Equal(t, http.StatusOK, protectedResp.StatusCode)

		var profileResult map[string]interface{}
		err = json.NewDecoder(protectedResp.Body).Decode(&profileResult)
		require.NoError(t, err)
		assert.Equal(t, email, profileResult["email"])
		assert.Equal(t, nickname, profileResult["nickname"])

		// Step 4: Refresh token
		refreshReq := map[string]string{
			"refresh_token": refreshToken,
		}

		refreshResp := makeRequest(t, ts, "POST", "/api/v1/auth/refresh", refreshReq)
		assert.Equal(t, http.StatusOK, refreshResp.StatusCode)

		var refreshResult map[string]interface{}
		err = json.NewDecoder(refreshResp.Body).Decode(&refreshResult)
		require.NoError(t, err)
		newAccessToken := refreshResult["access_token"].(string)
		assert.NotEmpty(t, newAccessToken)
		assert.NotEqual(t, accessToken, newAccessToken, "Should receive new access token")

		// Step 5: Sign out
		signoutResp := makeAuthRequest(t, ts, "POST", "/api/v1/auth/signout", nil, newAccessToken)
		assert.Equal(t, http.StatusOK, signoutResp.StatusCode)

		// Step 6: Verify token is invalidated
		invalidResp := makeAuthRequest(t, ts, "GET", "/api/v1/user/profile", nil, newAccessToken)
		assert.Equal(t, http.StatusUnauthorized, invalidResp.StatusCode)
	})
}

// TestConcurrentAuthentication tests concurrent authentication requests
func TestConcurrentAuthentication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	router := setupTestRouter()
	ts := httptest.NewServer(router)
	defer ts.Close()

	// Create multiple users concurrently
	numUsers := 10
	results := make(chan bool, numUsers)

	for i := 0; i < numUsers; i++ {
		go func(index int) {
			email := fmt.Sprintf("concurrent%d@test.com", index)
			password := "SecurePass123!"
			nickname := fmt.Sprintf("User%d", index)

			// Sign up
			signupReq := map[string]string{
				"email":    email,
				"password": password,
				"nickname": nickname,
			}

			signupResp := makeRequest(t, ts, "POST", "/api/v1/auth/signup", signupReq)
			if signupResp.StatusCode != http.StatusCreated {
				results <- false
				return
			}

			// Sign in
			signinReq := map[string]string{
				"email":    email,
				"password": password,
			}

			signinResp := makeRequest(t, ts, "POST", "/api/v1/auth/signin", signinReq)
			results <- signinResp.StatusCode == http.StatusOK
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < numUsers; i++ {
		if <-results {
			successCount++
		}
	}

	assert.Equal(t, numUsers, successCount, "All concurrent auth requests should succeed")
}

// TestSessionExpiration tests token expiration handling
func TestSessionExpiration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	router := setupTestRouter()
	ts := httptest.NewServer(router)
	defer ts.Close()

	// Create user and get token
	email := "expiry@test.com"
	password := "SecurePass123!"

	signupReq := map[string]string{
		"email":    email,
		"password": password,
		"nickname": "ExpiryTest",
	}

	makeRequest(t, ts, "POST", "/api/v1/auth/signup", signupReq)

	signinReq := map[string]string{
		"email":    email,
		"password": password,
	}

	signinResp := makeRequest(t, ts, "POST", "/api/v1/auth/signin", signinReq)
	var signinResult map[string]interface{}
	json.NewDecoder(signinResp.Body).Decode(&signinResult)
	accessToken := signinResult["access_token"].(string)

	// Verify token works initially
	resp1 := makeAuthRequest(t, ts, "GET", "/api/v1/user/profile", nil, accessToken)
	assert.Equal(t, http.StatusOK, resp1.StatusCode)

	// Simulate token expiration (in real scenario, would wait or manipulate time)
	// For this test, we'll create an expired token
	expiredToken := createExpiredToken(email)

	// Try to use expired token
	resp2 := makeAuthRequest(t, ts, "GET", "/api/v1/user/profile", nil, expiredToken)
	assert.Equal(t, http.StatusUnauthorized, resp2.StatusCode)

	var errorResult map[string]interface{}
	json.NewDecoder(resp2.Body).Decode(&errorResult)
	assert.Contains(t, errorResult["error"], "token expired")
}

// TestPasswordReset tests password reset flow
func TestPasswordReset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	router := setupTestRouter()
	ts := httptest.NewServer(router)
	defer ts.Close()

	email := "reset@test.com"
	oldPassword := "OldPass123!"
	newPassword := "NewPass456!"

	// Create user
	signupReq := map[string]string{
		"email":    email,
		"password": oldPassword,
		"nickname": "ResetUser",
	}
	makeRequest(t, ts, "POST", "/api/v1/auth/signup", signupReq)

	// Request password reset
	resetReq := map[string]string{
		"email": email,
	}
	resetResp := makeRequest(t, ts, "POST", "/api/v1/auth/reset-password", resetReq)
	assert.Equal(t, http.StatusOK, resetResp.StatusCode)

	// In real scenario, would retrieve reset token from email
	// For testing, we'll simulate having the reset token
	resetToken := "mock-reset-token"

	// Reset password with token
	confirmResetReq := map[string]string{
		"token":        resetToken,
		"new_password": newPassword,
	}
	confirmResp := makeRequest(t, ts, "POST", "/api/v1/auth/confirm-reset", confirmResetReq)
	assert.Equal(t, http.StatusOK, confirmResp.StatusCode)

	// Try to sign in with old password (should fail)
	oldSigninReq := map[string]string{
		"email":    email,
		"password": oldPassword,
	}
	oldSigninResp := makeRequest(t, ts, "POST", "/api/v1/auth/signin", oldSigninReq)
	assert.Equal(t, http.StatusUnauthorized, oldSigninResp.StatusCode)

	// Sign in with new password (should succeed)
	newSigninReq := map[string]string{
		"email":    email,
		"password": newPassword,
	}
	newSigninResp := makeRequest(t, ts, "POST", "/api/v1/auth/signin", newSigninReq)
	assert.Equal(t, http.StatusOK, newSigninResp.StatusCode)
}

// Helper functions

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup your actual routes here
	// This is a placeholder - you would use your actual router setup
	// router.Use(middleware.CORS())
	// router.Use(middleware.Logger())
	// routes.SetupRoutes(router)

	return router
}

func makeRequest(t *testing.T, ts *httptest.Server, method, path string, body interface{}) *http.Response {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		require.NoError(t, err)
	}

	req, err := http.NewRequest(method, ts.URL+path, bytes.NewBuffer(reqBody))
	require.NoError(t, err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)

	return resp
}

func makeAuthRequest(t *testing.T, ts *httptest.Server, method, path string, body interface{}, token string) *http.Response {
	resp := makeRequest(t, ts, method, path, body)
	resp.Request.Header.Set("Authorization", "Bearer "+token)
	return resp
}

func createExpiredToken(email string) string {
	// In real implementation, this would create a JWT with expired timestamp
	return "expired.token.here"
}