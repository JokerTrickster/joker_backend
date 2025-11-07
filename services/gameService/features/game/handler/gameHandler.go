package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/JokerTrickster/joker_backend/shared/logger"
	"go.uber.org/zap"
)

// RegisterGameRoutes registers all game-related routes
func RegisterGameRoutes(e *echo.Echo) {
	// API v1 group
	v1 := e.Group("/api/v1")

	// Game routes
	game := v1.Group("/game")
	game.GET("/list", listGames)
	game.GET("/:id", getGame)
	game.POST("", createGame)
	game.PUT("/:id", updateGame)
	game.DELETE("/:id", deleteGame)

	logger.Info("Game routes registered successfully")
}

// listGames returns list of games
func listGames(c echo.Context) error {
	logger.Info("List games endpoint called",
		zap.String("request_id", c.Response().Header().Get("X-Request-ID")),
	)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Game list retrieved successfully",
		"data": []map[string]interface{}{
			{
				"id":   1,
				"name": "Find It",
				"type": "puzzle",
			},
			{
				"id":   2,
				"name": "Slime War",
				"type": "action",
			},
		},
	})
}

// getGame returns a specific game by ID
func getGame(c echo.Context) error {
	id := c.Param("id")

	logger.Info("Get game endpoint called",
		zap.String("request_id", c.Response().Header().Get("X-Request-ID")),
		zap.String("game_id", id),
	)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Game retrieved successfully",
		"data": map[string]interface{}{
			"id":   id,
			"name": "Sample Game",
			"type": "puzzle",
		},
	})
}

// createGame creates a new game
func createGame(c echo.Context) error {
	logger.Info("Create game endpoint called",
		zap.String("request_id", c.Response().Header().Get("X-Request-ID")),
	)

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Game created successfully",
		"data": map[string]interface{}{
			"id":   "new-game-id",
			"name": "New Game",
		},
	})
}

// updateGame updates an existing game
func updateGame(c echo.Context) error {
	id := c.Param("id")

	logger.Info("Update game endpoint called",
		zap.String("request_id", c.Response().Header().Get("X-Request-ID")),
		zap.String("game_id", id),
	)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Game updated successfully",
		"data": map[string]interface{}{
			"id": id,
		},
	})
}

// deleteGame deletes a game
func deleteGame(c echo.Context) error {
	id := c.Param("id")

	logger.Info("Delete game endpoint called",
		zap.String("request_id", c.Response().Header().Get("X-Request-ID")),
		zap.String("game_id", id),
	)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Game deleted successfully",
	})
}
