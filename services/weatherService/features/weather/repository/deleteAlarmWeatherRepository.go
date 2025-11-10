package repository

import (
	"context"
	"fmt"

	_interface "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/interface"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
	"gorm.io/gorm"
)

type DeleteAlarmWeatherRepository struct {
	GormDB *gorm.DB
}

func NewDeleteAlarmWeatherRepository(gormDB *gorm.DB) _interface.IDeleteAlarmWeatherRepository {
	return &DeleteAlarmWeatherRepository{GormDB: gormDB}
}

// CheckAlarmOwnership 알람이 해당 유저의 것인지 확인합니다
func (r *DeleteAlarmWeatherRepository) CheckAlarmOwnership(ctx context.Context, userID int, alarmID int) (bool, error) {
	var count int64
	if err := r.GormDB.WithContext(ctx).
		Model(&entity.UserAlarm{}).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", alarmID, userID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check alarm ownership: %w", err)
	}

	return count > 0, nil
}

// DeleteUserAlarm 사용자 알람을 삭제합니다 (soft delete)
func (r *DeleteAlarmWeatherRepository) DeleteUserAlarm(ctx context.Context, userID int, alarmID int) error {
	result := r.GormDB.WithContext(ctx).
		Where("id = ? AND user_id = ?", alarmID, userID).
		Delete(&entity.UserAlarm{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete alarm: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("alarm not found or already deleted")
	}

	return nil
}
