package repository

import (
	"context"
	"fmt"

	_interface "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/interface"
	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
	"gorm.io/gorm"
)

type CheckEmailAuthRepository struct {
	GormDB *gorm.DB
}

func NewCheckEmailAuthRepository(gormDB *gorm.DB) _interface.ICheckEmailAuthRepository {
	return &CheckEmailAuthRepository{GormDB: gormDB}
}

// CheckEmailExists 이메일이 이미 등록되어 있는지 확인합니다
func (r *CheckEmailAuthRepository) CheckEmailExists(ctx context.Context, email string, provider string) (bool, error) {
	var count int64
	if err := r.GormDB.WithContext(ctx).
		Model(&mysql.Users{}).
		Where("email = ? AND provider = ?", email, provider).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return count > 0, nil
}
