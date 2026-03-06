package handler

import (
	"errors"
	"net/http"
	"strconv"

	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/interface"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/usecase"
	"github.com/JokerTrickster/joker_backend/shared/utils"
	"github.com/labstack/echo/v4"
)

type CommentDeleteHandler struct {
	UseCase *usecase.CommentDeleteUseCase
	Repo    _interface.IGalleryRepository
}

func NewCommentDeleteHandler(uc *usecase.CommentDeleteUseCase, repo _interface.IGalleryRepository) *CommentDeleteHandler {
	return &CommentDeleteHandler{UseCase: uc, Repo: repo}
}

func (h *CommentDeleteHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()
	commentID, err := strconv.ParseUint(c.Param("commentId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "INVALID_ID")
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "UNAUTHORIZED")
	}

	role, _ := h.Repo.GetUserRole(ctx, userID)

	if err := h.UseCase.Delete(ctx, uint(commentID), userID, role); err != nil {
		if errors.Is(err, usecase.ErrForbidden) {
			return echo.NewHTTPError(http.StatusForbidden, "FORBIDDEN")
		}
		return echo.NewHTTPError(http.StatusNotFound, "NOT_FOUND")
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "삭제되었습니다."})
}
