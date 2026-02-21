package handler

import (
	"time"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/repository"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/usecase"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// RegisterRoutes registers all Lotto Defense game routes
func RegisterRoutes(g *echo.Group, db *gorm.DB) {
	timeout := 30 * time.Second
	roundRepo := repository.NewGameRoundRepository(db)
	drawRepo := repository.NewLottoDrawRepository(db)
	roundUC := usecase.NewGameRoundUseCase(roundRepo, drawRepo, timeout)
	leaderboardUC := usecase.NewLeaderboardUseCase(roundRepo)

	NewRoundHandler(g, roundUC)
	NewLeaderboardHandler(g, leaderboardUC)
}
