//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"
)

var (
	baseURL = getEnvOrDefault("E2E_BASE_URL", "http://localhost:18080")
	testDB  = getEnvOrDefault("E2E_DB_URL", "root:rootpassword@tcp(localhost:3307)/test_e2e")
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func TestCompleteUserJourney_E2E(t *testing.T) {
	// Create HTTP expect instance
	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  baseURL,
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	})

	// Generate unique email for this test run
	email := fmt.Sprintf("e2e-test-%d@example.com", time.Now().Unix())

	t.Run("1. User Registration", func(t *testing.T) {
		resp := e.POST("/api/v1/auth/signup").
			WithJSON(map[string]interface{}{
				"email":        email,
				"password":     "Test123!@#",
				"name":         "E2E Test User",
				"service_type": "game",
			}).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		resp.Value("success").Boolean().IsTrue()
		resp.Value("data").Object().ContainsKey("user_id")
		resp.Value("data").Object().ContainsKey("access_token")
		resp.Value("data").Object().ContainsKey("refresh_token")
	})

	var accessToken string
	var refreshToken string
	var userID int

	t.Run("2. User Login", func(t *testing.T) {
		resp := e.POST("/api/v1/auth/signin").
			WithJSON(map[string]interface{}{
				"email":        email,
				"password":     "Test123!@#",
				"service_type": "game",
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		resp.Value("success").Boolean().IsTrue()
		data := resp.Value("data").Object()

		userID = int(data.Value("user_id").Number().Raw())
		accessToken = data.Value("access_token").String().Raw()
		refreshToken = data.Value("refresh_token").String().Raw()

		require.NotEmpty(t, accessToken, "Access token should not be empty")
		require.NotEmpty(t, refreshToken, "Refresh token should not be empty")
	})

	t.Run("3. Get User Profile", func(t *testing.T) {
		resp := e.GET("/api/v1/user/profile").
			WithHeader("Authorization", "Bearer "+accessToken).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		resp.Value("success").Boolean().IsTrue()
		profile := resp.Value("data").Object()
		profile.Value("email").String().IsEqual(email)
		profile.Value("name").String().IsEqual("E2E Test User")
	})

	t.Run("4. Upload File", func(t *testing.T) {
		// Create a test file
		testContent := []byte("This is a test file for E2E testing")

		resp := e.POST("/api/v1/files/upload").
			WithHeader("Authorization", "Bearer "+accessToken).
			WithMultipart().
			WithFile("file", "test.txt", testContent).
			WithFormField("description", "E2E test file").
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		resp.Value("success").Boolean().IsTrue()
		fileData := resp.Value("data").Object()
		fileData.ContainsKey("file_id")
		fileData.ContainsKey("file_name")
		fileData.Value("file_name").String().IsEqual("test.txt")
	})

	t.Run("5. Create Game Room", func(t *testing.T) {
		resp := e.POST("/api/v1/tower-defense/rooms").
			WithHeader("Authorization", "Bearer "+accessToken).
			WithJSON(map[string]interface{}{
				"name":        "E2E Test Room",
				"max_players": 4,
			}).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		resp.Value("success").Boolean().IsTrue()
		roomData := resp.Value("data").Object()
		roomData.ContainsKey("room_id")
		roomData.Value("name").String().IsEqual("E2E Test Room")
		roomData.Value("host_user_id").Number().IsEqual(userID)
	})

	t.Run("6. Refresh Token", func(t *testing.T) {
		resp := e.POST("/api/v1/auth/refresh").
			WithJSON(map[string]interface{}{
				"refresh_token": refreshToken,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		resp.Value("success").Boolean().IsTrue()
		data := resp.Value("data").Object()
		newAccessToken := data.Value("access_token").String().Raw()
		require.NotEmpty(t, newAccessToken, "New access token should not be empty")
		require.NotEqual(t, accessToken, newAccessToken, "Should receive new access token")
	})

	t.Run("7. Logout", func(t *testing.T) {
		e.POST("/api/v1/auth/logout").
			WithHeader("Authorization", "Bearer "+accessToken).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			Value("success").Boolean().IsTrue()

		// Verify token is invalidated
		e.GET("/api/v1/user/profile").
			WithHeader("Authorization", "Bearer "+accessToken).
			Expect().
			Status(http.StatusUnauthorized)
	})
}

func TestConcurrentGameSessions_E2E(t *testing.T) {
	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  baseURL,
		Reporter: httpexpect.NewAssertReporter(t),
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	})

	// Create multiple users
	const numPlayers = 4
	players := make([]struct {
		email       string
		accessToken string
		userID      int
	}, numPlayers)

	// Register and login all players
	for i := 0; i < numPlayers; i++ {
		email := fmt.Sprintf("player%d-%d@example.com", i+1, time.Now().Unix())

		// Register
		e.POST("/api/v1/auth/signup").
			WithJSON(map[string]interface{}{
				"email":        email,
				"password":     "Player123!",
				"name":         fmt.Sprintf("Player %d", i+1),
				"service_type": "game",
			}).
			Expect().
			Status(http.StatusCreated)

		// Login
		resp := e.POST("/api/v1/auth/signin").
			WithJSON(map[string]interface{}{
				"email":        email,
				"password":     "Player123!",
				"service_type": "game",
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		data := resp.Value("data").Object()
		players[i].email = email
		players[i].accessToken = data.Value("access_token").String().Raw()
		players[i].userID = int(data.Value("user_id").Number().Raw())
	}

	// Player 1 creates a room
	var roomID int
	t.Run("Create Room", func(t *testing.T) {
		resp := e.POST("/api/v1/tower-defense/rooms").
			WithHeader("Authorization", "Bearer "+players[0].accessToken).
			WithJSON(map[string]interface{}{
				"name":        "Concurrent Test Room",
				"max_players": 4,
			}).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		roomID = int(resp.Value("data").Object().Value("room_id").Number().Raw())
	})

	// Other players join the room concurrently
	t.Run("Concurrent Join", func(t *testing.T) {
		joinResults := make(chan bool, numPlayers-1)

		for i := 1; i < numPlayers; i++ {
			go func(playerIndex int) {
				resp := e.POST(fmt.Sprintf("/api/v1/tower-defense/rooms/%d/join", roomID)).
					WithHeader("Authorization", "Bearer "+players[playerIndex].accessToken).
					Expect()

				joinResults <- (resp.Raw().StatusCode == http.StatusOK)
			}(i)
		}

		// Collect results
		successCount := 0
		for i := 1; i < numPlayers; i++ {
			if <-joinResults {
				successCount++
			}
		}

		require.Equal(t, numPlayers-1, successCount, "All players should join successfully")
	})

	// Verify room status
	t.Run("Verify Room State", func(t *testing.T) {
		resp := e.GET(fmt.Sprintf("/api/v1/tower-defense/rooms/%d", roomID)).
			WithHeader("Authorization", "Bearer "+players[0].accessToken).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		roomData := resp.Value("data").Object()
		roomData.Value("player_count").Number().IsEqual(numPlayers)
		roomData.Value("max_players").Number().IsEqual(4)
	})

	// Start game
	t.Run("Start Game", func(t *testing.T) {
		e.POST(fmt.Sprintf("/api/v1/tower-defense/rooms/%d/start", roomID)).
			WithHeader("Authorization", "Bearer "+players[0].accessToken).
			Expect().
			Status(http.StatusOK)
	})
}

func TestErrorHandling_E2E(t *testing.T) {
	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  baseURL,
		Reporter: httpexpect.NewAssertReporter(t),
	})

	t.Run("Invalid Email Format", func(t *testing.T) {
		resp := e.POST("/api/v1/auth/signup").
			WithJSON(map[string]interface{}{
				"email":        "invalid-email",
				"password":     "Test123!",
				"name":         "Test User",
				"service_type": "game",
			}).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object()

		resp.Value("success").Boolean().IsFalse()
		resp.Value("error").String().Contains("email")
	})

	t.Run("Duplicate Email", func(t *testing.T) {
		email := fmt.Sprintf("duplicate-%d@example.com", time.Now().Unix())

		// First registration
		e.POST("/api/v1/auth/signup").
			WithJSON(map[string]interface{}{
				"email":        email,
				"password":     "Test123!",
				"name":         "First User",
				"service_type": "game",
			}).
			Expect().
			Status(http.StatusCreated)

		// Duplicate registration
		resp := e.POST("/api/v1/auth/signup").
			WithJSON(map[string]interface{}{
				"email":        email,
				"password":     "Test123!",
				"name":         "Second User",
				"service_type": "game",
			}).
			Expect().
			Status(http.StatusConflict).
			JSON().Object()

		resp.Value("success").Boolean().IsFalse()
		resp.Value("error").String().Contains("already exists")
	})

	t.Run("Unauthorized Access", func(t *testing.T) {
		e.GET("/api/v1/user/profile").
			Expect().
			Status(http.StatusUnauthorized)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		e.GET("/api/v1/user/profile").
			WithHeader("Authorization", "Bearer invalid-token-here").
			Expect().
			Status(http.StatusUnauthorized)
	})

	t.Run("Resource Not Found", func(t *testing.T) {
		e.GET("/api/v1/tower-defense/rooms/999999").
			Expect().
			Status(http.StatusNotFound)
	})
}