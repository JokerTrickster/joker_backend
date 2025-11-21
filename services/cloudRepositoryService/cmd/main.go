// @title Cloud Repository Service API
// @version 1.0
// @description API for managing image and video files with S3 presigned URLs
// @host localhost:8080
// @BasePath /
// @schemes http
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/handler"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model"
	"github.com/JokerTrickster/joker_backend/shared"
	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
	"github.com/JokerTrickster/joker_backend/shared/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	// Initialize all shared components
	e, err := shared.Init(&shared.InitConfig{
		LogLevel:    os.Getenv("LOG_LEVEL"),
		Environment: os.Getenv("ENV"),
	})
	if err != nil {
		panic("Failed to initialize: " + err.Error())
	}
	defer shared.Cleanup()

	logger.Info("Starting Cloud Repository Service",
		zap.String("environment", shared.AppConfig.Env),
		zap.String("log_level", shared.AppConfig.LogLevel),
	)

	// Get configuration
	bucket := os.Getenv("CLOUD_REPOSITORY_BUCKET")
	if bucket == "" {
		bucket = "cloud-repository-dev"
	}

	// Get database connection
	database := mysql.GormMysqlDB
	if database == nil {
		logger.Fatal("Database connection is nil")
	}

	// Auto-migrate database
	if err := mysql.GormMysqlDB.AutoMigrate(&model.CloudFile{}, &model.Tag{}); err != nil {
		logger.GetLogger().Fatal("Failed to migrate database", zap.Error(err))
	}

	// Register routes
	api := e.Group("/api/v1")
	handler.RegisterRoutes(api, database, bucket)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Server starting",
		zap.String("port", port),
		zap.String("bucket", bucket),
	)

	// Start server in goroutine
	go func() {
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited gracefully")
}

func migrateDatabase(database *gorm.DB) error {
	logger.Info("Starting database migration...")
	
	if err := database.AutoMigrate(&model.CloudFile{}); err != nil {
		return fmt.Errorf("failed to migrate CloudFile model: %w", err)
	}
	
	logger.Info("Database migration completed successfully")
	return nil
}
