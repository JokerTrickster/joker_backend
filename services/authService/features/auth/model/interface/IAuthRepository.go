package _interface

import "context"

type ISigninAuthRepository interface {
	FindUserByEmail(c context.Context, email string, password string, serviceType string) error
}

type ISignupAuthRepository interface {
	SignupAuth(ctx context.Context, req interface{}) (interface{}, error)
}
