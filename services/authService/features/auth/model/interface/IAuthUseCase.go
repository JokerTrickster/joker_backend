package _interface

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/request"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/response"
)

type ISigninAuthUseCase interface {
	Signin(c context.Context, req *request.ReqSignIn) (response.ResSignIn, error)
}

type ISignupAuthUseCase interface {
	Signup(c context.Context, req *request.ReqSignUp) (response.ResSignUp, error)
}

type IRefreshTokenUseCase interface {
	RefreshToken(c context.Context, req *request.ReqRefreshToken) (response.ResRefreshToken, error)
}

type ILogoutAuthUseCase interface {
	Logout(c context.Context, userID uint) error
}
type ICheckEmailAuthUseCase interface {
	CheckEmail(ctx context.Context, req *request.ReqCheckEmail) (*response.ResCheckEmail, error)
}

type IGoogleSigninAuthUseCase interface {
	GoogleSignin(ctx context.Context, idToken string) (response.ResGoogleSignin, error)
}
