package handler

import (
	"net/http"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/ranking/usecase"
	"github.com/JokerTrickster/joker_backend/shared/utils"
	"github.com/labstack/echo/v4"
)

type MeHandler struct {
	UseCase *usecase.MeUseCase
}

func NewMeHandler(uc *usecase.MeUseCase) *MeHandler {
	return &MeHandler{UseCase: uc}
}

func (h *MeHandler) MyRanking(c echo.Context) error {
	ctx := c.Request().Context()
	gameType := c.Param("gameType")
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "UNAUTHORIZED")
	}

	res, err := h.UseCase.MyRanking(ctx, userID, gameType)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "NOT_FOUND")
	}
	return c.JSON(http.StatusOK, res)
}
