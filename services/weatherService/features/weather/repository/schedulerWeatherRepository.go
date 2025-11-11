package repository

import (
	"context"
	"fmt"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/interface"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
	"gorm.io/gorm"
)

type SchedulerWeatherRepository struct {
	GormDB *gorm.DB
}

func NewSchedulerWeatherRepository(gormDB *gorm.DB) _interface.ISchedulerWeatherRepository {
	return &SchedulerWeatherRepository{GormDB: gormDB}
}

// GetAlarmsToNotify retrieves alarms that need to be notified at the target time
// Filters by:
// - alarm_time matches target time (typically current_time + 1 minute)
// - is_enabled = true
// - deleted_at IS NULL
// - Duplicate prevention: last_sent IS NULL OR DATE(last_sent) < CURDATE()
func (r *SchedulerWeatherRepository) GetAlarmsToNotify(ctx context.Context, targetTime time.Time) ([]entity.UserAlarm, error) {
	var alarms []entity.UserAlarm

	// Format target time to HH:MM:SS
	targetTimeStr := targetTime.Format("15:04:05")

	// Use Go time calculation instead of database-specific CURDATE() for cross-DB compatibility
	todayStart := time.Now().Truncate(24 * time.Hour)

	if err := r.GormDB.WithContext(ctx).
		Where("alarm_time = ?", targetTimeStr).
		Where("is_enabled = ?", true).
		Where("deleted_at IS NULL").
		Where("last_sent IS NULL OR last_sent < ?", todayStart).
		Find(&alarms).Error; err != nil {
		return nil, fmt.Errorf("failed to get alarms to notify: %w", err)
	}

	return alarms, nil
}

// UpdateLastSent updates the last_sent timestamp after successful notification
// This prevents duplicate notifications on the same day
func (r *SchedulerWeatherRepository) UpdateLastSent(ctx context.Context, alarmID int, sentTime time.Time) error {
	result := r.GormDB.WithContext(ctx).
		Model(&entity.UserAlarm{}).
		Where("id = ?", alarmID).
		Update("last_sent", sentTime)

	if result.Error != nil {
		return fmt.Errorf("failed to update last_sent: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("alarm with id %d not found", alarmID)
	}

	return nil
}

// GetFCMTokens retrieves all active FCM tokens for a user
// Filters by:
// - user_id matches
// - deleted_at IS NULL
func (r *SchedulerWeatherRepository) GetFCMTokens(ctx context.Context, userID int) ([]entity.WeatherServiceToken, error) {
	var tokens []entity.WeatherServiceToken

	if err := r.GormDB.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Find(&tokens).Error; err != nil {
		return nil, fmt.Errorf("failed to get FCM tokens: %w", err)
	}

	return tokens, nil
}
