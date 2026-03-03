package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/entity"
	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/request"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var testPlayer = &entity.Player{
	ID:         testPlayerID,
	UserID:     testUserID,
	Nickname:   "testuser",
	AvatarID:   "default_avatar",
	Level:      1,
	Experience: 0,
}

var testSession = &entity.GameSession{
	ID:          1,
	SessionID:    testSessionID,
	RoomCode:     testRoomCode,
	Mode:         entity.GameModeSingle,
	Status:       entity.GameStatusPlaying,
	PlayerCount:  1,
	PlayerID:     testPlayerID,
	Player:       testPlayer,
}

func TestGameHandler_CreateSession_MissingAuth(t *testing.T) {
	t.Log("GameHandler.CreateSession: no auth header -> 401")
	gin.SetMode(gin.TestMode)

	mockPlayer := new(MockPlayerRepository)
	mockSession := new(MockGameSessionRepository)
	authUC := newTestAuthUseCase(mockPlayer)
	gameUC := newTestGameUseCase(mockSession, mockPlayer)
	h := NewGameHandler(gameUC, authUC)

	body := mustJSON(&request.CreateSessionRequest{Mode: "single", PlayerCount: 1})
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/session", h.CreateSession)
	req := httptest.NewRequest(http.MethodPost, "/session", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "error")
	mockPlayer.AssertNotCalled(t, "GetPlayerByUserID")
	mockSession.AssertNotCalled(t, "CreateSession")
}

func TestGameHandler_CreateSession_InvalidAuth(t *testing.T) {
	t.Log("GameHandler.CreateSession: bad Bearer token -> 401")
	gin.SetMode(gin.TestMode)

	mockPlayer := new(MockPlayerRepository)
	mockSession := new(MockGameSessionRepository)
	authUC := newTestAuthUseCase(mockPlayer)
	gameUC := newTestGameUseCase(mockSession, mockPlayer)
	h := NewGameHandler(gameUC, authUC)

	body := mustJSON(&request.CreateSessionRequest{Mode: "single", PlayerCount: 1})
	req := httptest.NewRequest(http.MethodPost, "/session", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer invalid-token")

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/session", h.CreateSession)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockPlayer.AssertNotCalled(t, "GetPlayerByUserID")
	mockSession.AssertNotCalled(t, "CreateSession")
}

func TestGameHandler_CreateSession_InvalidJSON(t *testing.T) {
	t.Log("GameHandler.CreateSession: valid auth, bad body -> 400")
	gin.SetMode(gin.TestMode)

	mockPlayer := new(MockPlayerRepository)
	mockSession := new(MockGameSessionRepository)
	authUC := newTestAuthUseCase(mockPlayer)
	gameUC := newTestGameUseCase(mockSession, mockPlayer)
	h := NewGameHandler(gameUC, authUC)

	token := createValidJWT(t, testUserID)
	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/session", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/session", h.CreateSession)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockPlayer.AssertNotCalled(t, "GetPlayerByUserID")
	mockSession.AssertNotCalled(t, "CreateSession")
}

