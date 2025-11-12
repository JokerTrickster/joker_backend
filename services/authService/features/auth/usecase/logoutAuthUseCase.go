package usecase

import (
	"context"
	"fmt"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/interface"
)

type LogoutAuthUseCase struct {
	Repository     _interface.ILogoutAuthRepository
	ContextTimeout time.Duration
}

func NewLogoutAuthUseCase(repo _interface.ILogoutAuthRepository, timeout time.Duration) _interface.ILogoutAuthUseCase {
	return &LogoutAuthUseCase{Repository: repo, ContextTimeout: timeout}
}

func (d *LogoutAuthUseCase) Logout(c context.Context, userID uint) error {
	ctx, cancel := context.WithTimeout(c, d.ContextTimeout)
	defer cancel()
	fmt.Println(ctx)

	// 해당 유저 토큰 무효화 처리한다.
	err := d.Repository.DeleteTokenByUserID(ctx, userID)
	if err != nil {
		return err
	}

	return nil

}
