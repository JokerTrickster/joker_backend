package usecase

import (
	"context"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/interface"
)

type RegisterAlarmWeatherUseCase struct {
	Repository     _interface.IRegisterAlarmWeatherRepository
	ContextTimeout time.Duration
}

func NewRegisterAlarmWeatherUseCase(repo _interface.IRegisterAlarmWeatherRepository, timeout time.Duration) _interface.IRegisterAlarmWeatherUseCase {
	return &RegisterAlarmWeatherUseCase{Repository: repo, ContextTimeout: timeout}
}

func (d *RegisterAlarmWeatherUseCase) RegisterAlarm(c context.Context) error {
	ctx, cancel := context.WithTimeout(c, d.ContextTimeout)
	defer cancel()
	_ = ctx

	return nil
}
