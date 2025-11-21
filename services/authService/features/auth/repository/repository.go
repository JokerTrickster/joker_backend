package repository

import "gorm.io/gorm"

type SigninAuthRepository struct {
	GormDB *gorm.DB
}

type SignupAuthRepository struct {
	GormDB *gorm.DB
}

type LogoutAuthRepository struct {
	GormDB *gorm.DB
}

type CheckEmailAuthRepository struct {
	GormDB *gorm.DB
}

type RefreshTokenAuthRepository struct {
	GormDB *gorm.DB
}

type GoogleSigninAuthRepository struct {
	GormDB *gorm.DB
}
