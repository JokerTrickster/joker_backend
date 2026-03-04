package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/product/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/product/model/interface"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/product/model/request"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/product/model/response"
)

type CreateUseCase struct {
	Repo           _interface.IProductRepository
	ContextTimeout time.Duration
}

func NewCreateUseCase(repo _interface.IProductRepository, timeout time.Duration) *CreateUseCase {
	return &CreateUseCase{Repo: repo, ContextTimeout: timeout}
}

func (uc *CreateUseCase) Create(c context.Context, req *request.ReqCreateProduct) (*response.ResProductDetail, error) {
	ctx, cancel := context.WithTimeout(c, uc.ContextTimeout)
	defer cancel()

	inStock := true
	if req.InStock != nil {
		inStock = *req.InStock
	}

	product := &entity.Product{
		Name:          req.Name,
		Price:         req.Price,
		OriginalPrice: req.OriginalPrice,
		Description:   req.Description,
		Image:         req.Image,
		Category:      req.Category,
		Badge:         req.Badge,
		InStock:       inStock,
	}

	created, err := uc.Repo.Create(ctx, product)
	if err != nil {
		return nil, err
	}

	return &response.ResProductDetail{
		ID:            fmt.Sprintf("product-%03d", created.ID),
		Name:          created.Name,
		Price:         created.Price,
		OriginalPrice: created.OriginalPrice,
		Description:   created.Description,
		Image:         created.Image,
		Category:      created.Category,
		Badge:         created.Badge,
		InStock:       created.InStock,
		CreatedAt:     created.CreatedAt,
		UpdatedAt:     created.UpdatedAt,
	}, nil
}
