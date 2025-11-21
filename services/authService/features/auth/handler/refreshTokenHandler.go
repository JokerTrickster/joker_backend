package handler

import (
	"context"
	"net/http"

	_interface "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/interface"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/request"
	_ "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/response"
	"github.com/labstack/echo/v4"
)

type RefreshTokenHandler struct {
	UseCase _interface.IRefreshTokenUseCase
}

func NewRefreshTokenHandler(c *echo.Echo, useCase _interface.IRefreshTokenUseCase) _interface.IRefreshTokenHandler {
	handler := &RefreshTokenHandler{
		UseCase: useCase,
	}
	c.POST("/v0.1/auth/refresh", handler.RefreshToken)
	return handler
}

// 토큰 재발급
// @Router /v0.1/auth/refresh [post]
// @Summary 토큰 재발급
// @Description
// @Description ■ errCode with 400
// @Description PARAM_BAD : 파라미터 오류
// @Description TOKEN_INVALID : 토큰이 유효하지 않음
// @Description TOKEN_EXPIRED : 토큰이 만료됨
// @Description
// @Description ■ errCode with 500
// @Description INTERNAL_SERVER : 내부 로직 처리 실패
// @Param json body request.ReqRefreshToken true "리프레시 토큰"
// @Produce json
// @Success 200 {object} response.ResRefreshToken
// @Failure 400 {object} error
// @Failure 500 {object} error
// @Tags auth
func (d *RefreshTokenHandler) RefreshToken(c echo.Context) error {
	ctx := context.Background()
	req := &request.ReqRefreshToken{}
	if err := c.Bind(req); err != nil {
		return err
	}
	if err := c.Validate(req); err != nil {
		return err
	}
	res, err := d.UseCase.RefreshToken(ctx, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}
