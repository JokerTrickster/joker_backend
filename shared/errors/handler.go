package errors

import (
	"net/http"

	"github.com/JokerTrickster/joker_backend/shared/logger"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// CustomErrorHandler handles errors in a centralized way
func CustomErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	var (
		code    = http.StatusInternalServerError
		errCode = ErrCodeInternalServer
		message = "Internal server error"
	)

	// Check if it's our custom AppError
	if appErr, ok := err.(*AppError); ok {
		code = appErr.HTTPStatus
		errCode = appErr.Code
		message = appErr.Message

		// Log internal errors
		if code >= 500 {
			logger.Error("Application error",
				zap.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
				zap.String("error_code", errCode),
				zap.String("message", message),
				zap.Error(appErr.Err),
			)
		} else {
			logger.Warn("Client error",
				zap.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
				zap.String("error_code", errCode),
				zap.String("message", message),
			)
		}
	} else if echoErr, ok := err.(*echo.HTTPError); ok {
		// Handle Echo's HTTPError
		code = echoErr.Code
		if echoErr.Internal != nil {
			message = echoErr.Internal.Error()
		} else if msg, ok := echoErr.Message.(string); ok {
			message = msg
		}

		// Map HTTP status to error code
		errCode = mapHTTPStatusToCode(code)

		logger.Warn("HTTP error",
			zap.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
			zap.Int("status_code", code),
			zap.String("message", message),
		)
	} else {
		// Unknown error
		logger.Error("Unknown error",
			zap.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
			zap.Error(err),
		)
	}

	// Send JSON response
	c.JSON(code, map[string]interface{}{
		"error": map[string]interface{}{
			"code":    errCode,
			"message": message,
		},
	})
}

// mapHTTPStatusToCode maps HTTP status codes to error codes
func mapHTTPStatusToCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return ErrCodeBadRequest
	case http.StatusUnauthorized:
		return ErrCodeUnauthorized
	case http.StatusForbidden:
		return ErrCodeForbidden
	case http.StatusNotFound:
		return ErrCodeNotFound
	case http.StatusConflict:
		return ErrCodeConflict
	case http.StatusUnprocessableEntity:
		return ErrCodeUnprocessableEntity
	case http.StatusTooManyRequests:
		return "RATE_LIMIT_EXCEEDED"
	case http.StatusServiceUnavailable:
		return ErrCodeServiceUnavailable
	default:
		return ErrCodeInternalServer
	}
}
