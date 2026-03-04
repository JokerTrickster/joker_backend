package handler

import (
	"context"
	"fmt"
	"net/http"

	authEntity "github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/model/entity"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/ranking/model/request"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/ranking/usecase"
	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
	"github.com/JokerTrickster/joker_backend/shared/utils"
	"github.com/labstack/echo/v4"
)

type SubmitHandler struct {
	UseCase *usecase.SubmitUseCase
}

func NewSubmitHandler(uc *usecase.SubmitUseCase) *SubmitHandler {
	return &SubmitHandler{UseCase: uc}
}

func (h *SubmitHandler) Submit(c echo.Context) error {
	ctx := c.Request().Context()
	gameType := c.Param("gameType")
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "UNAUTHORIZED")
	}

	req := &request.ReqSubmitRanking{}
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "BAD_REQUEST")
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	nickname, err := getUserNickname(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "INTERNAL_ERROR")
	}

	res, err := h.UseCase.Submit(ctx, userID, nickname, gameType, req.ClearTimeMs)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "INTERNAL_ERROR")
	}
	return c.JSON(http.StatusOK, res)
}

func getUserNickname(ctx context.Context, userID uint) (string, error) {
	var user authEntity.MorandoranUser
	if err := mysql.GormMysqlDB.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return "", fmt.Errorf("user not found")
	}
	return user.Nickname, nil
}
