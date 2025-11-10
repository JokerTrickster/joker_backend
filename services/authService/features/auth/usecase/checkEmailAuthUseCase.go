package usecase

import (
	"context"
	"fmt"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/interface"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/request"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/response"
)

type CheckEmailAuthUseCase struct {
	Repository     _interface.ICheckEmailAuthRepository
	ContextTimeout time.Duration
}

func NewCheckEmailAuthUseCase(repo _interface.ICheckEmailAuthRepository, timeout time.Duration) _interface.ICheckEmailAuthUseCase {
	return &CheckEmailAuthUseCase{Repository: repo, ContextTimeout: timeout}
}

func (u *CheckEmailAuthUseCase) CheckEmail(c context.Context, req *request.ReqCheckEmail) (*response.ResCheckEmail, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// 이메일 중복 체크
	exists, err := u.Repository.CheckEmailExists(ctx, req.Email, req.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}

	return &response.ResCheckEmail{
		Email:     req.Email,
		Exists:    exists,
		Available: !exists,
	}, nil
}
