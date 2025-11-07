package repository

import "gorm.io/gorm"

type SigninAuthRepository struct {
	GormDB *gorm.DB
}
