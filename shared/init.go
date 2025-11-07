package shared

import (
	"fmt"
	"time"

	"github.com/JokerTrickster/joker_backend/shared/config"
	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
	"github.com/JokerTrickster/joker_backend/shared/errors"
	"github.com/JokerTrickster/joker_backend/shared/logger"
	"github.com/JokerTrickster/joker_backend/shared/middleware"
	"github.com/JokerTrickster/joker_backend/shared/utils"

	_aws "github.com/JokerTrickster/joker_backend/shared/aws"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

var (
	// Global instances
	AppConfig   *config.Config
	EchoServer  *echo.Echo
	RateLimiter *middleware.RateLimiter
)

// InitConfig holds shared module initialization configuration
type InitConfig struct {
	// Log level (debug, info, warn, error)
	LogLevel string
	// Environment (development, production)
	Environment string
}

// Init initializes all shared components and returns configured Echo server
// Call this in main.go before starting the application
func Init(cfg *InitConfig) (*echo.Echo, error) {
	if cfg == nil {
		cfg = &InitConfig{
			LogLevel:    "info",
			Environment: "development",
		}
	}

	// 1. Logger initialization
	if err := initLogger(cfg.LogLevel); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	logger.Info("Logger initialized successfully")

	// 2. Config loading
	appConfig, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	AppConfig = appConfig
	logger.Info("Configuration loaded successfully")

	// 4. AWS initialization (optional)
	// Uncomment when AWS services are needed
	if err := _aws.InitAws(); err != nil {
		return nil, fmt.Errorf("failed to initialize AWS: %w", err)
	}
	logger.Info("AWS services initialized successfully")

	// 3. Database initialization (MySQL)
	if err := initDatabase(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	logger.Info("Database initialized successfully")

	// 5. Echo server initialization
	e := initEchoServer(appConfig)
	EchoServer = e
	logger.Info("Echo server initialized successfully")

	// 6. Middleware initialization
	if err := initMiddleware(e, appConfig); err != nil {
		return nil, fmt.Errorf("failed to initialize middleware: %w", err)
	}
	logger.Info("Middleware initialized successfully")

	logger.Info("Shared module initialization completed",
		zap.String("environment", appConfig.Env),
		zap.String("log_level", cfg.LogLevel),
	)

	return e, nil
}

// initLogger initializes the logging system
func initLogger(logLevel string) error {
	if logLevel == "" {
		logLevel = "info"
	}
	logger.Init(logLevel)
	return nil
}

// initDatabase initializes the database connection
func initDatabase() error {
	if err := mysql.InitMySQL(); err != nil {
		return fmt.Errorf("mysql initialization failed: %w", err)
	}
	return nil
}

// initEchoServer initializes Echo web framework
func initEchoServer(cfg *config.Config) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Set custom validator
	e.Validator = utils.NewValidator()

	// Set custom error handler
	e.HTTPErrorHandler = errors.CustomErrorHandler

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"success":   true,
			"message":   "Joker Backend is running",
			"timestamp": time.Now().Unix(),
		})
	})

	return e
}

// initMiddleware initializes all middleware
func initMiddleware(e *echo.Echo, cfg *config.Config) error {
	// Core middleware (order matters!)
	e.Use(middleware.RequestID())
	e.Use(middleware.Recovery())
	e.Use(middleware.RequestLogger())
	e.Use(middleware.CORS(cfg.CORS.AllowedOrigins, cfg.Env))

	// Rate limiting (10 requests per second, burst of 20)
	rateLimiter := middleware.NewRateLimiter(10, 20)
	RateLimiter = rateLimiter
	e.Use(rateLimiter.Middleware())

	// Request timeout (30 seconds) - use Echo's built-in timeout middleware
	e.Use(echoMiddleware.TimeoutWithConfig(echoMiddleware.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))

	logger.Info("Middleware initialized",
		zap.String("cors_origins", fmt.Sprintf("%v", cfg.CORS.AllowedOrigins)),
		zap.String("timeout", "30s"),
	)

	return nil
}

// initAWS initializes AWS services
// S3, SES, SSM and other AWS services can be initialized here
// func initAWS() error {
// 	// AWS initialization logic
// 	// aws.Init() or similar initialization code
// 	return nil
// }

// Cleanup performs cleanup of shared module resources
// Call this with defer in main.go
func Cleanup() error {
	logger.Info("Starting shared module cleanup...")

	// Rate limiter cleanup
	if RateLimiter != nil {
		RateLimiter.Close()
		logger.Info("Rate limiter cleaned up")
	}

	// Logger sync (flush any buffered logs)
	logger.Sync()
	// Note: Sync errors on stderr are acceptable for cleanup

	// Database cleanup
	// Close DB connection pool if needed
	// mysql.Close() or similar cleanup code

	logger.Info("Shared module cleanup completed")
	return nil
}

// MustInit calls Init and panics on error
// Use this when initialization is required and failure should stop the application
func MustInit(cfg *InitConfig) *echo.Echo {
	e, err := Init(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize shared module: %v", err))
	}
	return e
}
