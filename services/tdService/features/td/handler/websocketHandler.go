package handler

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/usecase"
	"github.com/JokerTrickster/joker_backend/services/tdService/pkg/websocket"
	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
)

var allowedOrigins = parseAllowedOrigins()

func parseAllowedOrigins() map[string]bool {
	origins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if origins == "" {
		origins = "http://localhost:3000"
	}
	m := make(map[string]bool)
	for _, o := range strings.Split(origins, ",") {
		m[strings.TrimSpace(o)] = true
	}
	return m
}

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return allowedOrigins[origin]
	},
}

type WebSocketHandler struct {
	hub         *websocket.Hub
	authUseCase *usecase.AuthUseCase
}

func NewWebSocketHandler(hub *websocket.Hub, authUseCase *usecase.AuthUseCase) *WebSocketHandler {
	return &WebSocketHandler{
		hub:         hub,
		authUseCase: authUseCase,
	}
}

// HandleWebSocket handles WebSocket connections for game sessions
// @Summary Connect to game session via WebSocket
// @Description Establish WebSocket connection for real-time game communication
// @Tags TD-WebSocket
// @Param sessionId path string true "Session ID"
// @Param token query string true "JWT token for authentication"
// @Router /ws/game/{sessionId} [get]
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session ID required"})
		return
	}

	// Get token from query parameter (WebSocket doesn't support headers well)
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
		return
	}

	// Validate token and get user ID
	userID, err := h.authUseCase.ValidateTokenAndGetUserID(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Create client
	client := &websocket.Client{
		Hub:       h.hub,
		Conn:      conn,
		Send:      make(chan []byte, 256),
		SessionID: sessionID,
		UserID:    strconv.FormatUint(uint64(userID), 10),
		PlayerNum: h.getPlayerNumber(sessionID, userID),
	}

	// Register client
	h.hub.Register <- client

	// Start client pumps
	go client.WritePump()
	go client.ReadPump()
}

// getPlayerNumber determines the player number based on current session occupancy
func (h *WebSocketHandler) getPlayerNumber(sessionID string, userID uint) int {
	count := h.hub.GetRoomClientCount(sessionID)
	return count + 1
}