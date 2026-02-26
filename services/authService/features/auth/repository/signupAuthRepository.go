package repository

import (
	"context"
	"fmt"

	_interface "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/interface"
	"github.com/JokerTrickster/joker_backend/shared/db/mysql"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func NewSignupAuthRepository(gormDB *gorm.DB) _interface.ISignupAuthRepository {
	return &SignupAuthRepository{GormDB: gormDB}
}

func (r *SignupAuthRepository) CreateUser(ctx context.Context, name string, email string, password string, provider string) (uint, error) {
	var count int64
	if err := r.GormDB.WithContext(ctx).
		Model(&mysql.Users{}).
		Where("email = ? AND provider = ?", email, provider).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to check email duplication: %w", err)
	}

	if count > 0 {
		return 0, fmt.Errorf("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &mysql.Users{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
		Provider: provider,
	}

	if err := r.GormDB.WithContext(ctx).Create(user).Error; err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return uint(user.ID), nil
}
