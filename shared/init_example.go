package shared

// This file shows example usage of init.go
// Use it in main.go like this:

/*
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	authHandler "main/features/auth/handler"
	"main/shared"
	"github.com/JokerTrickster/joker_backend/shared/logger"

	"go.uber.org/zap"
)

func main() {
	// Method 1: Initialize with default configuration
	e, err := shared.Init(nil)
	if err != nil {
		log.Fatal("Failed to initialize:", err)
	}
	defer shared.Cleanup()

	// Method 2: Initialize with custom configuration
	// e, err := shared.Init(&shared.InitConfig{
	// 	LogLevel:    "debug",
	// 	Environment: "development",
	// })
	// if err != nil {
	// 	log.Fatal("Failed to initialize:", err)
	// }
	// defer shared.Cleanup()

	// Method 3: Must initialize (panics on failure)
	// e := shared.MustInit(&shared.InitConfig{
	// 	LogLevel:    "info",
	// 	Environment: "production",
	// })
	// defer shared.Cleanup()

	logger.Info("Starting application",
		zap.String("environment", shared.AppConfig.Env),
		zap.String("log_level", shared.AppConfig.LogLevel),
	)

	// Register routes
	authHandler.NewAuthHandler(e)

	// Start server with graceful shutdown
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Server starting", zap.String("port", port))

	go func() {
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited gracefully")
}
*/
