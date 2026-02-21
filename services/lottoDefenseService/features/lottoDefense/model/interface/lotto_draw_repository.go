package _interface

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/entity"
)

// ILottoDrawRepository defines persistence for lotto draws
type ILottoDrawRepository interface {
	Create(ctx context.Context, draw *entity.LottoDraw) error
	GetByRoundID(ctx context.Context, roundID uint) (*entity.LottoDraw, error)
}
