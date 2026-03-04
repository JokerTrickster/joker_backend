package handler

import (
	"net/http"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/product/model/request"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/product/usecase"
	"github.com/labstack/echo/v4"
)

type CreateHandler struct {
	UseCase *usecase.CreateUseCase
}

func NewCreateHandler(uc *usecase.CreateUseCase) *CreateHandler {
	return &CreateHandler{UseCase: uc}
}

func (h *CreateHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	req := &request.ReqCreateProduct{}
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "BAD_REQUEST")
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	res, err := h.UseCase.Create(ctx, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "INTERNAL_ERROR")
	}
	return c.JSON(http.StatusCreated, res)
}
