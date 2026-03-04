package usecase

import (
	"context"
	"fmt"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/morandoranService/features/auth/model/interface"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/auth/model/request"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/auth/model/response"
	"github.com/JokerTrickster/joker_backend/shared/jwt"
	"golang.org/x/crypto/bcrypt"
)

type LoginUseCase struct {
	Repo           _interface.IAuthRepository
	ContextTimeout time.Duration
}

func NewLoginUseCase(repo _interface.IAuthRepository, timeout time.Duration) *LoginUseCase {
	return &LoginUseCase{Repo: repo, ContextTimeout: timeout}
}

func (uc *LoginUseCase) Login(c context.Context, req *request.ReqLogin) (*response.ResLogin, error) {
	ctx, cancel := context.WithTimeout(c, uc.ContextTimeout)
	defer cancel()

	user, err := uc.Repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid email or password")
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
