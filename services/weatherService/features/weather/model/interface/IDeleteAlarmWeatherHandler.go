package _interface

import "github.com/labstack/echo/v4"

type IDeleteAlarmWeatherHandler interface {
	DeleteAlarm(c echo.Context) error
}
