package handler

import (
	"context"
	"net/http"

	_interface "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/interface"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/request"
	"github.com/JokerTrickster/joker_backend/shared/middleware"
	"github.com/labstack/echo/v4"
)

type RegisterAlarmWeatherHandler struct {
	UseCase _interface.IRegisterAlarmWeatherUseCase
}

func NewRegisterAlarmWeatherHandler(c *echo.Echo, useCase _interface.IRegisterAlarmWeatherUseCase) _interface.IRegisterAlarmWeatherHandler {
	handler := &RegisterAlarmWeatherHandler{
		UseCase: useCase,
	}
	c.POST("/v0.1/weather/register", handler.RegisterAlarm, middleware.JWTAuth())
	return handler
}

// 날씨 알람 등록
// @Router /v0.1/weather/register [post]
// @Summary 날씨 알람 등록
// @Description 사용자의 날씨 알람을 등록합니다. JWT 토큰 인증이 필요합니다.
// @Description
// @Description ■ errCode with 400
// @Description PARAM_BAD : 파라미터 오류
// @Description
// @Description ■ errCode with 401
// @Description TOKEN_INVALID : 토큰이 유효하지 않음
// @Description TOKEN_EXPIRED : 토큰이 만료됨
// @Description
// @Description ■ errCode with 500
// @Description INTERNAL_SERVER : 내부 로직 처리 실패
// @Param Authorization header string true "Bearer {access_token}"
// @Param json body request.ReqRegisterAlarm true "알람 등록 정보"
// @Produce json
// @Success 200 {object} response.ResRegisterAlarm
// @Failure 400 {object} error
// @Failure 401 {object} error
// @Failure 500 {object} error
// @Tags weather
func (d *RegisterAlarmWeatherHandler) RegisterAlarm(c echo.Context) error {
	ctx := context.Background()

	// JWT 미들웨어에서 설정한 userID 추출
	userID, ok := c.Get("userID").(uint)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID in token")
	}

	// Request body parsing
	req := &request.ReqRegisterAlarm{}
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Validation
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// UseCase 호출
	res, err := d.UseCase.RegisterAlarm(ctx, int(userID), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)
}
