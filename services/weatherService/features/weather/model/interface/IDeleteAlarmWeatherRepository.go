package _interface

import (
	"context"
)

type IDeleteAlarmWeatherRepository interface {
	DeleteUserAlarm(ctx context.Context, userID int, alarmID int) error
	CheckAlarmOwnership(ctx context.Context, userID int, alarmID int) (bool, error)
}
