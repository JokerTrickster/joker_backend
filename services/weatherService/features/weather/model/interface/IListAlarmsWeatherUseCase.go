package _interface

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/response"
)

type IListAlarmsWeatherUseCase interface {
	ListAlarms(ctx context.Context, userID int) (*response.ResListAlarms, error)
}
