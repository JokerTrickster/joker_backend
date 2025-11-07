package repository

import (
	"context"
	"errors"
	"fmt"
	_interface "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/interface"

	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
	"gorm.io/gorm"
)

func NewSigninAuthRepository(gormDB *gorm.DB) _interface.ISigninAuthRepository {
	return &SigninAuthRepository{GormDB: gormDB}
}

func (d *SigninAuthRepository) FindUserByEmail(c context.Context, email string, password string, serviceType string) error {
	user := &mysql.Users{}
	// 유저 정보로 조회
	result := d.GormDB.WithContext(c).
		Where("email = ? AND password = ? AND provider = ?", email, password, serviceType).
		First(user)
	// 에러 처리
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// 유저 없음
		return fmt.Errorf("user not found")
	}

	if result.Error != nil {
		// DB 조회 중 다른 에러 발생
		return result.Error
	}
	return nil
}
