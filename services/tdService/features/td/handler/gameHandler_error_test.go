package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/entity"
	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/request"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func ptrUint(n uint) *uint { return &n }

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
				mp.On("GetPlayerByUserID", testUserID).Return(testPlayer, nil)
				ms.On("CreateSession", mock.AnythingOfType("*entity.GameSession")).Return(errors.New("invalid game mode"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "invalid game mode",
		},
		{
			name: "Player count exceeds maximum",
			request: request.CreateSessionRequest{
				Mode:        "single",
				PlayerCount: 10,
			},
			setupMocks: func(mp *MockPlayerRepository, ms *MockGameSessionRepository) {
				mp.On("GetPlayerByUserID", testUserID).Return(testPlayer, nil)
				ms.On("CreateSession", mock.AnythingOfType("*entity.GameSession")).Return(errors.New("player count exceeds maximum"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "player count exceeds maximum",
		},
		{
			name: "Player not found",
			request: request.CreateSessionRequest{
				Mode:        "single",
				PlayerCount: 1,
			},
			setupMocks: func(mp *MockPlayerRepository, ms *MockGameSessionRepository) {
				mp.On("GetPlayerByUserID", testUserID).Return(nil, gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "not found",
		},
		{
			name: "Database connection error",
			request: request.CreateSessionRequest{
				Mode:        "single",
				PlayerCount: 1,
			},
			setupMocks: func(mp *MockPlayerRepository, ms *MockGameSessionRepository) {
				mp.On("GetPlayerByUserID", testUserID).Return(nil, errors.New("database connection lost"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "database connection lost",
		},
		{
			name: "Session creation fails",
			request: request.CreateSessionRequest{
				Mode:        "single",
				PlayerCount: 1,
			},
			setupMocks: func(mp *MockPlayerRepository, ms *MockGameSessionRepository) {
				mp.On("GetPlayerByUserID", testUserID).Return(testPlayer, nil)
				ms.On("CreateSession", mock.AnythingOfType("*entity.GameSession")).Return(errors.New("failed to create session"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to create session",
		},
		{
			name: "Session creation limit error",
			request: request.CreateSessionRequest{
				Mode:        "single",
				PlayerCount: 1,
			},
			setupMocks: func(mp *MockPlayerRepository, ms *MockGameSessionRepository) {
				mp.On("GetPlayerByUserID", testUserID).Return(testPlayer, nil)
				ms.On("CreateSession", mock.AnythingOfType("*entity.GameSession")).Return(errors.New("session limit reached"))
			},
			expectedStatus: http.StatusInternalServerError,
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
				mp.On("GetPlayerByUserID", testUserID).Return(testPlayer, nil)
				ms.On("GetSessionByRoomCode", "").Return(nil, errors.New("room code is required"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "room code is required",
		},
		{
			name:     "Invalid room code format",
			roomCode: "INVALID!@#",
			setupMocks: func(mp *MockPlayerRepository, ms *MockGameSessionRepository) {
				mp.On("GetPlayerByUserID", testUserID).Return(testPlayer, nil)
				ms.On("GetSessionByRoomCode", "INVALID!@#").Return(nil, errors.New("invalid room code format"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "invalid room code format",
		},
		{
			name:     "Room not found",
			roomCode: "ROOM123",
			setupMocks: func(mp *MockPlayerRepository, ms *MockGameSessionRepository) {
				mp.On("GetPlayerByUserID", testUserID).Return(testPlayer, nil)
				ms.On("GetSessionByRoomCode", "ROOM123").Return(nil, gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "not found",
		},
		{
			name:     "Room is full",
			roomCode: "ROOM123",
			setupMocks: func(mp *MockPlayerRepository, ms *MockGameSessionRepository) {
				fullSession := &entity.GameSession{
					SessionID:   "sess-full",
					RoomCode:    "ROOM123",
					Mode:        entity.GameModeCoop,
					PlayerCount: 2,
					PlayerID:    testPlayerID,
					Player2ID:   ptrUint(2),
					Status:      entity.GameStatusWaiting,
				}
				mp.On("GetPlayerByUserID", testUserID).Return(testPlayer, nil)
				ms.On("GetSessionByRoomCode", "ROOM123").Return(fullSession, nil)
				ms.On("JoinSession", "sess-full", testPlayerID).Return(errors.New("session is already full"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "already full",
		},
		{
			name:     "Game already started",
			roomCode: "ROOM123",
			setupMocks: func(mp *MockPlayerRepository, ms *MockGameSessionRepository) {
				startedSession := &entity.GameSession{
					SessionID:   "sess-started",
					RoomCode:    "ROOM123",
					Mode:        entity.GameModeCoop,
					PlayerCount: 1,
					PlayerID:    testPlayerID,
					Status:      entity.GameStatusPlaying,
				}
				mp.On("GetPlayerByUserID", testUserID).Return(testPlayer, nil)
				ms.On("GetSessionByRoomCode", "ROOM123").Return(startedSession, nil)
				ms.On("JoinSession", "sess-started", testPlayerID).Return(errors.New("session already started"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "already started",
		},
		{
			name:     "Join session fails with conflict",
			roomCode: "ROOM123",
			setupMocks: func(mp *MockPlayerRepository, ms *MockGameSessionRepository) {
				sess := &entity.GameSession{SessionID: "sess-conflict", RoomCode: "ROOM123", Mode: entity.GameModeCoop, PlayerCount: 1, PlayerID: testPlayerID, Status: entity.GameStatusWaiting}
				mp.On("GetPlayerByUserID", testUserID).Return(testPlayer, nil)
				ms.On("GetSessionByRoomCode", "ROOM123").Return(sess, nil)
				ms.On("JoinSession", "sess-conflict", testPlayerID).Return(errors.New("already in another session"))
			},
			expectedStatus: http.StatusInternalServerError,
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
			body := mustJSON(map[string]string{"roomCode": tt.roomCode})
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

// TestGameHandler_ConcurrentSessionCreation tests concurrent session creation
func TestGameHandler_ConcurrentSessionCreation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockPlayer := new(MockPlayerRepository)
	mockSession := new(MockGameSessionRepository)

	mockPlayer.On("GetPlayerByUserID", testUserID).Return(testPlayer, nil)
	mockSession.On("CreateSession", mock.AnythingOfType("*entity.GameSession")).Return(nil).Once()
	mockSession.On("CreateSession", mock.AnythingOfType("*entity.GameSession")).Return(errors.New("duplicate session"))

	authUC := newTestAuthUseCase(mockPlayer)
	gameUC := newTestGameUseCase(mockSession, mockPlayer)
	h := NewGameHandler(gameUC, authUC)

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

	successCount := 0
	errorCount := 0
	for i := 0; i < numRequests; i++ {
		code := <-results
		if code == http.StatusOK {
			successCount++
		} else if code == http.StatusInternalServerError {
			errorCount++
		}
	}

	assert.Equal(t, 1, successCount, "Only one session should be created")
	assert.Equal(t, numRequests-1, errorCount, "Other requests should get 500 error")
}