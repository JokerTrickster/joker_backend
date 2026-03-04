package usecase

import (
	"context"
	"fmt"
	"math"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/morandoranService/features/news/model/interface"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/news/model/request"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/news/model/response"
)

type ListUseCase struct {
	Repo           _interface.INewsRepository
	ContextTimeout time.Duration
}

func NewListUseCase(repo _interface.INewsRepository, timeout time.Duration) *ListUseCase {
	return &ListUseCase{Repo: repo, ContextTimeout: timeout}
}

func (uc *ListUseCase) List(c context.Context, req *request.ReqListNews) (*response.ResNewsList, error) {
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

	items, total, err := uc.Repo.List(ctx, page, limit, req.Category)
	if err != nil {
		return nil, err
	}

	resItems := make([]response.ResNewsItem, len(items))
	for i, item := range items {
		resItems[i] = response.ResNewsItem{
			ID:        fmt.Sprintf("news-%03d", item.ID),
			Title:     item.Title,
			Summary:   item.Summary,
			Thumbnail: item.Thumbnail,
			Category:  item.Category,
			Date:      item.Date,
			CreatedAt: item.CreatedAt,
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &response.ResNewsList{
		Items: resItems,
		Pagination: response.ResPagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}
