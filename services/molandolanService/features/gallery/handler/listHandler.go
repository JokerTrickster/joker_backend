package handler

import (
	"net/http"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/request"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/usecase"
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
	req := &request.ReqListGallery{}
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "BAD_REQUEST")
	}

	var userID *uint
	if uid, ok := c.Get("userID").(uint); ok {
		userID = &uid
	}

	res, err := h.UseCase.List(ctx, req, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "INTERNAL_ERROR")
	}
	return c.JSON(http.StatusOK, res)
}
