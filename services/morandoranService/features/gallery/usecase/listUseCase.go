package usecase

import (
	"context"
	"fmt"
	"math"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/morandoranService/features/gallery/model/interface"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/gallery/model/request"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/gallery/model/response"
)

type ListUseCase struct {
	Repo           _interface.IGalleryRepository
	ContextTimeout time.Duration
}

func NewListUseCase(repo _interface.IGalleryRepository, timeout time.Duration) *ListUseCase {
	return &ListUseCase{Repo: repo, ContextTimeout: timeout}
}

func (uc *ListUseCase) List(c context.Context, req *request.ReqListGallery) (*response.ResGalleryList, error) {
	ctx, cancel := context.WithTimeout(c, uc.ContextTimeout)
	defer cancel()

	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 30
	}

	items, total, err := uc.Repo.List(ctx, page, limit)
	if err != nil {
		return nil, err
	}

	resItems := make([]response.ResGalleryListItem, len(items))
	for i, item := range items {
		resItems[i] = response.ResGalleryListItem{
			ID:           fmt.Sprintf("gallery-%03d", item.ID),
			MediaType:    item.MediaType,
			ThumbnailURL: item.ThumbnailURL,
			LikeCount:    item.LikeCount,
			CommentCount: item.CommentCount,
			CreatedAt:    item.CreatedAt,
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &response.ResGalleryList{
		Items: resItems,
		Pagination: response.ResPagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}
