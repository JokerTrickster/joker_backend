package repository

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/interface"
	"gorm.io/gorm"
)

type LottoDrawRepository struct {
	db *gorm.DB
}

func NewLottoDrawRepository(db *gorm.DB) _interface.ILottoDrawRepository {
	return &LottoDrawRepository{db: db}
}

func (r *LottoDrawRepository) Create(ctx context.Context, draw *entity.LottoDraw) error {
	return r.db.WithContext(ctx).Create(draw).Error
}

func (r *LottoDrawRepository) GetByRoundID(ctx context.Context, roundID uint) (*entity.LottoDraw, error) {
	var draw entity.LottoDraw
	err := r.db.WithContext(ctx).Where("round_id = ?", roundID).First(&draw).Error
	if err != nil {
		return nil, err
	}
	return &draw, nil
}
