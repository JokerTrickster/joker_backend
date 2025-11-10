package _interface

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/request"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/response"
)

type IRegisterAlarmWeatherUseCase interface {
	RegisterAlarm(ctx context.Context, userID int, req *request.ReqRegisterAlarm) (*response.ResRegisterAlarm, error)
}
