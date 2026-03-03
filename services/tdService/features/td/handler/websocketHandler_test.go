package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/tdService/pkg/websocket"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestWebSocketHandler_HandleWebSocket_MissingSessionID(t *testing.T) {
	t.Log("WebSocketHandler.HandleWebSocket: no sessionId param -> 400")
	gin.SetMode(gin.TestMode)

	hub := websocket.NewHub()
	mockPlayer := new(MockPlayerRepository)
	authUC := newTestAuthUseCase(mockPlayer)
	h := NewWebSocketHandler(hub, authUC)

	req := httptest.NewRequest(http.MethodGet, "/ws/game/", nil)
	req.URL.RawQuery = "token=" + createValidJWT(t, testUserID)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// Force empty sessionId param - Gin's router returns 404 for /ws/game/
	c.Params = gin.Params{{Key: "sessionId", Value: ""}}
	c.Request = req
	h.HandleWebSocket(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "session ID required")
}

func TestWebSocketHandler_HandleWebSocket_MissingToken(t *testing.T) {
	t.Log("WebSocketHandler.HandleWebSocket: no query token -> 401")
	gin.SetMode(gin.TestMode)

	hub := websocket.NewHub()
	mockPlayer := new(MockPlayerRepository)
	authUC := newTestAuthUseCase(mockPlayer)
	h := NewWebSocketHandler(hub, authUC)

	req := httptest.NewRequest(http.MethodGet, "/ws/game/"+testSessionID, nil)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/ws/game/:sessionId", h.HandleWebSocket)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "token required")
}

func TestWebSocketHandler_HandleWebSocket_InvalidToken(t *testing.T) {
	t.Log("WebSocketHandler.HandleWebSocket: bad token -> 401")
	gin.SetMode(gin.TestMode)

	hub := websocket.NewHub()
	mockPlayer := new(MockPlayerRepository)
	authUC := newTestAuthUseCase(mockPlayer)
	h := NewWebSocketHandler(hub, authUC)

	req := httptest.NewRequest(http.MethodGet, "/ws/game/"+testSessionID, nil)
	req.URL.RawQuery = "token=invalid-jwt-token"

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/ws/game/:sessionId", h.HandleWebSocket)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid token")
}
