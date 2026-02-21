package _interface

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/request"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/response"
)

// IGameRoundUseCase defines game round business logic
type IGameRoundUseCase interface {
	StartRound(ctx context.Context, userID uint) (*response.RoundResponse, error)
	EndRound(ctx context.Context, userID uint, roundID uint, req *request.EndRoundRequest) (*response.RoundWithDrawResponse, error)
	GetMyRounds(ctx context.Context, userID uint, limit int) ([]response.RoundResponse, error)
	GetRound(ctx context.Context, userID uint, roundID uint) (*response.RoundWithDrawResponse, error)
}

// ILeaderboardUseCase defines leaderboard business logic
type ILeaderboardUseCase interface {
	GetLeaderboard(ctx context.Context, limit int) (*response.LeaderboardResponse, error)
}
