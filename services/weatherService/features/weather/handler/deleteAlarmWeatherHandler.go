package handler

import (
	"context"
	"net/http"

	_interface "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/interface"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/request"
	"github.com/JokerTrickster/joker_backend/shared/middleware"
	"github.com/labstack/echo/v4"
)

type DeleteAlarmWeatherHandler struct {
	UseCase _interface.IDeleteAlarmWeatherUseCase
}

func NewDeleteAlarmWeatherHandler(c *echo.Echo, useCase _interface.IDeleteAlarmWeatherUseCase) _interface.IDeleteAlarmWeatherHandler {
	handler := &DeleteAlarmWeatherHandler{
		UseCase: useCase,
	}
	c.DELETE("/v0.1/weather/alarm", handler.DeleteAlarm, middleware.JWTAuth())
	return handler
}

// 날씨 알람 삭제
// @Router /v0.1/weather/alarm [delete]
// @Summary 날씨 알람 삭제
// @Description 사용자의 날씨 알람을 삭제합니다. JWT 토큰 인증이 필요합니다.
// @Description
// @Description ■ errCode with 400
// @Description PARAM_BAD : 파라미터 오류
// @Description
// @Description ■ errCode with 401
// @Description TOKEN_INVALID : 토큰이 유효하지 않음
// @Description TOKEN_EXPIRED : 토큰이 만료됨
// @Description
// @Description ■ errCode with 403
// @Description FORBIDDEN : 권한 없음 (다른 유저의 알람)
// @Description
// @Description ■ errCode with 404
// @Description NOT_FOUND : 알람을 찾을 수 없음
// @Description
// @Description ■ errCode with 500
// @Description INTERNAL_SERVER : 내부 로직 처리 실패
// @Param Authorization header string true "Bearer {access_token}"
// @Param json body request.ReqDeleteAlarm true "삭제할 알람 ID"
// @Produce json
// @Success 200 {object} bool
// @Failure 400 {object} error
// @Failure 401 {object} error
// @Failure 403 {object} error
// @Failure 404 {object} error
// @Failure 500 {object} error
// @Tags weather
func (d *DeleteAlarmWeatherHandler) DeleteAlarm(c echo.Context) error {
	ctx := context.Background()

	// JWT 미들웨어에서 설정한 userID 추출
	userID, ok := c.Get("userID").(uint)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID in token")
	}

	// Request body parsing
	req := &request.ReqDeleteAlarm{}
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Validation
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// UseCase 호출
	err := d.UseCase.DeleteAlarm(ctx, int(userID), req)
	if err != nil {
		if err.Error() == "alarm not found or you don't have permission" {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, true)
}
