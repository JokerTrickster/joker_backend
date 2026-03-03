package handler

import (
	"time"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/repository"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/usecase"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// RegisterRoutes registers all Tower Defense game routes.
// publicGroup is for unauthenticated routes (auth/register, auth/login).
// protectedGroup has JWTAuth middleware applied for authenticated routes.
func RegisterRoutes(publicGroup *echo.Group, protectedGroup *echo.Group, db *gorm.DB, jwtSecret string) {
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

	// Public routes (no auth required)
	NewTDAuthHandler(publicGroup, authUC)

	// Protected routes (JWT auth required)
	NewTDUserHandler(protectedGroup, authUC, gameUC)
	NewTDGameHandler(protectedGroup, gameUC)
	NewTDQuestHandler(protectedGroup, questUC)
	NewTDRoomHandler(protectedGroup, roomUC)
	NewTDCoopStateHandler(protectedGroup, roomUC)
}
