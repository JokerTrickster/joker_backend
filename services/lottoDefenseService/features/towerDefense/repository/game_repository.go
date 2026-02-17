package repository

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/interface"
	"gorm.io/gorm"
)

type TDGameRepository struct {
	db *gorm.DB
}

func NewTDGameRepository(db *gorm.DB) _interface.ITDGameRepository {
	return &TDGameRepository{db: db}
}

func (r *TDGameRepository) Create(ctx context.Context, result *entity.TDGameResult) error {
	return r.db.WithContext(ctx).Create(result).Error
}

func (r *TDGameRepository) GetByID(ctx context.Context, id uint) (*entity.TDGameResult, error) {
	var result entity.TDGameResult
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *TDGameRepository) GetHistory(ctx context.Context, userID uint, gameMode string, limit, offset int) ([]entity.TDGameResult, int64, error) {
	var results []entity.TDGameResult
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.TDGameResult{}).Where("user_id = ?", userID)

	if gameMode != "" && gameMode != "all" {
		query = query.Where("game_mode = ?", gameMode)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("played_at DESC").Limit(limit).Offset(offset).Find(&results).Error
	if err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func (r *TDGameRepository) GetHighestRound(ctx context.Context, userID uint, gameMode string) (uint, error) {
	var result struct {
		MaxRound uint
	}

	err := r.db.WithContext(ctx).Model(&entity.TDGameResult{}).
		Select("MAX(rounds_reached) as max_round").
		Where("user_id = ? AND game_mode = ?", userID, gameMode).
		Scan(&result).Error

	if err != nil {
		return 0, err
	}

	return result.MaxRound, nil
}
