package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	lottoHandler "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/handler"
	lottoEntity "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/entity"
	tdHandler "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/handler"
	tdEntity "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/entity"
	"github.com/JokerTrickster/joker_backend/shared"
	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
	"github.com/JokerTrickster/joker_backend/shared/logger"
	sharedMiddleware "github.com/JokerTrickster/joker_backend/shared/middleware"
	"go.uber.org/zap"
)

func main() {
	e, err := shared.Init(&shared.InitConfig{
		LogLevel:    os.Getenv("LOG_LEVEL"),
		Environment: os.Getenv("ENV"),
	})
	if err != nil {
		panic("Failed to initialize: " + err.Error())
	}
	defer shared.Cleanup()

	logger.Info("Starting Lotto Defense Service",
		zap.String("environment", shared.AppConfig.Env),
		zap.String("log_level", shared.AppConfig.LogLevel),
	)

	database := mysql.GormMysqlDB
	if database == nil {
		logger.Fatal("Database connection is nil - check DB environment variables")
	}

	logger.Info("Running database migration...")
	if err := database.AutoMigrate(
		// Lotto Defense (existing)
		&lottoEntity.GameRound{}, 
		&lottoEntity.LottoDraw{},
		// Tower Defense (new)
		&tdEntity.TDUser{},
		&tdEntity.TDUserStats{},
		&tdEntity.TDGameResult{},
		&tdEntity.TDQuest{},
		&tdEntity.TDReward{},
		&tdEntity.TDRoom{},
		&tdEntity.TDRoomPlayer{},
		&tdEntity.TDFriendship{},
	); err != nil {
		logger.Fatal("Failed to migrate database", zap.Error(err))
	}
	logger.Info("Database migration completed")

	// API groups
	api := e.Group("/api/v1/game")
	api.Use(sharedMiddleware.JWTAuth())

	tdPublic := e.Group("/api/v1/td")
	tdProtected := e.Group("/api/v1/td")
	tdProtected.Use(sharedMiddleware.JWTAuth())

	// Register routes
	lottoHandler.RegisterRoutes(api, database)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		if os.Getenv("IS_LOCAL") == "true" {
			jwtSecret = "local-dev-td-secret-do-not-use-in-production"
		} else {
			logger.Fatal("JWT_SECRET must be set in production")
		}
	}
	tdHandler.RegisterRoutes(tdPublic, tdProtected, database, jwtSecret)

	port := os.Getenv("PORT")
	if port == "" {
		port = "18082"
	}
	logger.Info("Server starting", zap.String("port", port))

	go func() {
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

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

