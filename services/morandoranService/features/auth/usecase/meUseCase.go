package usecase

import (
	"context"
	"fmt"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/morandoranService/features/auth/model/interface"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/auth/model/response"
)

type MeUseCase struct {
	Repo           _interface.IAuthRepository
	ContextTimeout time.Duration
}

func NewMeUseCase(repo _interface.IAuthRepository, timeout time.Duration) *MeUseCase {
	return &MeUseCase{Repo: repo, ContextTimeout: timeout}
}

func (uc *MeUseCase) Me(c context.Context, userID uint) (*response.ResMe, error) {
	ctx, cancel := context.WithTimeout(c, uc.ContextTimeout)
	defer cancel()

	user, err := uc.Repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return &response.ResMe{
		ID:        fmt.Sprintf("user-%03d", user.ID),
		Nickname:  user.Nickname,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}, nil
}
