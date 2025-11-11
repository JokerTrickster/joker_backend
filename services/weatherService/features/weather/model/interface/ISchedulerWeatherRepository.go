package _interface

import (
	"context"
	"time"

	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
)

type ISchedulerWeatherRepository interface {
	GetAlarmsToNotify(ctx context.Context, targetTime time.Time) ([]entity.UserAlarm, error)
	UpdateLastSent(ctx context.Context, alarmID int, sentTime time.Time) error
	GetFCMTokens(ctx context.Context, userID int) ([]entity.WeatherServiceToken, error)
}
