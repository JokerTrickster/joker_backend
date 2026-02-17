package handler

import (
	"net/http"
	"strconv"

	_interface "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/interface"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/request"
	"github.com/labstack/echo/v4"
)

type TDQuestHandler struct {
	uc _interface.ITDQuestUseCase
}

func NewTDQuestHandler(g *echo.Group, uc _interface.ITDQuestUseCase) {
	h := &TDQuestHandler{uc: uc}
	g.GET("/quests", h.GetActiveQuests)
	g.POST("/quests/:id/progress", h.UpdateProgress)
	g.POST("/quests/:id/claim", h.ClaimReward)
}

func (h *TDQuestHandler) GetActiveQuests(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	resp, err := h.uc.GetActiveQuests(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"quests": resp,
		},
	})
}

func (h *TDQuestHandler) UpdateProgress(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	idStr := c.Param("id")
	questID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid quest id"})
	}

	var req request.UpdateQuestProgressRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	resp, err := h.uc.UpdateQuestProgress(ctx, userID, uint(questID), &req)
	if err != nil {
		if err.Error() == "quest not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    resp,
	})
}

func (h *TDQuestHandler) ClaimReward(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	idStr := c.Param("id")
	questID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid quest id"})
	}

	resp, err := h.uc.ClaimReward(ctx, userID, uint(questID))
	if err != nil {
		if err.Error() == "quest not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		if err.Error() == "quest not completed" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    resp,
	})
}
