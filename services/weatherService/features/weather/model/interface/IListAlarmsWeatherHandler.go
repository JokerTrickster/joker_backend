package _interface

import "github.com/labstack/echo/v4"

type IListAlarmsWeatherHandler interface {
	ListAlarms(c echo.Context) error
}
