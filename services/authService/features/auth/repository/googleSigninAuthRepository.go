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
	
	// 구글 계정으로 유저 찾기
	result := r.GormDB.WithContext(ctx).
		Where("email = ? AND provider = ?", email, "google").
		First(user)
	
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// 유저가 없으면 생성
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
	
	return uint(user.ID), nil
}
