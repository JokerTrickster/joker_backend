package websocket

import (
	"net/http"
	"os"
	"strings"

	"github.com/JokerTrickster/joker_backend/shared/jwt"
	"github.com/JokerTrickster/joker_backend/shared/logger"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var allowedOrigins []string

func init() {
	env := os.Getenv("ENV")
	origins := os.Getenv("CORS_ALLOWED_ORIGINS")

	if origins != "" {
		for _, o := range strings.Split(origins, ",") {
			allowedOrigins = append(allowedOrigins, strings.TrimSpace(o))
		}
	}

	isDev := env == "" || env == "development" || env == "dev"
	if isDev && len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		for _, allowed := range allowedOrigins {
			if allowed == "*" || allowed == origin {
				return true
			}
		}
		logger.Warn("WebSocket origin rejected", zap.String("origin", origin))
		return false
	},
}

// Handler handles WebSocket connections
type Handler struct {
	hub *Hub
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub) *Handler {
	return &Handler{
		hub: hub,
	}
}

// HandleWebSocket handles WebSocket upgrade requests
func (h *Handler) HandleWebSocket(c echo.Context) error {
	// Extract token from query parameter
	token := c.QueryParam("token")
	if token == "" {
		logger.Warn("WebSocket connection without token")
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Authentication token required",
		})
	}

	// Parse token to get userID
	userID, _, err := jwt.ParseToken(token)
	if err != nil {
		logger.Warn("Invalid WebSocket token", zap.Error(err))
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid authentication token",
		})
	}

	// Upgrade HTTP connection to WebSocket
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		logger.Error("Failed to upgrade WebSocket connection", zap.Error(err))
		return err
	}

	// Create and register client
	client := NewClient(h.hub, ws, userID)
	client.hub.register <- client

	logger.Info("WebSocket connection established",
		zap.Uint("user_id", userID),
	)

	// Start client goroutines
	client.Start()

	return nil
}

// RegisterRoutes registers WebSocket routes
func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.GET("/ws", h.HandleWebSocket)
}
