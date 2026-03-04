package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type LogoutHandler struct{}

func NewLogoutHandler() *LogoutHandler {
	return &LogoutHandler{}
}

func (h *LogoutHandler) Logout(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"message": "로그아웃 되었습니다.",
	})
}
