package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/model/interface"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/model/request"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/model/response"
)

type CreateUseCase struct {
	Repo           _interface.INewsRepository
	ContextTimeout time.Duration
}

func NewCreateUseCase(repo _interface.INewsRepository, timeout time.Duration) *CreateUseCase {
	return &CreateUseCase{Repo: repo, ContextTimeout: timeout}
}

func (uc *CreateUseCase) Create(c context.Context, req *request.ReqCreateNews) (*response.ResNewsDetail, error) {
	ctx, cancel := context.WithTimeout(c, uc.ContextTimeout)
	defer cancel()

	news := &entity.News{
		Title:     req.Title,
		Summary:   req.Summary,
		Content:   req.Content,
		Thumbnail: req.Thumbnail,
		Category:  req.Category,
		Date:      req.Date,
	}

	created, err := uc.Repo.Create(ctx, news)
	if err != nil {
		return nil, err
	}

	return &response.ResNewsDetail{
		ID:        fmt.Sprintf("news-%03d", created.ID),
		Title:     created.Title,
		Summary:   created.Summary,
		Content:   created.Content,
		Thumbnail: created.Thumbnail,
		Category:  created.Category,
		Date:      created.Date,
		CreatedAt: created.CreatedAt,
		UpdatedAt: created.UpdatedAt,
	}, nil
}
