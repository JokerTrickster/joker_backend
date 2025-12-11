package repository

import (
	"context"
	"errors"
	"fmt"

	_interface "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/interface"
	"github.com/JokerTrickster/joker_backend/shared/db/mysql"

	"gorm.io/gorm"
)

func NewGoogleSigninAuthRepository(gormDB *gorm.DB) _interface.IGoogleSigninAuthRepository {
	return &GoogleSigninAuthRepository{GormDB: gormDB}
}

// FindOrCreateUserByGoogleEmail 구글 이메일로 유저를 찾거나 생성합니다
func (r *GoogleSigninAuthRepository) FindOrCreateUserByGoogleEmail(ctx context.Context, email string, name string) (uint, error) {
	user := &mysql.Users{}

	// Query including soft-deleted users using Unscoped()
	result := r.GormDB.Unscoped().WithContext(ctx).
		Where("email = ? AND provider = ?", email, "google").
		First(user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// User doesn't exist - create new
		user = &mysql.Users{
			Name:     name,
			Email:    email,
			Password: "", // 구글 로그인은 비밀번호 없음
			Provider: "google",
		}
		if err := r.GormDB.WithContext(ctx).Create(user).Error; err != nil {
			return 0, fmt.Errorf("failed to create user: %w", err)
		}
		return uint(user.ID), nil
	}

	if result.Error != nil {
		return 0, fmt.Errorf("failed to find user: %w", result.Error)
	}

	// If user was soft-deleted, restore it
	if user.DeletedAt.Valid {
		user.DeletedAt = gorm.DeletedAt{}
		if err := r.GormDB.Unscoped().WithContext(ctx).Save(user).Error; err != nil {
			return 0, fmt.Errorf("failed to restore user: %w", err)
		}
	}

	return uint(user.ID), nil
}
