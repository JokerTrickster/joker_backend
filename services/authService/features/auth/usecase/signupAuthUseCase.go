package usecase

import (
	"context"
	"fmt"
	_interface "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/interface"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/request"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/response"
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
	fmt.Println(ctx)

	return response.ResSignUp{}, nil
}
