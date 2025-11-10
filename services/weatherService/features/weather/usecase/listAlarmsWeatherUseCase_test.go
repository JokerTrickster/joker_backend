package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/request"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupTestDBForList(t *testing.T) *gorm.DB {
	dsn := "root:rootpassword@tcp(localhost:3307)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	return db
}

func TestListAlarmsWeatherUseCase_Success(t *testing.T) {
	db := setupTestDBForList(t)

	// 테스트 알람 생성
	registerRepo := repository.NewRegisterAlarmWeatherRepository(db)
	registerUC := NewRegisterAlarmWeatherUseCase(registerRepo, 10*time.Second)

	ctx := context.Background()
	userID := 1

	// 3개의 알람 등록
	alarms := []request.ReqRegisterAlarm{
		{
			AlarmTime: "07:00",
			Region:    "서울시 강남구",
			FCMToken:  "test-fcm-token-list-1",
			DeviceID:  "test-device-list-1",
		},
		{
			AlarmTime: "12:00",
			Region:    "서울시 서초구",
			FCMToken:  "test-fcm-token-list-2",
			DeviceID:  "test-device-list-2",
		},
		{
			AlarmTime: "18:30",
			Region:    "서울시 강서구",
			FCMToken:  "test-fcm-token-list-3",
			DeviceID:  "test-device-list-3",
		},
	}

	for _, alarm := range alarms {
		_, err := registerUC.RegisterAlarm(ctx, userID, &alarm)
		assert.NoError(t, err)
	}

	// 알람 목록 조회
	listRepo := repository.NewListAlarmsWeatherRepository(db)
	listUC := NewListAlarmsWeatherUseCase(listRepo, 10*time.Second)

	res, err := listUC.ListAlarms(ctx, userID)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.GreaterOrEqual(t, res.Total, 3)
	assert.GreaterOrEqual(t, len(res.Alarms), 3)

	// 최소 3개의 알람이 있는지 확인
	t.Logf("Found %d alarms for user %d", res.Total, userID)

	for i, alarm := range res.Alarms {
		if i < 3 {
			t.Logf("Alarm %d: ID=%d, Time=%s, Region=%s, Enabled=%v",
				i+1, alarm.AlarmID, alarm.AlarmTime, alarm.Region, alarm.IsEnabled)
		}
	}
}

func TestListAlarmsWeatherUseCase_EmptyList(t *testing.T) {
	db := setupTestDBForList(t)

	listRepo := repository.NewListAlarmsWeatherRepository(db)
	listUC := NewListAlarmsWeatherUseCase(listRepo, 10*time.Second)

	ctx := context.Background()
	// 존재하지 않는 사용자 또는 알람이 없는 사용자
	userID := 9999

	res, err := listUC.ListAlarms(ctx, userID)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 0, res.Total)
	assert.Equal(t, 0, len(res.Alarms))

	t.Logf("Empty list result: Total=%d", res.Total)
}

func TestListAlarmsWeatherUseCase_OnlyActiveAlarms(t *testing.T) {
	db := setupTestDBForList(t)

	// 알람 생성
	registerRepo := repository.NewRegisterAlarmWeatherRepository(db)
	registerUC := NewRegisterAlarmWeatherUseCase(registerRepo, 10*time.Second)

	ctx := context.Background()
	userID := 1

	regReq := &request.ReqRegisterAlarm{
		AlarmTime: "09:00",
		Region:    "서울시 용산구",
		FCMToken:  "test-fcm-token-active",
		DeviceID:  "test-device-active",
	}

	regRes, err := registerUC.RegisterAlarm(ctx, userID, regReq)
	assert.NoError(t, err)
	createdAlarmID := regRes.AlarmID

	// 알람 삭제 (soft delete)
	deleteRepo := repository.NewDeleteAlarmWeatherRepository(db)
	deleteUC := NewDeleteAlarmWeatherUseCase(deleteRepo, 10*time.Second)

	delReq := &request.ReqDeleteAlarm{
		AlarmID: createdAlarmID,
	}

	err = deleteUC.DeleteAlarm(ctx, userID, delReq)
	assert.NoError(t, err)

	// 목록 조회 - 삭제된 알람은 나오지 않아야 함
	listRepo := repository.NewListAlarmsWeatherRepository(db)
	listUC := NewListAlarmsWeatherUseCase(listRepo, 10*time.Second)

	res, err := listUC.ListAlarms(ctx, userID)
	assert.NoError(t, err)

	// 삭제된 알람이 목록에 없는지 확인
	for _, alarm := range res.Alarms {
		assert.NotEqual(t, createdAlarmID, alarm.AlarmID, "Deleted alarm should not appear in the list")
	}

	t.Logf("Deleted alarm ID %d is not in the list (Total alarms: %d)", createdAlarmID, res.Total)
}
