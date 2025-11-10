package _interface

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
)

type IRegisterAlarmWeatherRepository interface {
	CreateOrUpdateFCMToken(ctx context.Context, userID int, fcmToken string, deviceID string) error
	CreateUserAlarm(ctx context.Context, userID int, alarmTime string, region string) (*entity.UserAlarm, error)
}
