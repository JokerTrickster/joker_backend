package handler

import (
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/repository"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/usecase"

	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
	"github.com/labstack/echo/v4"
)

func NewWeatherHandler(c *echo.Echo) {
	NewRegisterAlarmWeatherHandler(c, usecase.NewRegisterAlarmWeatherUseCase(repository.NewRegisterAlarmWeatherRepository(mysql.GormMysqlDB), mysql.DBTimeOut))
}
