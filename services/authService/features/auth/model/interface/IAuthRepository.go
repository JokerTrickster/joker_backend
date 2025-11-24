package _interface

import (
	"context"

	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
)

type ISigninAuthRepository interface {
	FindUserByEmail(c context.Context, email string, password string, serviceType string) (uint, string, error)
}

type ISignupAuthRepository interface {
	CreateUser(ctx context.Context, name string, email string, password string, provider string) (uint, error)
}

type ILogoutAuthRepository interface {
	DeleteTokenByUserID(ctx context.Context, userID uint) error
}

type ICheckEmailAuthRepository interface {
	CheckEmailExists(ctx context.Context, email string, provider string) (bool, error)
}

type IRefreshTokenAuthRepository interface {
	FindUserIDByRefreshToken(ctx context.Context, tokenDTO *mysql.Tokens) error
	FindOneByUserIDAndDeleteToken(ctx context.Context, userID uint) error
}

type IGoogleSigninAuthRepository interface{
	FindOrCreateUserByGoogleEmail(ctx context.Context, email string, name string) (uint, error)
}