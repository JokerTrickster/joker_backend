package handler

import (
	"net/http"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/ranking/model/request"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/ranking/usecase"
	"github.com/labstack/echo/v4"
)

type ListHandler struct {
	UseCase *usecase.ListUseCase
}

func NewListHandler(uc *usecase.ListUseCase) *ListHandler {
	return &ListHandler{UseCase: uc}
}

func (h *ListHandler) List(c echo.Context) error {
	ctx := c.Request().Context()
	gameType := c.Param("gameType")
	req := &request.ReqListRanking{}
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "BAD_REQUEST")
	}

	res, err := h.UseCase.List(ctx, gameType, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "INTERNAL_ERROR")
	}
	return c.JSON(http.StatusOK, res)
}
