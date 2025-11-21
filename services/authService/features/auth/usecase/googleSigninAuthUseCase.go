package usecase

import (
	"context"
	"fmt"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/interface"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/response"
	"github.com/JokerTrickster/joker_backend/shared/jwt"
)

type GoogleSigninAuthUseCase struct {
	Repository     _interface.IGoogleSigninAuthRepository
	ContextTimeout time.Duration
}

func NewGoogleSigninAuthUseCase(repo _interface.IGoogleSigninAuthRepository, timeout time.Duration) _interface.IGoogleSigninAuthUseCase {
	return &GoogleSigninAuthUseCase{Repository: repo, ContextTimeout: timeout}
}

func (d *GoogleSigninAuthUseCase) GoogleSignin(c context.Context) (response.ResGoogleSignin, error) {
	ctx, cancel := context.WithTimeout(c, d.ContextTimeout)
	defer cancel()
	fmt.Println(ctx)
	// JWT 토큰 발급
	accessToken, _, refreshToken, _, err := jwt.GenerateToken("tmp", 1)
	if err != nil {
		return response.ResGoogleSignin{}, fmt.Errorf("failed to generate tokens: %w", err)
	}

	res := response.ResGoogleSignin{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return res, nil
}
