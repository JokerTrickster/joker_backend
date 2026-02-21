package handler

import (
	"net/http"
	"strconv"

	_interface "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/interface"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/request"
	"github.com/labstack/echo/v4"
)

type RoundHandler struct {
	uc _interface.IGameRoundUseCase
}

func NewRoundHandler(g *echo.Group, uc _interface.IGameRoundUseCase) {
	h := &RoundHandler{uc: uc}
	g.POST("/rounds", h.StartRound)
	g.PATCH("/rounds/:id/end", h.EndRound)
	g.GET("/rounds", h.GetMyRounds)
	g.GET("/rounds/:id", h.GetRound)
}

func (h *RoundHandler) StartRound(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}
	resp, err := h.uc.StartRound(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, resp)
}

func (h *RoundHandler) EndRound(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}
	idStr := c.Param("id")
	roundID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid round id"})
	}
	var req request.EndRoundRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	resp, err := h.uc.EndRound(ctx, userID, uint(roundID), &req)
	if err != nil {
		if err.Error() == "round not found or not owned by you" || err.Error() == "round is already completed" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *RoundHandler) GetMyRounds(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}
	limit := 20
	if l := c.QueryParam("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	resp, err := h.uc.GetMyRounds(ctx, userID, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *RoundHandler) GetRound(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}
	idStr := c.Param("id")
	roundID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid round id"})
	}
	resp, err := h.uc.GetRound(ctx, userID, uint(roundID))
	if err != nil {
		if err.Error() == "round not found or not owned by you" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, resp)
}
