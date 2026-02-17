package repository

import (
	"context"
	"time"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/interface"
	"gorm.io/gorm"
)

type TDUserRepository struct {
	db *gorm.DB
}

func NewTDUserRepository(db *gorm.DB) _interface.ITDUserRepository {
	return &TDUserRepository{db: db}
}

func (r *TDUserRepository) Create(ctx context.Context, user *entity.TDUser) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *TDUserRepository) GetByID(ctx context.Context, id uint) (*entity.TDUser, error) {
	var user entity.TDUser
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *TDUserRepository) GetByEmail(ctx context.Context, email string) (*entity.TDUser, error) {
	var user entity.TDUser
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *TDUserRepository) GetByUsername(ctx context.Context, username string) (*entity.TDUser, error) {
	var user entity.TDUser
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *TDUserRepository) Update(ctx context.Context, user *entity.TDUser) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *TDUserRepository) UpdateLastLogin(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entity.TDUser{}).Where("id = ?", id).Update("last_login", now).Error
}

func (r *TDUserRepository) GetStats(ctx context.Context, userID uint) (*entity.TDUserStats, error) {
	var stats entity.TDUserStats
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *TDUserRepository) CreateStats(ctx context.Context, stats *entity.TDUserStats) error {
	return r.db.WithContext(ctx).Create(stats).Error
}

func (r *TDUserRepository) UpdateStats(ctx context.Context, stats *entity.TDUserStats) error {
	return r.db.WithContext(ctx).Save(stats).Error
}
