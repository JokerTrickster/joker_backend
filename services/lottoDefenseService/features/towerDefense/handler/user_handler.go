package handler

import (
	"net/http"

	_interface "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/interface"
	"github.com/labstack/echo/v4"
)

type TDUserHandler struct {
	authUC _interface.ITDAuthUseCase
	gameUC _interface.ITDGameUseCase
}

func NewTDUserHandler(g *echo.Group, authUC _interface.ITDAuthUseCase, gameUC _interface.ITDGameUseCase) {
	h := &TDUserHandler{
		authUC: authUC,
		gameUC: gameUC,
	}
	g.GET("/users/me", h.GetMe)
	g.GET("/users/me/stats", h.GetStats)
}

func (h *TDUserHandler) GetMe(c echo.Context) error {
	ctx := c.Request().Context()
	
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	resp, err := h.authUC.GetUserInfo(ctx, userID)
	if err != nil {
		return errorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return successResponse(c, http.StatusOK, resp)
}

func (h *TDUserHandler) GetStats(c echo.Context) error {
	ctx := c.Request().Context()
	
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	resp, err := h.gameUC.GetUserStats(ctx, userID)
	if err != nil {
		return errorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return successResponse(c, http.StatusOK, resp)
}
