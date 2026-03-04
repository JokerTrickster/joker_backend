package usecase

import (
	"context"
	"fmt"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/interface"
)

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
		return fmt.Errorf("FORBIDDEN")
	}

	return uc.Repo.Delete(ctx, id)
}
