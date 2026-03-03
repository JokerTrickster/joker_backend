package handler

import (
	"net/http"
	"strconv"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/gallery/usecase"
	"github.com/JokerTrickster/joker_backend/shared/utils"
	"github.com/labstack/echo/v4"
)

type LikeHandler struct {
	UseCase *usecase.LikeUseCase
}

func NewLikeHandler(uc *usecase.LikeUseCase) *LikeHandler {
	return &LikeHandler{UseCase: uc}
}

func (h *LikeHandler) Like(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "INVALID_ID")
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "UNAUTHORIZED")
	}

	res, err := h.UseCase.ToggleLike(ctx, userID, uint(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "INTERNAL_ERROR")
	}
	return c.JSON(http.StatusOK, res)
}
