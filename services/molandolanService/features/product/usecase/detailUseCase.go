package usecase

import (
	"context"
	"fmt"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/product/model/interface"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/product/model/response"
)

type DetailUseCase struct {
	Repo           _interface.IProductRepository
	ContextTimeout time.Duration
}

func NewDetailUseCase(repo _interface.IProductRepository, timeout time.Duration) *DetailUseCase {
	return &DetailUseCase{Repo: repo, ContextTimeout: timeout}
}

func (uc *DetailUseCase) Detail(c context.Context, id uint) (*response.ResProductDetail, error) {
	ctx, cancel := context.WithTimeout(c, uc.ContextTimeout)
	defer cancel()

	product, err := uc.Repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &response.ResProductDetail{
		ID:            fmt.Sprintf("product-%03d", product.ID),
		Name:          product.Name,
		Price:         product.Price,
		OriginalPrice: product.OriginalPrice,
		Description:   product.Description,
		Image:         product.Image,
		Category:      product.Category,
		Badge:         product.Badge,
		InStock:       product.InStock,
		CreatedAt:     product.CreatedAt,
		UpdatedAt:     product.UpdatedAt,
	}, nil
}
