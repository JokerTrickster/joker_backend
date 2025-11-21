package repository

import (
	"context"

	_interface "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/interface"
	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
	"gorm.io/gorm"
)

func NewRefreshTokenAuthRepository(gormDB *gorm.DB) _interface.IRefreshTokenAuthRepository {
	return &RefreshTokenAuthRepository{GormDB: gormDB}
}

func (r *RefreshTokenAuthRepository) FindUserIDByRefreshToken(ctx context.Context, tokenDTO *mysql.Tokens) error {
	result := r.GormDB.WithContext(ctx).Create(tokenDTO)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *RefreshTokenAuthRepository) FindOneByUserIDAndDeleteToken(ctx context.Context, userID uint) error {
	result := r.GormDB.WithContext(ctx).Where("user_id = ?", userID).Delete(&mysql.Tokens{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}
