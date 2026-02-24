package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// getUserIDFromContext extracts user ID from Echo context
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

// bindAndValidate binds request body and validates it
func bindAndValidate(c echo.Context, req interface{}) error {
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

// successResponse returns a standardized success JSON response
func successResponse(c echo.Context, statusCode int, data interface{}) error {
	return c.JSON(statusCode, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// errorResponse returns a standardized error JSON response
func errorResponse(c echo.Context, statusCode int, message string) error {
	return c.JSON(statusCode, map[string]string{
		"error": message,
	})
}

// getQueryInt parses an integer query parameter with a default value
func getQueryInt(c echo.Context, key string, defaultValue int) int {
	if str := c.QueryParam(key); str != "" {
		if val, err := strconv.Atoi(str); err == nil {
			return val
		}
	}
	return defaultValue
}

// getQueryString returns a query parameter string or default value
func getQueryString(c echo.Context, key string, defaultValue string) string {
	if str := c.QueryParam(key); str != "" {
		return str
	}
	return defaultValue
}

// getParamUint parses a URL parameter as uint
func getParamUint(c echo.Context, key string) (uint, error) {
	idStr := c.Param(key)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid "+key)
	}
	return uint(id), nil
}
