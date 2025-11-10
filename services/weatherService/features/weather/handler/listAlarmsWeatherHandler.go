package handler

import (
	"context"
	"net/http"

	_interface "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/interface"
	"github.com/JokerTrickster/joker_backend/shared/middleware"
	"github.com/labstack/echo/v4"
)

type ListAlarmsWeatherHandler struct {
	UseCase _interface.IListAlarmsWeatherUseCase
}

func NewListAlarmsWeatherHandler(c *echo.Echo, useCase _interface.IListAlarmsWeatherUseCase) _interface.IListAlarmsWeatherHandler {
	handler := &ListAlarmsWeatherHandler{
		UseCase: useCase,
	}
	c.GET("/v0.1/weather/alarms", handler.ListAlarms, middleware.JWTAuth())
	return handler
}

// 날씨 알람 목록 조회
// @Router /v0.1/weather/alarms [get]
// @Summary 날씨 알람 목록 조회
// @Description 사용자의 모든 날씨 알람 목록을 조회합니다. JWT 토큰 인증이 필요합니다.
// @Description
// @Description ■ errCode with 401
// @Description TOKEN_INVALID : 토큰이 유효하지 않음
// @Description TOKEN_EXPIRED : 토큰이 만료됨
// @Description
// @Description ■ errCode with 500
// @Description INTERNAL_SERVER : 내부 로직 처리 실패
// @Param Authorization header string true "Bearer {access_token}"
// @Produce json
// @Success 200 {object} response.ResListAlarms
// @Failure 401 {object} error
// @Failure 500 {object} error
// @Tags weather
func (d *ListAlarmsWeatherHandler) ListAlarms(c echo.Context) error {
	ctx := context.Background()

	// JWT 미들웨어에서 설정한 userID 추출
	userID, ok := c.Get("userID").(uint)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID in token")
	}

	// UseCase 호출
	res, err := d.UseCase.ListAlarms(ctx, int(userID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)
}
