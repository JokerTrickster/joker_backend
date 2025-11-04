package main

import (
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/luxrobo/joker_backend/config"
	"github.com/luxrobo/joker_backend/internal/handler"
	"github.com/luxrobo/joker_backend/pkg/database"
	"github.com/luxrobo/joker_backend/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize logger
	logger.Init(cfg.LogLevel)

	// Initialize database
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize Echo
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"success": true,
			"message": "Joker Backend is running",
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

	logger.Info("Starting server on port " + port)
	if err := e.Start(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
