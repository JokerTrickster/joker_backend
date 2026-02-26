//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/shared/testutils/db"
	"github.com/JokerTrickster/joker_backend/shared/testutils/fixtures"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type IntegrationTestSuite struct {
	db     *gorm.DB
	router *gin.Engine
}

func setupTestSuite(t *testing.T) *IntegrationTestSuite {
	// Setup test database
	testDB := db.SetupTestDB(t)

	// Run migrations
	err := db.RunMigrations(testDB,
		&User{},     // Your user entity
		&Session{},  // Your session entity
		&File{},     // Your file entity
		&Room{},     // Your room entity
	)
	require.NoError(t, err, "Failed to run migrations")

	// Setup router with real dependencies
	router := setupRouter(testDB)

	return &IntegrationTestSuite{
		db:     testDB,
		router: router,
	}
}

func setupRouter(database *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup your actual routes here with real implementations
	// Example:
	// authRepo := repository.NewAuthRepository(database)
	// authUseCase := usecase.NewAuthUseCase(authRepo)
	// authHandler := handler.NewAuthHandler(authUseCase)
	// router.POST("/auth/signup", authHandler.SignUp)
	// router.POST("/auth/signin", authHandler.SignIn)

	return router
}

func TestAuthFlow_Integration(t *testing.T) {
	suite := setupTestSuite(t)

	t.Run("Complete Authentication Flow", func(t *testing.T) {
		// Step 1: Sign up a new user
		signupReq := map[string]interface{}{
			"email":        "integration@test.com",
			"password":     "testpass123",
			"name":         "Integration Test",
			"service_type": "game",
		}

		body, _ := json.Marshal(signupReq)
		req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		suite.router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var signupResp map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &signupResp)
		require.NoError(t, err)

		data := signupResp["data"].(map[string]interface{})
		accessToken := data["access_token"].(string)
		assert.NotEmpty(t, accessToken)

		// Step 2: Sign in with the created user
		signinReq := map[string]interface{}{
			"email":        "integration@test.com",
			"password":     "testpass123",
			"service_type": "game",
		}

		body, _ = json.Marshal(signinReq)
		req = httptest.NewRequest(http.MethodPost, "/auth/signin", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec = httptest.NewRecorder()

		suite.router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		var signinResp map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &signinResp)
		require.NoError(t, err)

		data = signinResp["data"].(map[string]interface{})
		newAccessToken := data["access_token"].(string)
		assert.NotEmpty(t, newAccessToken)

		// Step 3: Use the token to access protected endpoint
		req = httptest.NewRequest(http.MethodGet, "/user/profile", nil)
		req.Header.Set("Authorization", "Bearer "+newAccessToken)
		rec = httptest.NewRecorder()

		suite.router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		// Step 4: Try with wrong password
		wrongPassReq := map[string]interface{}{
			"email":        "integration@test.com",
			"password":     "wrongpass",
			"service_type": "game",
		}

		body, _ = json.Marshal(wrongPassReq)
		req = httptest.NewRequest(http.MethodPost, "/auth/signin", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec = httptest.NewRecorder()

		suite.router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestConcurrentUserCreation_Integration(t *testing.T) {
	suite := setupTestSuite(t)

	const numUsers = 10
	results := make(chan int, numUsers)

	for i := 0; i < numUsers; i++ {
		go func(index int) {
			signupReq := map[string]interface{}{
				"email":        fmt.Sprintf("concurrent%d@test.com", index),
				"password":     "testpass123",
				"name":         fmt.Sprintf("User %d", index),
				"service_type": "game",
			}

			body, _ := json.Marshal(signupReq)
			req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			suite.router.ServeHTTP(rec, req)
			results <- rec.Code
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < numUsers; i++ {
		code := <-results
		if code == http.StatusCreated {
			successCount++
		}
	}

	assert.Equal(t, numUsers, successCount, "All concurrent users should be created successfully")

	// Verify all users exist in database
	var count int64
	suite.db.Model(&User{}).Where("email LIKE ?", "concurrent%@test.com").Count(&count)
	assert.Equal(t, int64(numUsers), count, "All users should be in database")
}

func TestDatabaseRollback_Integration(t *testing.T) {
	suite := setupTestSuite(t)

	// Use transaction for testing rollback
	db.WithTransaction(t, suite.db, func(tx *gorm.DB) {
		// Create a user within transaction
		user := &User{
			Email:    "rollback@test.com",
			Password: "hashed_password",
			Name:     "Rollback Test",
		}

		err := tx.Create(user).Error
		require.NoError(t, err, "User should be created in transaction")

		// Verify user exists in transaction
		var count int64
		tx.Model(&User{}).Where("email = ?", "rollback@test.com").Count(&count)
		assert.Equal(t, int64(1), count, "User should exist in transaction")

		// Transaction will be rolled back automatically
	})

	// Verify user doesn't exist after rollback
	var count int64
	suite.db.Model(&User{}).Where("email = ?", "rollback@test.com").Count(&count)
	assert.Equal(t, int64(0), count, "User should not exist after rollback")
}

func TestFileUploadWithAuth_Integration(t *testing.T) {
	suite := setupTestSuite(t)

	// Setup: Create and authenticate user
	loader := fixtures.NewFixtureLoader(suite.db)
	user, err := loader.LoadUser(&fixtures.UserFixture{
		Email:       "filetest@example.com",
		Password:    "testpass123",
		Name:        "File Test User",
		ServiceType: "game",
	})
	require.NoError(t, err)

	// Get auth token
	signinReq := map[string]interface{}{
		"email":        user.Email,
		"password":     "testpass123",
		"service_type": "game",
	}

	body, _ := json.Marshal(signinReq)
	req := httptest.NewRequest(http.MethodPost, "/auth/signin", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	suite.router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var signinResp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &signinResp)
	token := signinResp["data"].(map[string]interface{})["access_token"].(string)

	// Test file upload with authentication
	uploadReq := map[string]interface{}{
		"filename": "test.txt",
		"size":     1024,
		"type":     "text/plain",
	}

	body, _ = json.Marshal(uploadReq)
	req = httptest.NewRequest(http.MethodPost, "/files/upload", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()

	suite.router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify file record in database
	var fileCount int64
	suite.db.Model(&File{}).Where("user_id = ?", user.ID).Count(&fileCount)
	assert.Equal(t, int64(1), fileCount, "File should be associated with user")
}