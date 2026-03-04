package _interface

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/auth/model/entity"
)

type IAuthRepository interface {
	FindUserByEmail(ctx context.Context, email string) (*entity.MorandoranUser, error)
	FindUserByID(ctx context.Context, userID uint) (*entity.MorandoranUser, error)
}
