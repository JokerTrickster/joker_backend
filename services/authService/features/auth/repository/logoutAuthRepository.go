package repository

import (
	"context"
	"fmt"

	_interface "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/interface"

	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
	"gorm.io/gorm"
)

func NewLogoutAuthRepository(gormDB *gorm.DB) _interface.ILogoutAuthRepository {
	return &LogoutAuthRepository{GormDB: gormDB}
}

func (d *LogoutAuthRepository) DeleteTokenByUserID(ctx context.Context, userID uint) error {
	result := d.GormDB.WithContext(ctx).Where("user_id = ?", userID).Delete(&mysql.Tokens{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no tokens found for user_id: %d", userID)
	}
	return nil
}
