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
	"github.com/stretchr/testify/require"
)

func TestAuthHandler_Login_Success(t *testing.T) {
	t.Log("AuthHandler.Login: valid request -> 200 with token")
	gin.SetMode(gin.TestMode)

	mockPlayer := new(MockPlayerRepository)
	authUC := newTestAuthUseCase(mockPlayer)
	h := NewAuthHandler(authUC)

	player := &entity.Player{
		ID:       testPlayerID,
		UserID:   testUserID,
		Nickname: "testuser",
		AvatarID: "default_avatar",
		Level:    1,
		Experience: 0,
	}
	mockPlayer.On("GetOrCreatePlayer", mock.AnythingOfType("uint"), "testuser", "default_avatar").
		Return(player, nil)

	body := mustJSON(&request.LoginRequest{Username: "testuser", Password: "password123"})
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/login", h.Login)
	c.Request = httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "token")
	assert.Contains(t, w.Body.String(), "testuser")
	mockPlayer.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	t.Log("AuthHandler.Login: invalid body -> 400")
	gin.SetMode(gin.TestMode)

	mockPlayer := new(MockPlayerRepository)
	authUC := newTestAuthUseCase(mockPlayer)
	h := NewAuthHandler(authUC)

	body := `{"username": "test", invalid}`
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/login", h.Login)
	c.Request = httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "error")
	mockPlayer.AssertNotCalled(t, "GetOrCreatePlayer")
}

func TestAuthHandler_Login_RepoError(t *testing.T) {
	t.Log("AuthHandler.Login: repo returns error -> 500")
	gin.SetMode(gin.TestMode)

	mockPlayer := new(MockPlayerRepository)
	authUC := newTestAuthUseCase(mockPlayer)
	h := NewAuthHandler(authUC)

	mockPlayer.On("GetOrCreatePlayer", mock.AnythingOfType("uint"), "testuser", "default_avatar").
		Return(nil, errors.New("db error"))

	body := mustJSON(&request.LoginRequest{Username: "testuser", Password: "password123"})
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/login", h.Login)
	c.Request = httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "error")
	mockPlayer.AssertExpectations(t)
}

func TestAuthHandler_RefreshToken_Success(t *testing.T) {
	t.Log("AuthHandler.RefreshToken: valid refresh token -> 200")
	gin.SetMode(gin.TestMode)

	mockPlayer := new(MockPlayerRepository)
	authUC := newTestAuthUseCase(mockPlayer)
	h := NewAuthHandler(authUC)

	token := createValidJWT(t, testUserID)
	body := mustJSON(&request.RefreshTokenRequest{RefreshToken: token})

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/refresh", h.RefreshToken)
	c.Request = httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "token")
}

func TestAuthHandler_RefreshToken_InvalidJSON(t *testing.T) {
	t.Log("AuthHandler.RefreshToken: invalid body -> 400")
	gin.SetMode(gin.TestMode)

	mockPlayer := new(MockPlayerRepository)
	authUC := newTestAuthUseCase(mockPlayer)
	h := NewAuthHandler(authUC)

	body := `{invalid}`
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/refresh", h.RefreshToken)
	c.Request = httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_RefreshToken_InvalidToken(t *testing.T) {
	t.Log("AuthHandler.RefreshToken: bad token -> 401")
	gin.SetMode(gin.TestMode)

	mockPlayer := new(MockPlayerRepository)
	authUC := newTestAuthUseCase(mockPlayer)
	h := NewAuthHandler(authUC)

	body := mustJSON(&request.RefreshTokenRequest{RefreshToken: "invalid-jwt-token"})

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/refresh", h.RefreshToken)
	c.Request = httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, c.Request)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "error")
}
