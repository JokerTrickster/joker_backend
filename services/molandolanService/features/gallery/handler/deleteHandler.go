package handler

import (
	"net/http"
	"strconv"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/gallery/usecase"
	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
	"github.com/JokerTrickster/joker_backend/shared/utils"
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

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "UNAUTHORIZED")
	}

	var role string
	mysql.GormMysqlDB.Raw("SELECT role FROM morandoran_users WHERE id = ? AND deleted_at IS NULL", userID).Scan(&role)

	if err := h.UseCase.Delete(ctx, uint(id), userID, role); err != nil {
		if err.Error() == "FORBIDDEN" {
			return echo.NewHTTPError(http.StatusForbidden, "FORBIDDEN")
		}
		return echo.NewHTTPError(http.StatusNotFound, "NOT_FOUND")
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "삭제되었습니다."})
}
