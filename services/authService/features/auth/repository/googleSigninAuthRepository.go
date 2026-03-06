package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
		user = &mysql.Users{
			Name:     name,
			Email:    email,
			Password: "",
			Provider: "google",
		}
		if err := r.GormDB.WithContext(ctx).Create(user).Error; err != nil {
			if strings.Contains(err.Error(), "Duplicate entry") {
				// Another request created this user concurrently; fetch it
				found := &mysql.Users{}
				if findErr := r.GormDB.WithContext(ctx).Where("email = ? AND provider = ?", email, "google").First(found).Error; findErr != nil {
					return 0, fmt.Errorf("failed to find concurrently created user: %w", findErr)
				}
				return uint(found.ID), nil
			}
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
