package _interface

import "github.com/labstack/echo/v4"

type ICheckEmailAuthHandler interface {
	CheckEmail(c echo.Context) error
}
