package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/gallery/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/morandoranService/features/gallery/model/interface"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/gallery/model/request"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/gallery/model/response"
)

type CommentCreateUseCase struct {
	Repo           _interface.IGalleryRepository
	ContextTimeout time.Duration
}

func NewCommentCreateUseCase(repo _interface.IGalleryRepository, timeout time.Duration) *CommentCreateUseCase {
	return &CommentCreateUseCase{Repo: repo, ContextTimeout: timeout}
}

func (uc *CommentCreateUseCase) Create(c context.Context, galleryID, userID uint, req *request.ReqCreateComment) (*response.ResCommentItem, error) {
	ctx, cancel := context.WithTimeout(c, uc.ContextTimeout)
	defer cancel()

	comment := &entity.GalleryComment{
		GalleryID: galleryID,
		AuthorID:  userID,
		Content:   req.Content,
	}

	created, err := uc.Repo.CreateComment(ctx, comment)
	if err != nil {
		return nil, err
	}

	nickname, _ := uc.Repo.GetAuthorNickname(ctx, userID)

	return &response.ResCommentItem{
		ID:        fmt.Sprintf("comment-%03d", created.ID),
		Author:    response.ResAuthor{ID: fmt.Sprintf("user-%03d", userID), Nickname: nickname},
		Content:   created.Content,
		CreatedAt: created.CreatedAt,
	}, nil
}
