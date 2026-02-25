package handler

import (
	"net/http"

	_interface "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/interface"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/request"
	"github.com/labstack/echo/v4"
)

// TDCoopStateHandler handles co-op game state sync via REST polling
type TDCoopStateHandler struct {
	uc _interface.ITDRoomUseCase
}

// NewTDCoopStateHandler registers co-op state sync routes
func NewTDCoopStateHandler(g *echo.Group, uc _interface.ITDRoomUseCase) {
	h := &TDCoopStateHandler{uc: uc}
	
	// REST polling endpoints for simple co-op sync
	g.POST("/coop/rooms/:id/state", h.UpdateState)
	g.GET("/coop/rooms/:id/opponent-state", h.GetOpponentState)
}

// UpdateState updates the current player's game state
func (h *TDCoopStateHandler) UpdateState(c echo.Context) error {
	ctx := c.Request().Context()
	
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	roomID, err := getParamUint(c, "id")
	if err != nil {
		return err
	}

	var req request.UpdateGameStateRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	// Update player state in room
	if err := h.uc.UpdatePlayerState(ctx, roomID, userID, &req); err != nil {
		return errorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return successResponse(c, http.StatusOK, map[string]string{
		"status": "updated",
	})
}

// GetOpponentState returns the opponent's current game state
func (h *TDCoopStateHandler) GetOpponentState(c echo.Context) error {
	ctx := c.Request().Context()
	
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	roomID, err := getParamUint(c, "id")
	if err != nil {
		return err
	}

	// Get opponent's state
	state, err := h.uc.GetOpponentState(ctx, roomID, userID)
	if err != nil {
		if err.Error() == "opponent not found" {
			return errorResponse(c, http.StatusNotFound, "opponent not in room")
		}
		return errorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return successResponse(c, http.StatusOK, state)
}
