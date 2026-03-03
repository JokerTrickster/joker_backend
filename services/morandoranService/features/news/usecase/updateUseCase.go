package usecase

import (
	"context"
	"fmt"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/morandoranService/features/news/model/interface"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/news/model/request"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/news/model/response"
)

type UpdateUseCase struct {
	Repo           _interface.INewsRepository
	ContextTimeout time.Duration
}

func NewUpdateUseCase(repo _interface.INewsRepository, timeout time.Duration) *UpdateUseCase {
	return &UpdateUseCase{Repo: repo, ContextTimeout: timeout}
}

func (uc *UpdateUseCase) Update(c context.Context, id uint, req *request.ReqUpdateNews) (*response.ResNewsDetail, error) {
	ctx, cancel := context.WithTimeout(c, uc.ContextTimeout)
	defer cancel()

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Summary != nil {
		updates["summary"] = *req.Summary
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.Thumbnail != nil {
		updates["thumbnail"] = *req.Thumbnail
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.Date != nil {
		updates["date"] = *req.Date
	}

	if len(updates) == 0 {
		news, err := uc.Repo.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}
		return &response.ResNewsDetail{
			ID: fmt.Sprintf("news-%03d", news.ID), Title: news.Title, Summary: news.Summary,
			Content: news.Content, Thumbnail: news.Thumbnail,
			Category: news.Category, Date: news.Date,
			CreatedAt: news.CreatedAt, UpdatedAt: news.UpdatedAt,
		}, nil
	}

	news, err := uc.Repo.Update(ctx, id, updates)
	if err != nil {
		return nil, err
	}

	return &response.ResNewsDetail{
		ID:        fmt.Sprintf("news-%03d", news.ID),
		Title:     news.Title,
		Summary:   news.Summary,
		Content:   news.Content,
		Thumbnail: news.Thumbnail,
		Category:  news.Category,
		Date:      news.Date,
		CreatedAt: news.CreatedAt,
		UpdatedAt: news.UpdatedAt,
	}, nil
}
