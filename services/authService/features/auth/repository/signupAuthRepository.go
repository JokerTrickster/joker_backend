package repository

import (
	"context"
	"fmt"
	"strings"

	_interface "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/interface"
	"github.com/JokerTrickster/joker_backend/shared/db/mysql"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func NewSignupAuthRepository(gormDB *gorm.DB) _interface.ISignupAuthRepository {
	return &SignupAuthRepository{GormDB: gormDB}
}

func (r *SignupAuthRepository) CreateUser(ctx context.Context, name string, email string, password string, provider string) (uint, error) {
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
		if strings.Contains(err.Error(), "Duplicate entry") {
			return 0, fmt.Errorf("email already exists")
		}
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return uint(user.ID), nil
}
