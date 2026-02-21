package handler

import (
	"net/http"
	"strconv"

	_interface "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/interface"
	"github.com/labstack/echo/v4"
)

type LeaderboardHandler struct {
	uc _interface.ILeaderboardUseCase
}

func NewLeaderboardHandler(g *echo.Group, uc _interface.ILeaderboardUseCase) {
	h := &LeaderboardHandler{uc: uc}
	g.GET("/leaderboard", h.GetLeaderboard)
}

func (h *LeaderboardHandler) GetLeaderboard(c echo.Context) error {
	ctx := c.Request().Context()
	limit := 10
	if l := c.QueryParam("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	resp, err := h.uc.GetLeaderboard(ctx, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, resp)
}
