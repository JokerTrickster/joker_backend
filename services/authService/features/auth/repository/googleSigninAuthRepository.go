package repository

import (
	_interface "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/interface"

	"gorm.io/gorm"
)

func NewGoogleSigninAuthRepository(gormDB *gorm.DB) _interface.IGoogleSigninAuthRepository {
	return &GoogleSigninAuthRepository{GormDB: gormDB}
}
