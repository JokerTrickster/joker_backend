package usecase

import (
	"context"
	"fmt"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/model/interface"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/model/response"
)

type UpdateMeUseCase struct {
	Repo           _interface.IAuthRepository
	ContextTimeout time.Duration
}

func NewUpdateMeUseCase(repo _interface.IAuthRepository, timeout time.Duration) *UpdateMeUseCase {
	return &UpdateMeUseCase{Repo: repo, ContextTimeout: timeout}
}

func (uc *UpdateMeUseCase) UpdateNickname(c context.Context, userID uint, nickname string) (*response.ResMe, error) {
	ctx, cancel := context.WithTimeout(c, uc.ContextTimeout)
	defer cancel()

	user, err := uc.Repo.UpdateNickname(ctx, userID, nickname)
	if err != nil {
		return nil, fmt.Errorf("failed to update nickname: %w", err)
	}

	return &response.ResMe{
		ID:           fmt.Sprintf("user-%03d", user.ID),
		Nickname:     user.Nickname,
		Email:        user.Email,
		ProfileImage: user.ProfileImage,
		Role:         user.Role,
		Provider:     user.Provider,
		CreatedAt:    user.CreatedAt,
	}, nil
}
