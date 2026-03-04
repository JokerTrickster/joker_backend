package handler

import (
	"net/http"
	"strconv"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/request"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/usecase"
	"github.com/labstack/echo/v4"
)

type CommentListHandler struct {
	UseCase *usecase.CommentListUseCase
}

func NewCommentListHandler(uc *usecase.CommentListUseCase) *CommentListHandler {
	return &CommentListHandler{UseCase: uc}
}

func (h *CommentListHandler) List(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "INVALID_ID")
	}

	req := &request.ReqListComment{}
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "BAD_REQUEST")
	}

	res, err := h.UseCase.List(ctx, uint(id), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "INTERNAL_ERROR")
	}
	return c.JSON(http.StatusOK, res)
}
