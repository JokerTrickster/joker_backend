package repository

import (
	"context"
	"fmt"

	_interface "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/interface"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
	"gorm.io/gorm"
)

type ListAlarmsWeatherRepository struct {
	GormDB *gorm.DB
}

func NewListAlarmsWeatherRepository(gormDB *gorm.DB) _interface.IListAlarmsWeatherRepository {
	return &ListAlarmsWeatherRepository{GormDB: gormDB}
}

// GetUserAlarms 사용자의 모든 알람을 조회합니다
func (r *ListAlarmsWeatherRepository) GetUserAlarms(ctx context.Context, userID int) ([]entity.UserAlarm, error) {
	var alarms []entity.UserAlarm

	if err := r.GormDB.WithContext(ctx).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		Find(&alarms).Error; err != nil {
		return nil, fmt.Errorf("failed to get user alarms: %w", err)
	}

	return alarms, nil
}
