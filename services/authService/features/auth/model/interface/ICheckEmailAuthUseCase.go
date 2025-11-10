package _interface

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/request"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/response"
)

type ICheckEmailAuthUseCase interface {
	CheckEmail(ctx context.Context, req *request.ReqCheckEmail) (*response.ResCheckEmail, error)
}
