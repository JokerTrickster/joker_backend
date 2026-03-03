package usecase

import (
	"context"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/morandoranService/features/ranking/model/interface"
)

type DeleteUseCase struct {
	Repo           _interface.IRankingRepository
	ContextTimeout time.Duration
}

func NewDeleteUseCase(repo _interface.IRankingRepository, timeout time.Duration) *DeleteUseCase {
	return &DeleteUseCase{Repo: repo, ContextTimeout: timeout}
}

func (uc *DeleteUseCase) Delete(c context.Context, id uint) error {
	ctx, cancel := context.WithTimeout(c, uc.ContextTimeout)
	defer cancel()
	return uc.Repo.Delete(ctx, id)
}
