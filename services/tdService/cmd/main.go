package main

import (
	"fmt"
	"log"
	"os"

	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/handler"
	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/entity"
	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/repository"
	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/usecase"
	"github.com/JokerTrickster/joker_backend/services/tdService/pkg/websocket"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "rootpassword")
	dbName := getEnv("DB_NAME", "td_service")
	jwtSecret := getEnv("JWT_SECRET", "test-secret-key")
	env := getEnv("ENV", "local")
	if env != "local" && env != "development" {
		if dbPassword == "rootpassword" {
			log.Fatal("DB_PASSWORD must be set in production")
		}
		if jwtSecret == "test-secret-key" {
			log.Fatal("JWT_SECRET must be set in production")
		}
	}
	wsBaseURL := getEnv("WS_BASE_URL", "ws://localhost:18082/ws")
	port := getEnv("PORT", "18082")
	corsOrigins := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")

	// Connect to database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate
	if err := db.AutoMigrate(
		&entity.Player{},
		&entity.PlayerStats{},
		&entity.GameSession{},
		&entity.GameResult{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize repositories
	playerRepo := repository.NewPlayerRepository(db)
	sessionRepo := repository.NewGameSessionRepository(db)

	// Initialize use cases
	authUseCase := usecase.NewAuthUseCase(playerRepo, jwtSecret)
	gameUseCase := usecase.NewGameUseCase(sessionRepo, playerRepo, wsBaseURL)

	// Initialize WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authUseCase)
	gameHandler := handler.NewGameHandler(gameUseCase, authUseCase)
	wsHandler := handler.NewWebSocketHandler(hub, authUseCase)

	// Setup Gin router
	r := gin.Default()

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{corsOrigins}
	config.AllowHeaders = append(config.AllowHeaders, "Authorization")
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	r.Use(cors.New(config))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// API routes
	api := r.Group("/api/v1/td")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Game routes
		game := api.Group("/game")
		{
			game.POST("/session", gameHandler.CreateSession)
			game.POST("/session/join", gameHandler.JoinSession)
			game.GET("/session/:sessionId", gameHandler.GetSessionInfo)
			game.POST("/result", gameHandler.SaveResult)
		}

		// Profile routes
		profile := api.Group("/profile")
		{
			profile.GET("/:userId", gameHandler.GetProfile)
		}
	}

	// WebSocket routes
	ws := r.Group("/ws")
	{
		ws.GET("/game/:sessionId", wsHandler.HandleWebSocket)
	}

	// Start server
	log.Printf("TD Service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}