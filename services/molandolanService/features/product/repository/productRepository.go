package repository

import (
	"context"
	"fmt"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/product/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/product/model/interface"
	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) _interface.IProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) List(ctx context.Context, page, limit int, category string, inStock *bool) ([]entity.Product, int64, error) {
	var items []entity.Product
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Product{})
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if inStock != nil {
		query = query.Where("in_stock = ?", *inStock)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	offset := (page - 1) * limit
	if err := query.Order("id DESC").Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}

	return items, total, nil
}

func (r *ProductRepository) FindByID(ctx context.Context, id uint) (*entity.Product, error) {
	var product entity.Product
	if err := r.db.WithContext(ctx).First(&product, id).Error; err != nil {
		return nil, fmt.Errorf("product not found")
	}
	return &product, nil
}

func (r *ProductRepository) Create(ctx context.Context, product *entity.Product) (*entity.Product, error) {
	if err := r.db.WithContext(ctx).Create(product).Error; err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}
	return product, nil
}

func (r *ProductRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) (*entity.Product, error) {
	if err := r.db.WithContext(ctx).Model(&entity.Product{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}
	var product entity.Product
	if err := r.db.WithContext(ctx).First(&product, id).Error; err != nil {
		return nil, fmt.Errorf("product not found after update")
	}
	return &product, nil
}

func (r *ProductRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entity.Product{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete product: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("product not found")
	}
	return nil
}
