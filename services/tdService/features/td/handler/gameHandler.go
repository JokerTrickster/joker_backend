package handler

import (
	"net/http"
	"strings"

	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/request"
	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/usecase"
	"github.com/gin-gonic/gin"
)

type GameHandler struct {
	gameUseCase *usecase.GameUseCase
	authUseCase *usecase.AuthUseCase
}

func NewGameHandler(gameUseCase *usecase.GameUseCase, authUseCase *usecase.AuthUseCase) *GameHandler {
	return &GameHandler{
		gameUseCase: gameUseCase,
		authUseCase: authUseCase,
	}
}

// CreateSession creates a new game session
// @Summary Create a new game session
// @Description Create a new single or coop game session
// @Tags TD-Game
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Bearer token"
// @Param request body request.CreateSessionRequest true "Session creation data"
// @Success 200 {object} response.CreateSessionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/td/game/session [post]
func (h *GameHandler) CreateSession(c *gin.Context) {
	// Get user ID from token
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req request.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.gameUseCase.CreateSession(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// JoinSession joins an existing game session
// @Summary Join an existing game session
// @Description Join a coop game session using room code
// @Tags TD-Game
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Bearer token"
// @Param request body request.JoinSessionRequest true "Join session data"
// @Success 200 {object} response.CreateSessionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/td/game/session/join [post]
func (h *GameHandler) JoinSession(c *gin.Context) {
	// Get user ID from token
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req request.JoinSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.gameUseCase.JoinSession(userID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// SaveResult saves game result
// @Summary Save game result
// @Description Save the result of a completed game
// @Tags TD-Game
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Bearer token"
// @Param request body request.SaveResultRequest true "Game result data"
// @Success 200 {object} response.GameResultResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/td/game/result [post]
func (h *GameHandler) SaveResult(c *gin.Context) {
	// Get user ID from token
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req request.SaveResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.gameUseCase.SaveResult(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetProfile gets player profile
// @Summary Get player profile
// @Description Get player profile including stats
// @Tags TD-Profile
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Bearer token"
// @Param userId path string true "User ID"
// @Success 200 {object} response.ProfileResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/td/profile/{userId} [get]
func (h *GameHandler) GetProfile(c *gin.Context) {
	// Validate token
	_, err := h.getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user ID required"})
		return
	}

	resp, err := h.gameUseCase.GetProfile(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetSessionInfo gets session information
// @Summary Get session information
// @Description Get detailed information about a game session
// @Tags TD-Game
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Bearer token"
// @Param sessionId path string true "Session ID"
// @Success 200 {object} response.SessionInfoResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/td/game/session/{sessionId} [get]
func (h *GameHandler) GetSessionInfo(c *gin.Context) {
	// Validate token
	_, err := h.getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session ID required"})
		return
	}

	resp, err := h.gameUseCase.GetSessionInfo(sessionID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// getUserIDFromToken extracts user ID from JWT token
func (h *GameHandler) getUserIDFromToken(c *gin.Context) (uint, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return 0, ErrMissingAuthHeader
	}

	// Remove "Bearer " prefix
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		return 0, ErrInvalidAuthHeader
	}

	return h.authUseCase.ValidateTokenAndGetUserID(token)
}