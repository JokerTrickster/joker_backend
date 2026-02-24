package handler

import (
	"net/http"

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
		return errorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return successResponse(c, http.StatusOK, map[string]interface{}{
		"quests": resp,
	})
}

func (h *TDQuestHandler) UpdateProgress(c echo.Context) error {
	ctx := c.Request().Context()
	
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	questID, err := getParamUint(c, "id")
	if err != nil {
		return err
	}

	var req request.UpdateQuestProgressRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	resp, err := h.uc.UpdateQuestProgress(ctx, userID, questID, &req)
	if err != nil {
		return h.handleQuestError(c, err)
	}

	return successResponse(c, http.StatusOK, resp)
}

func (h *TDQuestHandler) ClaimReward(c echo.Context) error {
	ctx := c.Request().Context()
	
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	questID, err := getParamUint(c, "id")
	if err != nil {
		return err
	}

	resp, err := h.uc.ClaimReward(ctx, userID, questID)
	if err != nil {
		return h.handleQuestError(c, err)
	}

	return successResponse(c, http.StatusOK, resp)
}

func (h *TDQuestHandler) handleQuestError(c echo.Context, err error) error {
	errMsg := err.Error()
	
	switch errMsg {
	case "quest not found":
		return errorResponse(c, http.StatusNotFound, errMsg)
	case "quest not completed":
		return errorResponse(c, http.StatusBadRequest, errMsg)
	default:
		return errorResponse(c, http.StatusInternalServerError, errMsg)
	}
}
