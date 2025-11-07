package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/JokerTrickster/joker_backend/shared/logger"
	"go.uber.org/zap"
)

// CORS returns a configured CORS middleware with environment-based origin control
func CORS(allowedOrigins string, env string) echo.MiddlewareFunc {
	// Parse comma-separated origins
	var origins []string
	if allowedOrigins != "" {
		origins = strings.Split(allowedOrigins, ",")
		// Trim spaces
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
		}
	}

	// Development mode: allow wildcard if explicitly set
	isDevelopment := env == "development" || env == "dev" || env == ""

	if isDevelopment {
		// In development, allow wildcard if configured
		if len(origins) == 1 && origins[0] == "*" {
			logger.Warn("⚠️  CORS wildcard (*) enabled in development mode - NOT FOR PRODUCTION!")
		} else if len(origins) == 0 {
			// Default to wildcard in development
			origins = []string{"*"}
			logger.Warn("⚠️  CORS not configured, defaulting to wildcard (*) in development mode")
		}
	} else {
		// Production mode: strict checks
		for _, origin := range origins {
			if origin == "*" {
				logger.Fatal("CORS wildcard (*) is not allowed in production. Configure CORS_ALLOWED_ORIGINS with specific domains")
			}
		}

		if len(origins) == 0 {
			logger.Fatal("CORS_ALLOWED_ORIGINS must be configured in production. Example: https://example.com,https://app.example.com")
		}
	}

	logger.Info("CORS configured",
		zap.Strings("allowed_origins", origins),
		zap.String("environment", env),
	)

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
