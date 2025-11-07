package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"main/internal/handler"
	"main/shared/config"
	"main/shared/database"
	customErrors "main/shared/errors"
	"main/shared/logger"
	customMiddleware "main/shared/middleware"
	"main/shared/migrate"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
	defer rateLimiter.Close() // Ensure cleanup on shutdown
	e.Use(rateLimiter.Middleware())

	// Request timeout (30 seconds) - use Echo's built-in timeout middleware
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"success":   true,
			"message":   "Joker Backend is running",
			"timestamp": time.Now().Unix(),
		})
	})

	// API v1 routes
	v1 := e.Group("/api/v1")
	handler.RegisterRoutes(v1, db)

	// Start server with graceful shutdown
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Server starting",
		zap.String("port", port),
		zap.String("address", ":"+port),
	)

	// Start server in goroutine
	go func() {
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Give outstanding requests 10 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited gracefully")
}