func TestGameHandler_CreateSession_Success(t *testing.T) {
	t.Log("GameHandler.CreateSession: valid auth + body -> 200")
	gin.SetMode(gin.TestMode)

	mockPlayer := new(MockPlayerRepository)
	mockSession := new(MockGameSessionRepository)
	authUC := newTestAuthUseCase(mockPlayer)
	gameUC := newTestGameUseCase(mockSession, mockPlayer)
	h := NewGameHandler(gameUC, authUC)

	mockPlayer.On("GetPlayerByUserID", testUserID).Return(testPlayer, nil)
	mockSession.On("CreateSession", mock.AnythingOfType("*entity.GameSession")).
		Run(func(args mock.Arguments) {
			s := args.Get(0).(*entity.GameSession)
			s.SessionID = testSessionID
			s.RoomCode = testRoomCode
		}).
		Return(nil)

	token := createValidJWT(t, testUserID)
	body := mustJSON(&request.CreateSessionRequest{Mode: "single", PlayerCount: 1})
	req := httptest.NewRequest(http.MethodPost, "/session", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/session", h.CreateSession)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), testSessionID)
	assert.Contains(t, w.Body.String(), testWsBaseURL)
	mockPlayer.AssertExpectations(t)
	mockSession.AssertExpectations(t)
}

func TestGameHandler_JoinSession_Success(t *testing.T) {
	t.Log("GameHandler.JoinSession: valid -> 200")
	gin.SetMode(gin.TestMode)

	mockPlayer := new(MockPlayerRepository)
	mockSession := new(MockGameSessionRepository)
	authUC := newTestAuthUseCase(mockPlayer)
	gameUC := newTestGameUseCase(mockSession, mockPlayer)
	h := NewGameHandler(gameUC, authUC)

	mockPlayer.On("GetPlayerByUserID", testUserID).Return(testPlayer, nil)
	mockSession.On("GetSessionByRoomCode", testRoomCode).Return(testSession, nil)
	mockSession.On("JoinSession", testSessionID, testPlayerID).Return(nil)

	token := createValidJWT(t, testUserID)
	body := mustJSON(&request.JoinSessionRequest{RoomCode: testRoomCode})
	req := httptest.NewRequest(http.MethodPost, "/session/join", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/session/join", h.JoinSession)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), testSessionID)
	mockPlayer.AssertExpectations(t)
	mockSession.AssertExpectations(t)
}

func TestGameHandler_JoinSession_NotFound(t *testing.T) {
	t.Log("GameHandler.JoinSession: session not found -> 404")
	gin.SetMode(gin.TestMode)

	mockPlayer := new(MockPlayerRepository)
	mockSession := new(MockGameSessionRepository)
	authUC := newTestAuthUseCase(mockPlayer)
	gameUC := newTestGameUseCase(mockSession, mockPlayer)
	h := NewGameHandler(gameUC, authUC)

	mockPlayer.On("GetPlayerByUserID", testUserID).Return(testPlayer, nil)
	mockSession.On("GetSessionByRoomCode", "BADCODE").Return(nil, errors.New("session not found or already started"))

	token := createValidJWT(t, testUserID)
	body := mustJSON(&request.JoinSessionRequest{RoomCode: "BADCODE"})
	req := httptest.NewRequest(http.MethodPost, "/session/join", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/session/join", h.JoinSession)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockPlayer.AssertExpectations(t)
	mockSession.AssertExpectations(t)
}

func TestGameHandler_SaveResult_Success(t *testing.T) {
	t.Log("GameHandler.SaveResult: valid -> 200")
	gin.SetMode(gin.TestMode)

	mockPlayer := new(MockPlayerRepository)
	mockSession := new(MockGameSessionRepository)
	authUC := newTestAuthUseCase(mockPlayer)
	gameUC := newTestGameUseCase(mockSession, mockPlayer)
	h := NewGameHandler(gameUC, authUC)

	startedAt := time.Now().Add(-60 * time.Second)
	sessionWithStarted := &entity.GameSession{
		ID: 1, SessionID: testSessionID, RoomCode: testRoomCode,
		Mode: entity.GameModeSingle, Status: entity.GameStatusPlaying,
		PlayerCount: 1, PlayerID: testPlayerID, StartedAt: &startedAt,
	}

	mockPlayer.On("GetPlayerByUserID", testUserID).Return(testPlayer, nil)
	mockSession.On("GetSessionByID", testSessionID).Return(sessionWithStarted, nil)
	mockSession.On("SaveGameResult", mock.AnythingOfType("*entity.GameResult")).Return(nil)
	mockPlayer.On("UpdatePlayerStats", testPlayerID, mock.AnythingOfType("*entity.GameResult")).Return(nil)
	mockPlayer.On("UpdatePlayer", mock.AnythingOfType("*entity.Player")).Return(nil)
	mockSession.On("GetGameResults", testSessionID).Return([]*entity.GameResult{}, nil)

	token := createValidJWT(t, testUserID)
	body := mustJSON(&request.SaveResultRequest{
		SessionID: testSessionID, Score: 1000, WavesCompleted: 5,
		UnitsPlaced: 10, EnemiesDefeated: 50, Victory: true,
	})
	req := httptest.NewRequest(http.MethodPost, "/result", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/result", h.SaveResult)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	mockPlayer.AssertExpectations(t)
	mockSession.AssertExpectations(t)
}

func TestGameHandler_GetProfile_Success(t *testing.T) {
	t.Log("GameHandler.GetProfile: valid userId param -> 200")
	gin.SetMode(gin.TestMode)

	mockPlayer := new(MockPlayerRepository)
	mockSession := new(MockGameSessionRepository)
	authUC := newTestAuthUseCase(mockPlayer)
	gameUC := newTestGameUseCase(mockSession, mockPlayer)
	h := NewGameHandler(gameUC, authUC)

	mockPlayer.On("GetPlayerByUserID", testUserID).Return(testPlayer, nil)

	token := createValidJWT(t, testUserID)
	req := httptest.NewRequest(http.MethodGet, "/profile/42", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/profile/:userId", h.GetProfile)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "testuser")
	mockPlayer.AssertExpectations(t)
}

func TestGameHandler_GetProfile_MissingUserId(t *testing.T) {
	t.Log("GameHandler.GetProfile: empty param -> 400")
	gin.SetMode(gin.TestMode)

	mockPlayer := new(MockPlayerRepository)
	mockSession := new(MockGameSessionRepository)
	authUC := newTestAuthUseCase(mockPlayer)
	gameUC := newTestGameUseCase(mockSession, mockPlayer)
	h := NewGameHandler(gameUC, authUC)

	token := createValidJWT(t, testUserID)
	req := httptest.NewRequest(http.MethodGet, "/profile/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// Force empty userId param - Gin's router returns 404 for /profile/ so we set param directly
	c.Params = gin.Params{{Key: "userId", Value: ""}}
	c.Request = req
	h.GetProfile(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "user ID required")
	mockPlayer.AssertNotCalled(t, "GetPlayerByUserID")
}

func TestGameHandler_GetSessionInfo_Success(t *testing.T) {
	t.Log("GameHandler.GetSessionInfo: valid -> 200")
	gin.SetMode(gin.TestMode)

	mockPlayer := new(MockPlayerRepository)
	mockSession := new(MockGameSessionRepository)
	authUC := newTestAuthUseCase(mockPlayer)
	gameUC := newTestGameUseCase(mockSession, mockPlayer)
	h := NewGameHandler(gameUC, authUC)

	mockSession.On("GetSessionByID", testSessionID).Return(testSession, nil)

	token := createValidJWT(t, testUserID)
	req := httptest.NewRequest(http.MethodGet, "/session/"+testSessionID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/session/:sessionId", h.GetSessionInfo)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), testSessionID)
	mockSession.AssertExpectations(t)
}

func TestGameHandler_GetSessionInfo_NotFound(t *testing.T) {
	t.Log("GameHandler.GetSessionInfo: not found -> 404")
	gin.SetMode(gin.TestMode)

	mockPlayer := new(MockPlayerRepository)
	mockSession := new(MockGameSessionRepository)
	authUC := newTestAuthUseCase(mockPlayer)
	gameUC := newTestGameUseCase(mockSession, mockPlayer)
	h := NewGameHandler(gameUC, authUC)

	mockSession.On("GetSessionByID", "nonexistent").Return(nil, errors.New("session not found"))

	token := createValidJWT(t, testUserID)
	req := httptest.NewRequest(http.MethodGet, "/session/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/session/:sessionId", h.GetSessionInfo)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "error")
	mockSession.AssertExpectations(t)
}
