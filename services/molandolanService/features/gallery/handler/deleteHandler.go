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

type DeleteHandler struct {
	UseCase *usecase.DeleteUseCase
	Repo    _interface.IGalleryRepository
}

func NewDeleteHandler(uc *usecase.DeleteUseCase, repo _interface.IGalleryRepository) *DeleteHandler {
	return &DeleteHandler{UseCase: uc, Repo: repo}
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

	role, _ := h.Repo.GetUserRole(ctx, userID)

	if err := h.UseCase.Delete(ctx, uint(id), userID, role); err != nil {
		if errors.Is(err, usecase.ErrForbidden) {
			return echo.NewHTTPError(http.StatusForbidden, "FORBIDDEN")
		}
		return echo.NewHTTPError(http.StatusNotFound, "NOT_FOUND")
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "삭제되었습니다."})
}
