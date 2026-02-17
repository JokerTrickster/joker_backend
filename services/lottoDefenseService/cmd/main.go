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
	"github.com/JokerTrickster/joker_backend/shared/jwt"
	"github.com/JokerTrickster/joker_backend/shared/logger"
	"github.com/labstack/echo/v4"
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
	tdApi := e.Group("/api/v1/td")
	
	// JWT middleware for local development
	if os.Getenv("IS_LOCAL") == "true" {
		api.Use(jwtMiddleware())
		tdApi.Use(jwtMiddleware())
	}
	
	// Register routes
	lottoHandler.RegisterRoutes(api, database)
	
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "tower-defense-secret-key-change-in-production"
	}
	tdHandler.RegisterRoutes(tdApi, database, jwtSecret)

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

func jwtMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				tokenString := authHeader[7:]
				userID, _, err := jwt.ParseToken(tokenString)
				if err == nil {
					c.Set("userID", userID)
				}
			}
			return next(c)
		}
	}
}
