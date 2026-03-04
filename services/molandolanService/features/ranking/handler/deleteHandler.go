package handler

import (
	"net/http"
	"strconv"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/ranking/usecase"
	"github.com/labstack/echo/v4"
)

type DeleteHandler struct {
	UseCase *usecase.DeleteUseCase
}

func NewDeleteHandler(uc *usecase.DeleteUseCase) *DeleteHandler {
	return &DeleteHandler{UseCase: uc}
}

func (h *DeleteHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "INVALID_ID")
	}

	if err := h.UseCase.Delete(ctx, uint(id)); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "NOT_FOUND")
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "삭제되었습니다."})
}
