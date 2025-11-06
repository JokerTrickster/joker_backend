package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/luxrobo/joker_backend/shared/logger"
	"go.uber.org/zap"
)

// CORS returns a configured CORS middleware with environment-based origin control
func CORS(allowedOrigins string) echo.MiddlewareFunc {
	// Parse comma-separated origins
	var origins []string
	if allowedOrigins != "" {
		origins = strings.Split(allowedOrigins, ",")
		// Trim spaces
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
		}
	}

	// Security check: Never allow wildcard in production
	for _, origin := range origins {
		if origin == "*" {
			logger.Fatal("CORS wildcard (*) is not allowed. Configure CORS_ALLOWED_ORIGINS environment variable")
		}
	}

	// If no origins configured, fail secure
	if len(origins) == 0 {
		logger.Fatal("CORS_ALLOWED_ORIGINS must be configured. Example: https://example.com,https://app.example.com")
	}

	logger.Info("CORS configured", zap.Strings("allowed_origins", origins))

	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: origins,
		AllowMethods: []string{
			echo.GET,
			echo.POST,
			echo.PUT,
			echo.PATCH,
			echo.DELETE,
			echo.OPTIONS,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			echo.HeaderXRequestID,
		},
		ExposeHeaders: []string{
			echo.HeaderXRequestID,
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	})
}
