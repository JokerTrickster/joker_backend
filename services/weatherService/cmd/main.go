// @title Joker Backend API
// @version 1.0
// @description This is the API documentation for the Joker Backend services.
// @host localhost:6000
// @BasePath /
// @schemes http
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	weatherHandler "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/handler"

	"github.com/JokerTrickster/joker_backend/shared"
	"github.com/JokerTrickster/joker_backend/shared/logger"

	"go.uber.org/zap"
)

func main() {
	// Initialize all shared components (logger, config, database, middleware, etc.)
	e, err := shared.Init(&shared.InitConfig{
		LogLevel:    os.Getenv("LOG_LEVEL"),
		Environment: os.Getenv("ENV"),
	})
	if err != nil {
		log.Fatal("Failed to initialize:", err)
	}
	defer shared.Cleanup()

	logger.Info("Starting Joker Backend",
		zap.String("environment", shared.AppConfig.Env),
		zap.String("log_level", shared.AppConfig.LogLevel),
	)

	// Register API routes
	weatherHandler.NewWeatherHandler(e)

	// Start server with graceful shutdown
	port := os.Getenv("PORT")
	if port == "" {
		port = "6000"
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
