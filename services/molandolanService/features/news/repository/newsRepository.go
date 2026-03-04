package repository

import (
	"context"
	"fmt"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/model/interface"
	"gorm.io/gorm"
)

type NewsRepository struct {
	db *gorm.DB
}

func NewNewsRepository(db *gorm.DB) _interface.INewsRepository {
	return &NewsRepository{db: db}
}

func (r *NewsRepository) List(ctx context.Context, page, limit int, category string) ([]entity.News, int64, error) {
	var items []entity.News
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.News{})
	if category != "" {
		query = query.Where("category = ?", category)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count news: %w", err)
	}

	offset := (page - 1) * limit
	if err := query.Order("date DESC, id DESC").Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list news: %w", err)
	}

	return items, total, nil
}

func (r *NewsRepository) FindByID(ctx context.Context, id uint) (*entity.News, error) {
	var news entity.News
	if err := r.db.WithContext(ctx).First(&news, id).Error; err != nil {
		return nil, fmt.Errorf("news not found")
	}
	return &news, nil
}

func (r *NewsRepository) Create(ctx context.Context, news *entity.News) (*entity.News, error) {
	if err := r.db.WithContext(ctx).Create(news).Error; err != nil {
		return nil, fmt.Errorf("failed to create news: %w", err)
	}
	return news, nil
}

func (r *NewsRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) (*entity.News, error) {
	if err := r.db.WithContext(ctx).Model(&entity.News{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update news: %w", err)
	}
	var news entity.News
	if err := r.db.WithContext(ctx).First(&news, id).Error; err != nil {
		return nil, fmt.Errorf("news not found after update")
	}
	return &news, nil
}

func (r *NewsRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entity.News{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete news: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("news not found")
	}
	return nil
}
