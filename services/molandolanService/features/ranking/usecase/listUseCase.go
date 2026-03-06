package usecase

import (
	"context"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/ranking/model/interface"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/ranking/model/request"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/ranking/model/response"
)

type ListUseCase struct {
	Repo           _interface.IRankingRepository
	ContextTimeout time.Duration
}

func NewListUseCase(repo _interface.IRankingRepository, timeout time.Duration) *ListUseCase {
	return &ListUseCase{Repo: repo, ContextTimeout: timeout}
}

func (uc *ListUseCase) List(c context.Context, gameType string, req *request.ReqListRanking) (*response.ResRankingList, error) {
	ctx, cancel := context.WithTimeout(c, uc.ContextTimeout)
	defer cancel()

	limit := req.Limit
	if limit < 1 {
		limit = 5
	}
	if limit > 100 {
		limit = 100
	}

	items, _, err := uc.Repo.List(ctx, gameType, 1, limit)
	if err != nil {
		return nil, err
	}

	rankings := make([]response.ResRankingItem, len(items))
	for i, item := range items {
		rankings[i] = response.ResRankingItem{
			Rank:        i + 1,
			Nickname:    item.Nickname,
			ClearTimeMs: item.ClearTimeMs,
			CreatedAt:   item.CreatedAt,
		}
	}

	return &response.ResRankingList{Rankings: rankings}, nil
}
