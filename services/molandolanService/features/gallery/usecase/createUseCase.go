package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/interface"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/request"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/response"
)

type CreateUseCase struct {
	Repo           _interface.IGalleryRepository
	ContextTimeout time.Duration
}

func NewCreateUseCase(repo _interface.IGalleryRepository, timeout time.Duration) *CreateUseCase {
	return &CreateUseCase{Repo: repo, ContextTimeout: timeout}
}

func (uc *CreateUseCase) Create(c context.Context, userID uint, req *request.ReqCreateGallery) (*response.ResGalleryCreate, error) {
	ctx, cancel := context.WithTimeout(c, uc.ContextTimeout)
	defer cancel()

	post := &entity.GalleryPost{
		AuthorID:     userID,
		MediaType:    req.MediaType,
		MediaURL:     req.MediaURL,
		ThumbnailURL: req.ThumbnailURL,
		Caption:      req.Caption,
	}

	created, err := uc.Repo.Create(ctx, post)
	if err != nil {
		return nil, err
	}

	nickname, _ := uc.Repo.GetAuthorNickname(ctx, userID)

	return &response.ResGalleryCreate{
		ID:           fmt.Sprintf("gallery-%03d", created.ID),
		Author:       response.ResAuthor{ID: fmt.Sprintf("user-%03d", userID), Nickname: nickname},
		MediaType:    created.MediaType,
		MediaURL:     created.MediaURL,
		ThumbnailURL: created.ThumbnailURL,
		Caption:      created.Caption,
		LikeCount:    0,
		CommentCount: 0,
		IsLiked:      false,
		CreatedAt:    created.CreatedAt,
	}, nil
}
