package main

import (
	"log"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/luxrobo/joker_backend/services/auth-service/internal/handler"
	"github.com/luxrobo/joker_backend/shared/config"
	"github.com/luxrobo/joker_backend/shared/database"
	customErrors "github.com/luxrobo/joker_backend/shared/errors"
	"github.com/luxrobo/joker_backend/shared/logger"
	customMiddleware "github.com/luxrobo/joker_backend/shared/middleware"
	"github.com/luxrobo/joker_backend/shared/migrate"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize logger
	logger.Init(cfg.LogLevel)
	defer logger.Sync()

	logger.Info("Starting Joker Backend",
		zap.String("environment", os.Getenv("ENV")),
		zap.String("log_level", cfg.LogLevel),
	)

	// Initialize database
	db, err := database.Connect(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	logger.Info("Database connected successfully")

	// Run migrations
	logger.Info("Running database migrations...")
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		// Default path for local development
		migrationsPath = "../../migrations"
	}
	migrateConfig := migrate.Config{
		MigrationsPath: migrationsPath,
		DatabaseName:   cfg.Database.Database,
	}
	if err := migrate.Run(db.DB, migrateConfig); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}
	logger.Info("Database migrations completed successfully")

	// Initialize Echo
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Set custom error handler
	e.HTTPErrorHandler = customErrors.CustomErrorHandler

	// Core middleware (order matters!)
	e.Use(customMiddleware.RequestID())
	e.Use(customMiddleware.Recovery())
	e.Use(customMiddleware.RequestLogger())
	e.Use(customMiddleware.CORS(cfg.CORS.AllowedOrigins, cfg.Env))

	// Rate limiting (10 requests per second, burst of 20)
	rateLimiter := customMiddleware.NewRateLimiter(10, 20)
	e.Use(rateLimiter.Middleware())

	// Request timeout (30 seconds)
	e.Use(customMiddleware.Timeout(30 * time.Second))

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"success": true,
			"message": "Joker Backend is running",
			"timestamp": time.Now().Unix(),
		})
	})

	// API v1 routes
	v1 := e.Group("/api/v1")
	handler.RegisterRoutes(v1, db)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Server starting",
		zap.String("port", port),
		zap.String("address", ":"+port),
	)

	if err := e.Start(":" + port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
