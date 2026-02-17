package handler

import (
	"net/http"
	"strconv"

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
}

func (h *TDGameHandler) SaveSingleResult(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	var req request.SaveGameResultRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	resp, err := h.uc.SaveSingleResult(ctx, userID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    resp,
	})
}

func (h *TDGameHandler) GetHistory(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	req := request.GameHistoryRequest{
		GameMode: c.QueryParam("mode"),
		Limit:    10,
		Offset:   0,
	}

	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = limit
		}
	}

	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			req.Offset = offset
		}
	}

	resp, err := h.uc.GetGameHistory(ctx, userID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    resp,
	})
}
