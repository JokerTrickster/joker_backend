package repository

import "gorm.io/gorm"

type SigninAuthRepository struct {
	GormDB *gorm.DB
}

type SignupAuthRepository struct {
	GormDB *gorm.DB
}
