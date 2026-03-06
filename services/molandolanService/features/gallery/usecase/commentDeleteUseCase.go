package usecase

import (
	"context"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/interface"
)

type CommentDeleteUseCase struct {
	Repo           _interface.IGalleryRepository
	ContextTimeout time.Duration
}

func NewCommentDeleteUseCase(repo _interface.IGalleryRepository, timeout time.Duration) *CommentDeleteUseCase {
	return &CommentDeleteUseCase{Repo: repo, ContextTimeout: timeout}
}

func (uc *CommentDeleteUseCase) Delete(c context.Context, commentID, userID uint, userRole string) error {
	ctx, cancel := context.WithTimeout(c, uc.ContextTimeout)
	defer cancel()

	comment, err := uc.Repo.FindCommentByID(ctx, commentID)
	if err != nil {
		return err
	}

	if comment.AuthorID != userID && userRole != "admin" {
		return ErrForbidden
	}

	return uc.Repo.DeleteComment(ctx, commentID)
}
