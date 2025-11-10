package handler

import (
	"context"
	"net/http"

	_interface "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/interface"
	"github.com/labstack/echo/v4"
)

type RegisterAlarmWeatherHandler struct {
	UseCase _interface.IRegisterAlarmWeatherUseCase
}

func NewRegisterAlarmWeatherHandler(c *echo.Echo, useCase _interface.IRegisterAlarmWeatherUseCase) _interface.IRegisterAlarmWeatherHandler {
	handler := &RegisterAlarmWeatherHandler{
		UseCase: useCase,
	}
	c.POST("/v0.1/weather/register", handler.RegisterAlarm)
	return handler
}

// 날씨 알람 등록
// @Router /v0.1/weather/register [post]
// @Summary 날씨 알람 등록
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
// @Success 200 {object} bool
// @Failure 400 {object} error
// @Failure 500 {object} error
// @Tags auth
func (d *RegisterAlarmWeatherHandler) RegisterAlarm(c echo.Context) error {
	ctx := context.Background()

	err := d.UseCase.RegisterAlarm(ctx)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, true)
}
