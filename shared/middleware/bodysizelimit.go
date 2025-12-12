package middleware

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	// DefaultMaxBodySize is 10MB - reasonable limit for most API requests
	DefaultMaxBodySize = 10 << 20 // 10 MB

	// MaxFileUploadSize is 100MB - for file upload endpoints
	MaxFileUploadSize = 100 << 20 // 100 MB
)

// BodySizeLimit creates middleware that limits request body size
// This prevents DoS attacks from excessively large request bodies
func BodySizeLimit(maxBytes int64) echo.MiddlewareFunc {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBodySize
	}

	return middleware.BodyLimit(fmt.Sprintf("%dB", maxBytes))
}

// DefaultBodySizeLimit applies a default 10MB body size limit
func DefaultBodySizeLimit() echo.MiddlewareFunc {
	return BodySizeLimit(DefaultMaxBodySize)
}

// FileUploadBodySizeLimit applies a 100MB body size limit for file uploads
func FileUploadBodySizeLimit() echo.MiddlewareFunc {
	return BodySizeLimit(MaxFileUploadSize)
}
