package usecase

import (
	"context"
	"fmt"
	"regexp"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/interface"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/request"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/response"
)

type RegisterAlarmWeatherUseCase struct {
	Repository     _interface.IRegisterAlarmWeatherRepository
	ContextTimeout time.Duration
}

func NewRegisterAlarmWeatherUseCase(repo _interface.IRegisterAlarmWeatherRepository, timeout time.Duration) _interface.IRegisterAlarmWeatherUseCase {
	return &RegisterAlarmWeatherUseCase{Repository: repo, ContextTimeout: timeout}
}

func (d *RegisterAlarmWeatherUseCase) RegisterAlarm(c context.Context, userID int, req *request.ReqRegisterAlarm) (*response.ResRegisterAlarm, error) {
	ctx, cancel := context.WithTimeout(c, d.ContextTimeout)
	defer cancel()

	// 1. 알람 시간 형식 검증 (HH:MM 또는 HH:MM:SS)
	alarmTime, err := d.validateAndFormatTime(req.AlarmTime)
	if err != nil {
		return nil, fmt.Errorf("invalid alarm time format: %w", err)
	}

	// 2. FCM 토큰 저장/업데이트
	if err := d.Repository.CreateOrUpdateFCMToken(ctx, userID, req.FCMToken, req.DeviceID); err != nil {
		return nil, fmt.Errorf("failed to save FCM token: %w", err)
	}

	// 3. 사용자 알람 생성
	alarm, err := d.Repository.CreateUserAlarm(ctx, userID, alarmTime, req.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to create alarm: %w", err)
	}

	// 4. 응답 생성
	return &response.ResRegisterAlarm{
		AlarmID:   alarm.ID,
		AlarmTime: alarm.AlarmTime,
		Region:    alarm.Region,
		IsEnabled: alarm.IsEnabled,
	}, nil
}

// validateAndFormatTime 시간 형식을 검증하고 HH:MM:SS 형식으로 변환합니다
func (d *RegisterAlarmWeatherUseCase) validateAndFormatTime(timeStr string) (string, error) {
	// HH:MM 형식
	timePattern := regexp.MustCompile(`^([01]\d|2[0-3]):([0-5]\d)$`)
	// HH:MM:SS 형식
	timeWithSecondsPattern := regexp.MustCompile(`^([01]\d|2[0-3]):([0-5]\d):([0-5]\d)$`)

	if timeWithSecondsPattern.MatchString(timeStr) {
		return timeStr, nil
	}

	if timePattern.MatchString(timeStr) {
		return timeStr + ":00", nil
	}

	return "", fmt.Errorf("time must be in HH:MM or HH:MM:SS format")
}
