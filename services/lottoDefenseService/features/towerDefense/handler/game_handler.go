package handler

import (
	"net/http"

	_interface "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/interface"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/request"
	"github.com/labstack/echo/v4"
)

type TDGameHandler struct {
	uc _interface.ITDGameUseCase
}

func NewTDGameHandler(g *echo.Group, uc _interface.ITDGameUseCase) {
	h := &TDGameHandler{uc: uc}
	g.POST("/game/single/result", h.SaveSingleResult)
	g.GET("/game/history", h.GetHistory)
	g.GET("/rankings/:mode", h.GetRankings)
}

func (h *TDGameHandler) SaveSingleResult(c echo.Context) error {
	ctx := c.Request().Context()
	
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	var req request.SaveGameResultRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	resp, err := h.uc.SaveSingleResult(ctx, userID, &req)
	if err != nil {
		return errorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return successResponse(c, http.StatusCreated, resp)
}

func (h *TDGameHandler) GetHistory(c echo.Context) error {
	ctx := c.Request().Context()
	
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	req := request.GameHistoryRequest{
		GameMode: getQueryString(c, "mode", ""),
		Limit:    getQueryInt(c, "limit", 10),
		Offset:   getQueryInt(c, "offset", 0),
	}

	resp, err := h.uc.GetGameHistory(ctx, userID, &req)
	if err != nil {
		return errorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return successResponse(c, http.StatusOK, resp)
}

func (h *TDGameHandler) GetRankings(c echo.Context) error {
	ctx := c.Request().Context()

	gameMode := c.Param("mode")
	if !isValidGameMode(gameMode) {
		return errorResponse(c, http.StatusBadRequest, "invalid game mode (use 'single' or 'coop')")
	}

	resp, err := h.uc.GetWeeklyRankings(ctx, gameMode)
	if err != nil {
		return errorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return successResponse(c, http.StatusOK, resp)
}

func isValidGameMode(mode string) bool {
	return mode == "single" || mode == "coop"
}
