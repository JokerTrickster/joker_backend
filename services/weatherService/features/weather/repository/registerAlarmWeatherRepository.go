package repository

import (
	"context"
	"fmt"

	_interface "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/interface"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func NewRegisterAlarmWeatherRepository(gormDB *gorm.DB) _interface.IRegisterAlarmWeatherRepository {
	return &RegisterAlarmWeatherRepository{GormDB: gormDB}
}

// CreateOrUpdateFCMToken FCM 토큰을 생성하거나 업데이트합니다
func (r *RegisterAlarmWeatherRepository) CreateOrUpdateFCMToken(ctx context.Context, userID int, fcmToken string, deviceID string) error {
	token := &entity.WeatherServiceToken{
		UserID:   userID,
		FCMToken: fcmToken,
		DeviceID: deviceID,
	}

	// UPSERT: user_id와 device_id가 동일하면 fcm_token 업데이트
	if err := r.GormDB.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "device_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"fcm_token", "updated_at"}),
		}).
		Create(token).Error; err != nil {
		return fmt.Errorf("failed to create or update FCM token: %w", err)
	}

	return nil
}

// CreateUserAlarm 사용자 알람을 생성합니다
func (r *RegisterAlarmWeatherRepository) CreateUserAlarm(ctx context.Context, userID int, alarmTime string, region string) (*entity.UserAlarm, error) {
	alarm := &entity.UserAlarm{
		UserID:    userID,
		AlarmTime: alarmTime,
		Region:    region,
		IsEnabled: true,
	}

	if err := r.GormDB.WithContext(ctx).Create(alarm).Error; err != nil {
		return nil, fmt.Errorf("failed to create user alarm: %w", err)
	}

	return alarm, nil
}
