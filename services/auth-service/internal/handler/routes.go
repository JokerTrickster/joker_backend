package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/luxrobo/joker_backend/shared/database"
)

func RegisterRoutes(g *echo.Group, db *database.DB) {
	// Example: User routes
	userHandler := NewUserHandler(db)
	g.GET("/users/:id", userHandler.GetUser)
	g.POST("/users", userHandler.CreateUser)
}
