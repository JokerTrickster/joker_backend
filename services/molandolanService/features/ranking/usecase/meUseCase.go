package usecase

import (
	"context"
	"fmt"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/morandoranService/features/ranking/model/interface"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/ranking/model/response"
	"gorm.io/gorm"
)

type MeUseCase struct {
	Repo           _interface.IRankingRepository
	ContextTimeout time.Duration
}

func NewMeUseCase(repo _interface.IRankingRepository, timeout time.Duration) *MeUseCase {
	return &MeUseCase{Repo: repo, ContextTimeout: timeout}
}

func (uc *MeUseCase) MyRanking(c context.Context, userID uint, gameType string) (*response.ResMyRanking, error) {
	ctx, cancel := context.WithTimeout(c, uc.ContextTimeout)
	defer cancel()

	ranking, err := uc.Repo.FindByUserAndGame(ctx, userID, gameType)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("ranking not found")
		}
		return nil, err
	}

	rank, err := uc.Repo.GetRank(ctx, gameType, ranking.ClearTimeMs)
	if err != nil {
		return nil, err
	}

	return &response.ResMyRanking{
		Rank: rank,
		Entry: response.ResMyRankingEntry{
			Rank:        rank,
			Nickname:    ranking.Nickname,
			ClearTimeMs: ranking.ClearTimeMs,
			CreatedAt:   ranking.CreatedAt,
		},
	}, nil
}
