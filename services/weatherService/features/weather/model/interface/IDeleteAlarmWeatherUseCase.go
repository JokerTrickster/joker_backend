package _interface

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/request"
)

type IDeleteAlarmWeatherUseCase interface {
	DeleteAlarm(ctx context.Context, userID int, req *request.ReqDeleteAlarm) error
}
