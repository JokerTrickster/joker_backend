package _interface

import (
	"context"
)

type IRegisterAlarmWeatherUseCase interface {
	RegisterAlarm(ctx context.Context) error
}
