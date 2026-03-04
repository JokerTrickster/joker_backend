package handler

import (
	"net/http"
	"strings"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/model/request"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/usecase"
	"github.com/JokerTrickster/joker_backend/shared/utils"
	"github.com/labstack/echo/v4"
)

type UpdateMeHandler struct {
	UseCase *usecase.UpdateMeUseCase
}

func NewUpdateMeHandler(uc *usecase.UpdateMeUseCase) *UpdateMeHandler {
	return &UpdateMeHandler{UseCase: uc}
}

func (h *UpdateMeHandler) UpdateMe(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "UNAUTHORIZED")
	}

	req := &request.ReqUpdateMe{}
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "BAD_REQUEST")
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if strings.Contains(req.Nickname, " ") {
		return echo.NewHTTPError(http.StatusBadRequest, "닉네임에 공백을 포함할 수 없습니다.")
	}

	res, err := h.UseCase.UpdateNickname(ctx, userID, req.Nickname)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "NOT_FOUND")
	}
	return c.JSON(http.StatusOK, res)
}
