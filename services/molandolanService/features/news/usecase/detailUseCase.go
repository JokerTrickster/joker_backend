package usecase

import (
	"context"
	"fmt"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/model/interface"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/model/response"
)

type DetailUseCase struct {
	Repo           _interface.INewsRepository
	ContextTimeout time.Duration
}

func NewDetailUseCase(repo _interface.INewsRepository, timeout time.Duration) *DetailUseCase {
	return &DetailUseCase{Repo: repo, ContextTimeout: timeout}
}

func (uc *DetailUseCase) Detail(c context.Context, id uint) (*response.ResNewsDetail, error) {
	ctx, cancel := context.WithTimeout(c, uc.ContextTimeout)
	defer cancel()

	news, err := uc.Repo.FindByID(ctx, id)
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
