package _interface

import "context"

type ISigninAuthRepository interface {
	FindUserByEmail(c context.Context, email string, password string, serviceType string) (uint, string, error)
}

type ISignupAuthRepository interface {
	CreateUser(ctx context.Context, name string, email string, password string, provider string) (uint, error)
}
