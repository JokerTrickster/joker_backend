package repository

import (
	"context"
	"errors"
	"fmt"

	_interface "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/interface"
	"github.com/JokerTrickster/joker_backend/shared/db/mysql"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func NewSigninAuthRepository(gormDB *gorm.DB) _interface.ISigninAuthRepository {
	return &SigninAuthRepository{GormDB: gormDB}
}

func (d *SigninAuthRepository) FindUserByEmail(c context.Context, email string, password string, serviceType string) (uint, string, error) {
	user := &mysql.Users{}
	result := d.GormDB.WithContext(c).
		Where("email = ? AND provider = ?", email, serviceType).
		First(user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return 0, "", fmt.Errorf("user not found")
	}
	if result.Error != nil {
		return 0, "", result.Error
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return 0, "", fmt.Errorf("password not match")
	}

	return uint(user.ID), user.Email, nil
}
