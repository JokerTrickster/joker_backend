package usecase

import (
	"context"

	_interface "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/interface"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/response"
)

type LeaderboardUseCase struct {
	roundRepo _interface.IGameRoundRepository
}

func NewLeaderboardUseCase(roundRepo _interface.IGameRoundRepository) _interface.ILeaderboardUseCase {
	return &LeaderboardUseCase{roundRepo: roundRepo}
}

func (u *LeaderboardUseCase) GetLeaderboard(ctx context.Context, limit int) (*response.LeaderboardResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	rows, err := u.roundRepo.Leaderboard(ctx, limit)
	if err != nil {
		return nil, err
	}
	entries := make([]response.LeaderboardEntry, len(rows))
	for i := range rows {
		entries[i] = response.LeaderboardEntry{
			Rank:   i + 1,
			UserID: rows[i].UserID,
			Name:   rows[i].Name,
			Score:  rows[i].Score,
		}
	}
	return &response.LeaderboardResponse{Entries: entries}, nil
}
