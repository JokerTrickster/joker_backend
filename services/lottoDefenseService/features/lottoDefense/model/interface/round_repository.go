package _interface

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/entity"
)

// LeaderboardRow is a single row for leaderboard (user_id, name, best score)
type LeaderboardRow struct {
	UserID uint
	Name   string
	Score  uint
}

// IGameRoundRepository defines persistence for game rounds
type IGameRoundRepository interface {
	Create(ctx context.Context, round *entity.GameRound) error
	GetByID(ctx context.Context, id uint) (*entity.GameRound, error)
	GetByIDAndUser(ctx context.Context, id, userID uint) (*entity.GameRound, error)
	Update(ctx context.Context, round *entity.GameRound) error
	ListByUserID(ctx context.Context, userID uint, limit int) ([]entity.GameRound, error)
	TopScores(ctx context.Context, limit int) ([]entity.GameRound, error)
	Leaderboard(ctx context.Context, limit int) ([]LeaderboardRow, error)
}
