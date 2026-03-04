package handler

import (
	"net/http"
	"strconv"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/usecase"
	"github.com/labstack/echo/v4"
)

type DetailHandler struct {
	UseCase *usecase.DetailUseCase
}

func NewDetailHandler(uc *usecase.DetailUseCase) *DetailHandler {
	return &DetailHandler{UseCase: uc}
}

func (h *DetailHandler) Detail(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "INVALID_ID")
	}

	res, err := h.UseCase.Detail(ctx, uint(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "NOT_FOUND")
	}
	return c.JSON(http.StatusOK, res)
}
