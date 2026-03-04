package usecase

import (
	"context"
	"fmt"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/morandoranService/features/gallery/model/interface"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/gallery/model/response"
)

type DetailUseCase struct {
	Repo           _interface.IGalleryRepository
	ContextTimeout time.Duration
}

func NewDetailUseCase(repo _interface.IGalleryRepository, timeout time.Duration) *DetailUseCase {
	return &DetailUseCase{Repo: repo, ContextTimeout: timeout}
}

func (uc *DetailUseCase) Detail(c context.Context, id uint, userID *uint) (*response.ResGalleryDetail, error) {
	ctx, cancel := context.WithTimeout(c, uc.ContextTimeout)
	defer cancel()

	post, err := uc.Repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	nickname, _ := uc.Repo.GetAuthorNickname(ctx, post.AuthorID)

	isLiked := false
	if userID != nil {
		isLiked, _ = uc.Repo.IsLiked(ctx, *userID, id)
	}

	return &response.ResGalleryDetail{
		ID:           fmt.Sprintf("gallery-%03d", post.ID),
		Author:       response.ResAuthor{ID: fmt.Sprintf("user-%03d", post.AuthorID), Nickname: nickname},
		MediaType:    post.MediaType,
		MediaURL:     post.MediaURL,
		ThumbnailURL: post.ThumbnailURL,
		Caption:      post.Caption,
		LikeCount:    post.LikeCount,
		CommentCount: post.CommentCount,
		IsLiked:      isLiked,
		CreatedAt:    post.CreatedAt,
	}, nil
}
