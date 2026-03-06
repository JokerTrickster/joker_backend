package handler

import (
	"net/http"
	"strconv"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/model/request"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/usecase"
	"github.com/labstack/echo/v4"
)

type UpdateHandler struct {
	UseCase *usecase.UpdateUseCase
}

func NewUpdateHandler(uc *usecase.UpdateUseCase) *UpdateHandler {
	return &UpdateHandler{UseCase: uc}
}

func (h *UpdateHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "INVALID_ID")
	}

	req := &request.ReqUpdateNews{}
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "BAD_REQUEST")
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "VALIDATION_ERROR")
	}

	res, err := h.UseCase.Update(ctx, uint(id), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "NOT_FOUND")
	}
	return c.JSON(http.StatusOK, res)
}
