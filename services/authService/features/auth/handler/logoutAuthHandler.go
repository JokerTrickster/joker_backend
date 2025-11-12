package handler

import (
	"context"
	"net/http"

	_interface "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/interface"

	"github.com/labstack/echo/v4"
)

type LogoutAuthHandler struct {
	UseCase _interface.ILogoutAuthUseCase
}

func NewLogoutAuthHandler(c *echo.Echo, useCase _interface.ILogoutAuthUseCase) _interface.ILogoutAuthHandler {
	handler := &LogoutAuthHandler{
		UseCase: useCase,
	}
	c.POST("/v0.1/auth/logout", handler.Logout)
	return handler
}

// 로그 아웃
// @Router /v0.1/auth/logout [post]
// @Summary 로그 아웃
// @Description
// @Description ■ errCode with 400
// @Description PARAM_BAD : 파라미터 오류
// @Description TOKEN_INVALID : 토큰이 유효하지 않음
// @Description TOKEN_EXPIRED : 토큰이 만료됨
// @Description
// @Description ■ errCode with 500
// @Description INTERNAL_SERVER : 내부 로직 처리 실패
// @Produce json
// @Success 200 {object} bool
// @Failure 400 {object} error
// @Failure 500 {object} error
// @Tags auth
func (d *LogoutAuthHandler) Logout(c echo.Context) error {
	ctx := context.Background()
	//토큰에서 유저 정보 추출
	uID := uint(0) //추출한 유저ID로 변경
	err := d.UseCase.Logout(ctx, uID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, true)
}
