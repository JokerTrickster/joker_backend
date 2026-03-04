package usecase

import (
	"context"
	"fmt"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/model/interface"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/model/response"
	"github.com/JokerTrickster/joker_backend/shared/jwt"
)

type OAuthUseCase struct {
	Repo           _interface.IAuthRepository
	ContextTimeout time.Duration
}

func NewOAuthUseCase(repo _interface.IAuthRepository, timeout time.Duration) *OAuthUseCase {
	return &OAuthUseCase{Repo: repo, ContextTimeout: timeout}
}

func (uc *OAuthUseCase) HandleCallback(c context.Context, email, nickname, provider string, profileImage *string) (*response.ResLogin, error) {
	ctx, cancel := context.WithTimeout(c, uc.ContextTimeout)
	defer cancel()

	user, err := uc.Repo.FindOrCreateByOAuth(ctx, email, nickname, provider, profileImage)
	if err != nil {
		return nil, fmt.Errorf("failed to find or create user: %w", err)
	}

	accessToken, _, _, _, err := jwt.GenerateToken(user.Email, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &response.ResLogin{
		Token: accessToken,
		User: response.ResUserInfo{
			ID:       fmt.Sprintf("user-%03d", user.ID),
			Nickname: user.Nickname,
			Email:    user.Email,
			Role:     user.Role,
		},
	}, nil
}
