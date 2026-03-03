package _interface

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/ranking/model/entity"
)

type IRankingRepository interface {
	List(ctx context.Context, gameType string, page, limit int) ([]entity.Ranking, int64, error)
	FindByUserAndGame(ctx context.Context, userID uint, gameType string) (*entity.Ranking, error)
	Create(ctx context.Context, ranking *entity.Ranking) error
	Update(ctx context.Context, ranking *entity.Ranking) error
	Delete(ctx context.Context, id uint) error
	GetRank(ctx context.Context, gameType string, clearTimeMs uint) (int, error)
}
