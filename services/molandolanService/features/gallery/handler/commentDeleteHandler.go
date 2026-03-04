package handler

import (
	"net/http"
	"strconv"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/usecase"
	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
	"github.com/JokerTrickster/joker_backend/shared/utils"
	"github.com/labstack/echo/v4"
)

type CommentDeleteHandler struct {
	UseCase *usecase.CommentDeleteUseCase
}

func NewCommentDeleteHandler(uc *usecase.CommentDeleteUseCase) *CommentDeleteHandler {
	return &CommentDeleteHandler{UseCase: uc}
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

	var role string
	mysql.GormMysqlDB.Raw("SELECT role FROM morandoran_users WHERE id = ? AND deleted_at IS NULL", userID).Scan(&role)

	if err := h.UseCase.Delete(ctx, uint(commentID), userID, role); err != nil {
		if err.Error() == "FORBIDDEN" {
			return echo.NewHTTPError(http.StatusForbidden, "FORBIDDEN")
		}
		return echo.NewHTTPError(http.StatusNotFound, "NOT_FOUND")
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "삭제되었습니다."})
}
