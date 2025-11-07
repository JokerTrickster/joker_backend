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
