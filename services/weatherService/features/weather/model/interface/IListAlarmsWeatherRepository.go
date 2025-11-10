package _interface

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
)

type IListAlarmsWeatherRepository interface {
	GetUserAlarms(ctx context.Context, userID int) ([]entity.UserAlarm, error)
}
