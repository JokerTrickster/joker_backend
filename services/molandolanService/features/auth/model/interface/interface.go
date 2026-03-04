package _interface

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/model/entity"
)

type IAuthRepository interface {
	FindUserByEmail(ctx context.Context, email string) (*entity.MorandoranUser, error)
	FindUserByID(ctx context.Context, userID uint) (*entity.MorandoranUser, error)
	FindOrCreateByOAuth(ctx context.Context, email, nickname, provider string, profileImage *string) (*entity.MorandoranUser, error)
	UpdateNickname(ctx context.Context, userID uint, nickname string) (*entity.MorandoranUser, error)
}
