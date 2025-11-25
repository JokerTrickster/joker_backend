package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// getUserIDFromContext extracts user ID from context (set by JWT middleware)
func getUserIDFromContext(c echo.Context) (uint, error) {
	// Get userID from context (set by JWT middleware)
	userIDValue := c.Get("userID")
	if userIDValue == nil {
		return 0, echo.NewHTTPError(http.StatusUnauthorized, "user not found in context")
	}

	// The JWT middleware sets userID as uint
	userID, ok := userIDValue.(uint)
	if !ok {
		// Try to handle other numeric types just in case
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
