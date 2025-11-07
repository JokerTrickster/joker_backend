package _interface

import "context"

type ISigninAuthUseCase interface {
	Signin(c context.Context) error
}
