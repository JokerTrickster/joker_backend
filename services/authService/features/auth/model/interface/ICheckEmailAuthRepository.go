package _interface

import "context"

type ICheckEmailAuthRepository interface {
	CheckEmailExists(ctx context.Context, email string, provider string) (bool, error)
}
