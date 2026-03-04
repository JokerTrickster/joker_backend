package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/ranking/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/morandoranService/features/ranking/model/interface"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/ranking/model/response"
	"gorm.io/gorm"
)

type SubmitUseCase struct {
	Repo           _interface.IRankingRepository
	ContextTimeout time.Duration
}

func NewSubmitUseCase(repo _interface.IRankingRepository, timeout time.Duration) *SubmitUseCase {
	return &SubmitUseCase{Repo: repo, ContextTimeout: timeout}
}

func (uc *SubmitUseCase) Submit(c context.Context, userID uint, nickname, gameType string, clearTimeMs uint) (*response.ResSubmitRanking, error) {
	ctx, cancel := context.WithTimeout(c, uc.ContextTimeout)
	defer cancel()

	existing, err := uc.Repo.FindByUserAndGame(ctx, userID, gameType)
	isNewRecord := false

	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to check existing ranking: %w", err)
		}
		newRanking := &entity.Ranking{
			UserID:      userID,
			GameType:    gameType,
			Nickname:    nickname,
			ClearTimeMs: clearTimeMs,
		}
		if err := uc.Repo.Create(ctx, newRanking); err != nil {
			return nil, fmt.Errorf("failed to create ranking: %w", err)
		}
		isNewRecord = true
	} else {
		if clearTimeMs < existing.ClearTimeMs {
			existing.ClearTimeMs = clearTimeMs
			existing.Nickname = nickname
			if err := uc.Repo.Update(ctx, existing); err != nil {
				return nil, fmt.Errorf("failed to update ranking: %w", err)
			}
			isNewRecord = true
		}
	}

	rank, err := uc.Repo.GetRank(ctx, gameType, clearTimeMs)
	if err != nil {
		return nil, fmt.Errorf("failed to get rank: %w", err)
	}

	return &response.ResSubmitRanking{
		Rank:        rank,
		IsNewRecord: isNewRecord,
	}, nil
}
