package usecase

import (
	"context"
	"fmt"
	"math"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/interface"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/request"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/response"
)

type CommentListUseCase struct {
	Repo           _interface.IGalleryRepository
	ContextTimeout time.Duration
}

func NewCommentListUseCase(repo _interface.IGalleryRepository, timeout time.Duration) *CommentListUseCase {
	return &CommentListUseCase{Repo: repo, ContextTimeout: timeout}
}

func (uc *CommentListUseCase) List(c context.Context, galleryID uint, req *request.ReqListComment) (*response.ResCommentList, error) {
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

	items, total, err := uc.Repo.ListComments(ctx, galleryID, page, limit)
	if err != nil {
		return nil, err
	}

	resItems := make([]response.ResCommentItem, len(items))
	for i, item := range items {
		nickname, _ := uc.Repo.GetAuthorNickname(ctx, item.AuthorID)
		resItems[i] = response.ResCommentItem{
			ID:        fmt.Sprintf("comment-%03d", item.ID),
			Author:    response.ResAuthor{ID: fmt.Sprintf("user-%03d", item.AuthorID), Nickname: nickname},
			Content:   item.Content,
			CreatedAt: item.CreatedAt,
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &response.ResCommentList{
		Items: resItems,
		Pagination: response.ResPagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}
