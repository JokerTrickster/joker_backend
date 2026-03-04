package usecase

import (
	"context"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/interface"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/response"
)

type LikeUseCase struct {
	Repo           _interface.IGalleryRepository
	ContextTimeout time.Duration
}

func NewLikeUseCase(repo _interface.IGalleryRepository, timeout time.Duration) *LikeUseCase {
	return &LikeUseCase{Repo: repo, ContextTimeout: timeout}
}

func (uc *LikeUseCase) ToggleLike(c context.Context, userID, galleryID uint) (*response.ResLikeToggle, error) {
	ctx, cancel := context.WithTimeout(c, uc.ContextTimeout)
	defer cancel()

	isLiked, likeCount, err := uc.Repo.ToggleLike(ctx, userID, galleryID)
	if err != nil {
		return nil, err
	}

	return &response.ResLikeToggle{
		IsLiked:   isLiked,
		LikeCount: likeCount,
	}, nil
}
