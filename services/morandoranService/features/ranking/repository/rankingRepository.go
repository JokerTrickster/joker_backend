package repository

import (
	"context"
	"fmt"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/ranking/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/morandoranService/features/ranking/model/interface"
	"gorm.io/gorm"
)

type RankingRepository struct {
	db *gorm.DB
}

func NewRankingRepository(db *gorm.DB) _interface.IRankingRepository {
	return &RankingRepository{db: db}
}

func (r *RankingRepository) List(ctx context.Context, gameType string, page, limit int) ([]entity.Ranking, int64, error) {
	var items []entity.Ranking
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Ranking{}).Where("game_type = ?", gameType)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count rankings: %w", err)
	}

	offset := (page - 1) * limit
	if err := query.Order("clear_time_ms ASC").Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list rankings: %w", err)
	}

	return items, total, nil
}

func (r *RankingRepository) FindByUserAndGame(ctx context.Context, userID uint, gameType string) (*entity.Ranking, error) {
	var ranking entity.Ranking
	if err := r.db.WithContext(ctx).Where("user_id = ? AND game_type = ?", userID, gameType).First(&ranking).Error; err != nil {
		return nil, err
	}
	return &ranking, nil
}

func (r *RankingRepository) Create(ctx context.Context, ranking *entity.Ranking) error {
	return r.db.WithContext(ctx).Create(ranking).Error
}

func (r *RankingRepository) Update(ctx context.Context, ranking *entity.Ranking) error {
	return r.db.WithContext(ctx).Save(ranking).Error
}

func (r *RankingRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entity.Ranking{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete ranking: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("ranking not found")
	}
	return nil
}

func (r *RankingRepository) GetRank(ctx context.Context, gameType string, clearTimeMs uint) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.Ranking{}).
		Where("game_type = ? AND clear_time_ms < ?", gameType, clearTimeMs).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to get rank: %w", err)
	}
	return int(count) + 1, nil
}
