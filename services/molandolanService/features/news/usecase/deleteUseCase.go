package usecase

import (
	"context"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/morandoranService/features/news/model/interface"
)

type DeleteUseCase struct {
	Repo           _interface.INewsRepository
	ContextTimeout time.Duration
}

func NewDeleteUseCase(repo _interface.INewsRepository, timeout time.Duration) *DeleteUseCase {
	return &DeleteUseCase{Repo: repo, ContextTimeout: timeout}
}

func (uc *DeleteUseCase) Delete(c context.Context, id uint) error {
	ctx, cancel := context.WithTimeout(c, uc.ContextTimeout)
	defer cancel()
	return uc.Repo.Delete(ctx, id)
}
