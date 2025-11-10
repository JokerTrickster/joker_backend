package usecase

import (
	"context"
	"fmt"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/interface"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/request"
)

type DeleteAlarmWeatherUseCase struct {
	Repository     _interface.IDeleteAlarmWeatherRepository
	ContextTimeout time.Duration
}

func NewDeleteAlarmWeatherUseCase(repo _interface.IDeleteAlarmWeatherRepository, timeout time.Duration) _interface.IDeleteAlarmWeatherUseCase {
	return &DeleteAlarmWeatherUseCase{Repository: repo, ContextTimeout: timeout}
}

func (d *DeleteAlarmWeatherUseCase) DeleteAlarm(c context.Context, userID int, req *request.ReqDeleteAlarm) error {
	ctx, cancel := context.WithTimeout(c, d.ContextTimeout)
	defer cancel()

	// 1. 알람 소유권 확인
	isOwner, err := d.Repository.CheckAlarmOwnership(ctx, userID, req.AlarmID)
	if err != nil {
		return fmt.Errorf("failed to check alarm ownership: %w", err)
	}

	if !isOwner {
		return fmt.Errorf("alarm not found or you don't have permission")
	}

	// 2. 알람 삭제
	if err := d.Repository.DeleteUserAlarm(ctx, userID, req.AlarmID); err != nil {
		return fmt.Errorf("failed to delete alarm: %w", err)
	}

	return nil
}
