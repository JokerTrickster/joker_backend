package usecase

import (
	"context"
	"errors"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/interface"
)

var ErrForbidden = errors.New("forbidden")

type DeleteUseCase struct {
	Repo           _interface.IGalleryRepository
	ContextTimeout time.Duration
}

func NewDeleteUseCase(repo _interface.IGalleryRepository, timeout time.Duration) *DeleteUseCase {
	return &DeleteUseCase{Repo: repo, ContextTimeout: timeout}
}

func (uc *DeleteUseCase) Delete(c context.Context, id, userID uint, userRole string) error {
	ctx, cancel := context.WithTimeout(c, uc.ContextTimeout)
	defer cancel()

	post, err := uc.Repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if post.AuthorID != userID && userRole != "admin" {
		return ErrForbidden
	}

	return uc.Repo.Delete(ctx, id)
}
