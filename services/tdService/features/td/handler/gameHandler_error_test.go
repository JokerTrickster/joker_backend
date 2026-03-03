package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/request"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// TestGameHandler_CreateSession_ErrorCases tests various error scenarios
func TestGameHandler_CreateSession_ErrorCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		request        interface{}
		setupMocks     func(*MockPlayerRepository, *MockGameSessionRepository)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Invalid game mode",
			request: request.CreateSessionRequest{
				Mode:        "invalid_mode",
				PlayerCount: 1,
			},
			setupMocks: func(mp *MockPlayerRepository, ms *MockGameSessionRepository) {
				mp.On("GetPlayerByUserID", mock.Anything, testUserID).Return(testPlayer, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid game mode",
		},
		{
			name: "Player count exceeds maximum",
			request: request.CreateSessionRequest{
				Mode:        "multiplayer",
				PlayerCount: 10, // assuming max is 4
			},
			setupMocks: func(mp *MockPlayerRepository, ms *MockGameSessionRepository) {
				mp.On("GetPlayerByUserID", mock.Anything, testUserID).Return(testPlayer, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "player count exceeds maximum",
		},
		{
			name: "Player not found",
			request: request.CreateSessionRequest{
				Mode:        "single",
				PlayerCount: 1,
			},
			setupMocks: func(mp *MockPlayerRepository, ms *MockGameSessionRepository) {
				mp.On("GetPlayerByUserID", mock.Anything, testUserID).Return(nil, gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "player not found",
		},
		{
			name: "Database connection error",
			request: request.CreateSessionRequest{
				Mode:        "single",
				PlayerCount: 1,
			},
			setupMocks: func(mp *MockPlayerRepository, ms *MockGameSessionRepository) {
				mp.On("GetPlayerByUserID", mock.Anything, testUserID).Return(nil, errors.New("database connection lost"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "database error",
		},
		{
			name: "Session creation fails",
			request: request.CreateSessionRequest{
				Mode:        "single",
				PlayerCount: 1,
			},
			setupMocks: func(mp *MockPlayerRepository, ms *MockGameSessionRepository) {
				mp.On("GetPlayerByUserID", mock.Anything, testUserID).Return(testPlayer, nil)
				ms.On("CreateSession", mock.Anything, mock.Anything).Return(nil, errors.New("failed to create session"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to create session",
		},
		{
			name: "Concurrent session limit reached",
			request: request.CreateSessionRequest{
				Mode:        "single",
				PlayerCount: 1,
			},
			setupMocks: func(mp *MockPlayerRepository, ms *MockGameSessionRepository) {
				mp.On("GetPlayerByUserID", mock.Anything, testUserID).Return(testPlayer, nil)
				ms.On("GetActiveSessionCount", mock.Anything, testPlayerID).Return(5, nil) // assuming max is 5
				ms.On("CreateSession", mock.Anything, mock.Anything).Return(nil, errors.New("session limit reached"))
			},
			expectedStatus: http.StatusTooManyRequests,
			expectedError:  "session limit reached",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockPlayer := new(MockPlayerRepository)
			mockSession := new(MockGameSessionRepository)

			if tt.setupMocks != nil {
				tt.setupMocks(mockPlayer, mockSession)
			}

			authUC := newTestAuthUseCase(mockPlayer)
			gameUC := newTestGameUseCase(mockSession, mockPlayer)
			h := NewGameHandler(gameUC, authUC)

			// Create valid JWT token
			token := createValidJWT(t, testUserID)

			// Create request
			body := mustJSON(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/session", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			// Setup router and execute
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/session", h.CreateSession)
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}

			// Verify mock expectations
			mockPlayer.AssertExpectations(t)
			mockSession.AssertExpectations(t)
		})
	}
}

// TestGameHandler_JoinSession_ErrorCases tests join session error scenarios
func TestGameHandler_JoinSession_ErrorCases(t *testing.T) {
	tests := []struct {
		name           string
		roomCode       string
		setupMocks     func(*MockPlayerRepository, *MockGameSessionRepository)
		expectedStatus int
		expectedError  string
	}{
		{
			name:     "Empty room code",
			roomCode: "",
			setupMocks: func(mp *MockPlayerRepository, ms *MockGameSessionRepository) {
				// No mocks needed, should fail at validation
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "room code is required",
		},
		{
			name:     "Invalid room code format",
			roomCode: "INVALID!@#",
			setupMocks: func(mp *MockPlayerRepository, ms *MockGameSessionRepository) {
				// No mocks needed, should fail at validation
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid room code format",
		},
		{
			name:     "Room not found",
			roomCode: "ROOM123",
			setupMocks: func(mp *MockPlayerRepository, ms *MockGameSessionRepository) {
				mp.On("GetPlayerByUserID", mock.Anything, testUserID).Return(testPlayer, nil)
				ms.On("GetSessionByRoomCode", mock.Anything, "ROOM123").Return(nil, gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "room not found",
		},
		{
			name:     "Room is full",
			roomCode: "ROOM123",
			setupMocks: func(mp *MockPlayerRepository, ms *MockGameSessionRepository) {
				fullSession := &entity.GameSession{
					RoomCode:    "ROOM123",
					PlayerCount: 4,
					MaxPlayers:  4,
					Status:      entity.GameStatusWaiting,
				}
				mp.On("GetPlayerByUserID", mock.Anything, testUserID).Return(testPlayer, nil)
				ms.On("GetSessionByRoomCode", mock.Anything, "ROOM123").Return(fullSession, nil)
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "room is full",
		},
		{
			name:     "Game already started",
			roomCode: "ROOM123",
			setupMocks: func(mp *MockPlayerRepository, ms *MockGameSessionRepository) {
				startedSession := &entity.GameSession{
					RoomCode:    "ROOM123",
					PlayerCount: 2,
					MaxPlayers:  4,
					Status:      entity.GameStatusPlaying,
				}
				mp.On("GetPlayerByUserID", mock.Anything, testUserID).Return(testPlayer, nil)
				ms.On("GetSessionByRoomCode", mock.Anything, "ROOM123").Return(startedSession, nil)
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "game already started",
		},
		{
			name:     "Player already in another session",
			roomCode: "ROOM123",
			setupMocks: func(mp *MockPlayerRepository, ms *MockGameSessionRepository) {
				mp.On("GetPlayerByUserID", mock.Anything, testUserID).Return(testPlayer, nil)
				ms.On("GetActiveSessionByPlayerID", mock.Anything, testPlayerID).Return(testSession, nil)
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "already in another session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			// Setup mocks
			mockPlayer := new(MockPlayerRepository)
			mockSession := new(MockGameSessionRepository)

			if tt.setupMocks != nil {
				tt.setupMocks(mockPlayer, mockSession)
			}

			authUC := newTestAuthUseCase(mockPlayer)
			gameUC := newTestGameUseCase(mockSession, mockPlayer)
			h := NewGameHandler(gameUC, authUC)

			// Create valid JWT token
			token := createValidJWT(t, testUserID)

			// Create request
			body := mustJSON(map[string]string{"room_code": tt.roomCode})
			req := httptest.NewRequest(http.MethodPost, "/session/join", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			// Setup router and execute
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/session/join", h.JoinSession)
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}

			// Verify mock expectations
			mockPlayer.AssertExpectations(t)
			mockSession.AssertExpectations(t)
		})
	}
}

// TestGameHandler_ConcurrentSessionCreation tests race condition handling
func TestGameHandler_ConcurrentSessionCreation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockPlayer := new(MockPlayerRepository)
	mockSession := new(MockGameSessionRepository)

	// Setup mocks for concurrent access
	mockPlayer.On("GetPlayerByUserID", mock.Anything, testUserID).Return(testPlayer, nil)

	// Simulate concurrent session creation attempts
	var sessionCreated bool
	mockSession.On("CreateSession", mock.Anything, mock.Anything).Return(
		func(ctx context.Context, session *entity.GameSession) *entity.GameSession {
			if sessionCreated {
				return nil
			}
			sessionCreated = true
			return testSession
		},
		func(ctx context.Context, session *entity.GameSession) error {
			if sessionCreated {
				return errors.New("duplicate session")
			}
			return nil
		},
	)

	authUC := newTestAuthUseCase(mockPlayer)
	gameUC := newTestGameUseCase(mockSession, mockPlayer)
	h := NewGameHandler(gameUC, authUC)

	// Create multiple concurrent requests
	numRequests := 5
	results := make(chan int, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			token := createValidJWT(t, testUserID)
			body := mustJSON(&request.CreateSessionRequest{Mode: "single", PlayerCount: 1})
			req := httptest.NewRequest(http.MethodPost, "/session", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/session", h.CreateSession)
			r.ServeHTTP(w, req)

			results <- w.Code
		}()
	}

	// Collect results
	successCount := 0
	conflictCount := 0

	for i := 0; i < numRequests; i++ {
		code := <-results
		if code == http.StatusCreated {
			successCount++
		} else if code == http.StatusConflict {
			conflictCount++
		}
	}

	// Only one request should succeed
	assert.Equal(t, 1, successCount, "Only one session should be created")
	assert.Equal(t, numRequests-1, conflictCount, "Other requests should get conflict error")
}