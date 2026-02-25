package handler

import (
	"time"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/repository"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/usecase"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// RegisterRoutes registers all Tower Defense game routes
func RegisterRoutes(g *echo.Group, db *gorm.DB, jwtSecret string) {
	timeout := 30 * time.Second

	// Repositories
	userRepo := repository.NewTDUserRepository(db)
	gameRepo := repository.NewTDGameRepository(db)
	questRepo := repository.NewTDQuestRepository(db)
	roomRepo := repository.NewTDRoomRepository(db)

	// Usecases
	authUC := usecase.NewTDAuthUseCase(userRepo, jwtSecret, timeout)
	gameUC := usecase.NewTDGameUseCase(gameRepo, userRepo, timeout)
	questUC := usecase.NewTDQuestUseCase(questRepo, userRepo, timeout)
	roomUC := usecase.NewTDRoomUseCase(roomRepo, userRepo, timeout)

	// Handlers
	NewTDAuthHandler(g, authUC)              // /auth/register, /auth/login
	NewTDUserHandler(g, authUC, gameUC)      // /users/me, /users/me/stats
	NewTDGameHandler(g, gameUC)              // /game/single/result, /game/history
	NewTDQuestHandler(g, questUC)            // /quests, /quests/:id/progress, /quests/:id/claim
	NewTDRoomHandler(g, roomUC)              // /coop/rooms, /coop/rooms/join, /coop/rooms/:id, etc
	NewTDCoopStateHandler(g, roomUC)         // /coop/rooms/:id/state, /coop/rooms/:id/opponent-state
}
