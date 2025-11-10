package usecase

import (
	"context"
	"fmt"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/interface"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/response"
)

type ListAlarmsWeatherUseCase struct {
	Repository     _interface.IListAlarmsWeatherRepository
	ContextTimeout time.Duration
}

func NewListAlarmsWeatherUseCase(repo _interface.IListAlarmsWeatherRepository, timeout time.Duration) _interface.IListAlarmsWeatherUseCase {
	return &ListAlarmsWeatherUseCase{Repository: repo, ContextTimeout: timeout}
}

func (d *ListAlarmsWeatherUseCase) ListAlarms(c context.Context, userID int) (*response.ResListAlarms, error) {
	ctx, cancel := context.WithTimeout(c, d.ContextTimeout)
	defer cancel()

	// 1. 사용자 알람 목록 조회
	alarms, err := d.Repository.GetUserAlarms(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user alarms: %w", err)
	}

	// 2. Response 변환
	alarmItems := make([]response.AlarmItem, len(alarms))
	for i, alarm := range alarms {
		alarmItems[i] = response.AlarmItem{
			AlarmID:   alarm.ID,
			AlarmTime: alarm.AlarmTime,
			Region:    alarm.Region,
			IsEnabled: alarm.IsEnabled,
		}

		// LastSent가 있으면 ISO 8601 형식으로 변환
		if alarm.LastSent != nil {
			alarmItems[i].LastSent = alarm.LastSent.Format(time.RFC3339)
		}
	}

	return &response.ResListAlarms{
		Alarms: alarmItems,
		Total:  len(alarmItems),
	}, nil
}
