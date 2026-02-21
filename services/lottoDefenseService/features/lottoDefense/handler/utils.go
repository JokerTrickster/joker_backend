package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func getUserIDFromContext(c echo.Context) (uint, error) {
	userIDValue := c.Get("userID")
	if userIDValue == nil {
		return 0, echo.NewHTTPError(http.StatusUnauthorized, "user not found in context")
	}
	userID, ok := userIDValue.(uint)
	if !ok {
		switch v := userIDValue.(type) {
		case int:
			return uint(v), nil
		case int64:
			return uint(v), nil
		case float64:
			return uint(v), nil
		default:
			return 0, echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID type")
		}
	}
	return userID, nil
}
