package usecase

import (
	"context"
	"fmt"
	"math"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/morandoranService/features/product/model/interface"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/product/model/request"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/product/model/response"
)

type ListUseCase struct {
	Repo           _interface.IProductRepository
	ContextTimeout time.Duration
}

func NewListUseCase(repo _interface.IProductRepository, timeout time.Duration) *ListUseCase {
	return &ListUseCase{Repo: repo, ContextTimeout: timeout}
}

func (uc *ListUseCase) List(c context.Context, req *request.ReqListProduct) (*response.ResProductList, error) {
	ctx, cancel := context.WithTimeout(c, uc.ContextTimeout)
	defer cancel()

	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 20
	}

	items, total, err := uc.Repo.List(ctx, page, limit, req.Category, req.InStock)
	if err != nil {
		return nil, err
	}

	resItems := make([]response.ResProductItem, len(items))
	for i, item := range items {
		resItems[i] = response.ResProductItem{
			ID:            fmt.Sprintf("product-%03d", item.ID),
			Name:          item.Name,
			Price:         item.Price,
			OriginalPrice: item.OriginalPrice,
			Description:   item.Description,
			Image:         item.Image,
			Category:      item.Category,
			Badge:         item.Badge,
			InStock:       item.InStock,
			CreatedAt:     item.CreatedAt,
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &response.ResProductList{
		Items: resItems,
		Pagination: response.ResPagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}
