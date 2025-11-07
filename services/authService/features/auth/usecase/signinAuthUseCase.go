package usecase

import (
	"context"
	_interface "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/interface"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/request"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/response"
	"time"
)

type SigninAuthUseCase struct {
	Repository     _interface.ISigninAuthRepository
	ContextTimeout time.Duration
}

func NewSigninAuthUseCase(repo _interface.ISigninAuthRepository, timeout time.Duration) _interface.ISigninAuthUseCase {
	return &SigninAuthUseCase{Repository: repo, ContextTimeout: timeout}
}

func (d *SigninAuthUseCase) Signin(c context.Context, req *request.ReqSignIn) (response.ResSignIn, error) {
	ctx, cancel := context.WithTimeout(c, d.ContextTimeout)
	defer cancel()
	// 유저 조회 쿼리문 만든다.

	// 유저 정보로 조회
	err := d.Repository.FindUserByEmail(ctx, req.Email, req.Password, req.ServiceType)
	if err != nil {
		return response.ResSignIn{}, err
	}
	// 로그인 성공 토큰 발급

	accessToken := "access_token_example"   // 실제로는 JWT 토큰 등을 생성해야 함
	refreshToken := "refresh_token_example" // 실제로는 JWT 토큰 등을 생성해야 함

	res := response.ResSignIn{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return res, nil
}
