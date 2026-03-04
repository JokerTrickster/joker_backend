package _interface

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/model/entity"
)

type INewsRepository interface {
	List(ctx context.Context, page, limit int, category string) ([]entity.News, int64, error)
	FindByID(ctx context.Context, id uint) (*entity.News, error)
	Create(ctx context.Context, news *entity.News) (*entity.News, error)
	Update(ctx context.Context, id uint, updates map[string]interface{}) (*entity.News, error)
	Delete(ctx context.Context, id uint) error
}
