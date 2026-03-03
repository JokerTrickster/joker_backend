package usecase

import (
	"context"
	"fmt"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/morandoranService/features/product/model/interface"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/product/model/request"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/product/model/response"
)

type UpdateUseCase struct {
	Repo           _interface.IProductRepository
	ContextTimeout time.Duration
}

func NewUpdateUseCase(repo _interface.IProductRepository, timeout time.Duration) *UpdateUseCase {
	return &UpdateUseCase{Repo: repo, ContextTimeout: timeout}
}

func (uc *UpdateUseCase) Update(c context.Context, id uint, req *request.ReqUpdateProduct) (*response.ResProductDetail, error) {
	ctx, cancel := context.WithTimeout(c, uc.ContextTimeout)
	defer cancel()

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.OriginalPrice != nil {
		updates["original_price"] = *req.OriginalPrice
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Image != nil {
		updates["image"] = *req.Image
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.Badge != nil {
		updates["badge"] = *req.Badge
	}
	if req.InStock != nil {
		updates["in_stock"] = *req.InStock
	}

	if len(updates) == 0 {
		p, err := uc.Repo.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}
		return &response.ResProductDetail{
			ID: fmt.Sprintf("product-%03d", p.ID), Name: p.Name, Price: p.Price, OriginalPrice: p.OriginalPrice,
			Description: p.Description, Image: p.Image, Category: p.Category,
			Badge: p.Badge, InStock: p.InStock, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
		}, nil
	}

	product, err := uc.Repo.Update(ctx, id, updates)
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
