package repository

import (
	"context"
	"time"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/interface"
	"gorm.io/gorm"
)

type TDQuestRepository struct {
	db *gorm.DB
}

func NewTDQuestRepository(db *gorm.DB) _interface.ITDQuestRepository {
	return &TDQuestRepository{db: db}
}

func (r *TDQuestRepository) Create(ctx context.Context, quest *entity.TDQuest) error {
	return r.db.WithContext(ctx).Create(quest).Error
}

func (r *TDQuestRepository) GetByID(ctx context.Context, id uint) (*entity.TDQuest, error) {
	var quest entity.TDQuest
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&quest).Error
	if err != nil {
		return nil, err
	}
	return &quest, nil
}

func (r *TDQuestRepository) GetActiveQuests(ctx context.Context, userID uint) ([]entity.TDQuest, error) {
	var quests []entity.TDQuest
	err := r.db.WithContext(ctx).Where("user_id = ? AND status = ?", userID, "active").Find(&quests).Error
	return quests, err
}

func (r *TDQuestRepository) UpdateProgress(ctx context.Context, questID uint, increment uint) error {
	return r.db.WithContext(ctx).Model(&entity.TDQuest{}).
		Where("id = ?", questID).
		UpdateColumn("current_count", gorm.Expr("current_count + ?", increment)).Error
}

func (r *TDQuestRepository) CompleteQuest(ctx context.Context, questID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entity.TDQuest{}).
		Where("id = ?", questID).
		Updates(map[string]interface{}{
			"status":       "completed",
			"completed_at": now,
		}).Error
}

func (r *TDQuestRepository) ClaimQuest(ctx context.Context, questID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entity.TDQuest{}).
		Where("id = ?", questID).
		Updates(map[string]interface{}{
			"status":     "claimed",
			"claimed_at": now,
		}).Error
}
