package _interface

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/product/model/entity"
)

type IProductRepository interface {
	List(ctx context.Context, page, limit int, category string, inStock *bool) ([]entity.Product, int64, error)
	FindByID(ctx context.Context, id uint) (*entity.Product, error)
	Create(ctx context.Context, product *entity.Product) (*entity.Product, error)
	Update(ctx context.Context, id uint, updates map[string]interface{}) (*entity.Product, error)
	Delete(ctx context.Context, id uint) error
}
