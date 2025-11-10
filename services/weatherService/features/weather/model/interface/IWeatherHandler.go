package _interface

import "github.com/labstack/echo/v4"

type IRegisterAlarmWeatherHandler interface {
	RegisterAlarm(c echo.Context) error
}
