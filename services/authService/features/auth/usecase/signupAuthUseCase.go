package usecase

import (
	"context"
	"fmt"
	_interface "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/interface"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/request"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/response"
	"github.com/JokerTrickster/joker_backend/shared/jwt"
	"time"
)

type SignupAuthUseCase struct {
	Repository     _interface.ISignupAuthRepository
	ContextTimeout time.Duration
}

func NewSignupAuthUseCase(repo _interface.ISignupAuthRepository, timeout time.Duration) _interface.ISignupAuthUseCase {
	return &SignupAuthUseCase{Repository: repo, ContextTimeout: timeout}
}

func (d *SignupAuthUseCase) Signup(c context.Context, req *request.ReqSignUp) (response.ResSignUp, error) {
	ctx, cancel := context.WithTimeout(c, d.ContextTimeout)
	defer cancel()

	// 유저 생성
	userID, err := d.Repository.CreateUser(ctx, req.Name, req.Email, req.Password, req.ServiceType)
	if err != nil {
		return response.ResSignUp{}, err
	}

	// JWT 토큰 발급
	accessToken, _, refreshToken, _, err := jwt.GenerateToken(req.Email, userID)
	if err != nil {
		return response.ResSignUp{}, fmt.Errorf("failed to generate tokens: %w", err)
	}

	res := response.ResSignUp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return res, nil
}
