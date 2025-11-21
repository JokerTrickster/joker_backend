package handler

import (
	"context"
	"net/http"

	_interface "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/interface"
	_ "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/response"
	"github.com/labstack/echo/v4"
)

type GoogleSigninAuthHandler struct {
	UseCase _interface.IGoogleSigninAuthUseCase
}

func NewGoogleSigninAuthHandler(c *echo.Echo, useCase _interface.IGoogleSigninAuthUseCase) _interface.IGoogleSigninAuthHandler {
	handler := &GoogleSigninAuthHandler{
		UseCase: useCase,
	}
	c.POST("/v0.1/auth/google/signin", handler.GoogleSignin)
	return handler
}

// 구글 로그인
// @Router /v0.1/auth/google/signin [post]
// @Summary 구글 로그인
// @Description
// @Description ■ errCode with 400
// @Description PARAM_BAD : 파라미터 오류
// @Description USER_NOT_EXIST : 유저가 존재하지 않음
// @Description USER_ALREADY_EXISTED : 유저가 이미 존재
// @Description USER_GOOGLE_ALREADY_EXISTED : 구글 계정이 이미 존재
// @Description PASSWORD_NOT_MATCH : 비밀번호가 일치하지 않음
// @Description
// @Description ■ errCode with 500
// @Description INTERNAL_SERVER : 내부 로직 처리 실패
// @Description INTERNAL_DB : DB 처리 실패
// @Produce json
// @Success 200 {object} response.ResGoogleSignin
// @Failure 400 {object} error
// @Failure 500 {object} error
// @Tags auth
func (d *GoogleSigninAuthHandler) GoogleSignin(c echo.Context) error {
	ctx := context.Background()

	res, err := d.UseCase.GoogleSignin(ctx)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}
