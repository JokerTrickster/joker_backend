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

func setupTestDBForDelete(t *testing.T) *gorm.DB {
	dsn := "root:rootpassword@tcp(localhost:3307)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	return db
}

func TestDeleteAlarmWeatherUseCase_Success(t *testing.T) {
	db := setupTestDBForDelete(t)

	// 먼저 알람을 생성
	registerRepo := repository.NewRegisterAlarmWeatherRepository(db)
	registerUC := NewRegisterAlarmWeatherUseCase(registerRepo, 10*time.Second)

	ctx := context.Background()
	userID := 1

	regReq := &request.ReqRegisterAlarm{
		AlarmTime: "11:00",
		Region:    "서울시 강남구",
		FCMToken:  "test-fcm-token-delete",
		DeviceID:  "test-device-delete",
	}

	regRes, err := registerUC.RegisterAlarm(ctx, userID, regReq)
	assert.NoError(t, err)
	assert.NotNil(t, regRes)
	alarmID := regRes.AlarmID

	// 생성된 알람 삭제
	deleteRepo := repository.NewDeleteAlarmWeatherRepository(db)
	deleteUC := NewDeleteAlarmWeatherUseCase(deleteRepo, 10*time.Second)

	delReq := &request.ReqDeleteAlarm{
		AlarmID: alarmID,
	}

	err = deleteUC.DeleteAlarm(ctx, userID, delReq)
	assert.NoError(t, err)

	t.Logf("Successfully deleted alarm ID: %d", alarmID)
}

func TestDeleteAlarmWeatherUseCase_NotFound(t *testing.T) {
	db := setupTestDBForDelete(t)

	deleteRepo := repository.NewDeleteAlarmWeatherRepository(db)
	deleteUC := NewDeleteAlarmWeatherUseCase(deleteRepo, 10*time.Second)

	ctx := context.Background()
	userID := 1

	delReq := &request.ReqDeleteAlarm{
		AlarmID: 999999, // 존재하지 않는 알람 ID
	}

	err := deleteUC.DeleteAlarm(ctx, userID, delReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	t.Logf("Expected error: %v", err)
}

func TestDeleteAlarmWeatherUseCase_WrongUser(t *testing.T) {
	db := setupTestDBForDelete(t)

	// User 1이 알람 생성
	registerRepo := repository.NewRegisterAlarmWeatherRepository(db)
	registerUC := NewRegisterAlarmWeatherUseCase(registerRepo, 10*time.Second)

	ctx := context.Background()
	user1ID := 1

	regReq := &request.ReqRegisterAlarm{
		AlarmTime: "12:00",
		Region:    "서울시 종로구",
		FCMToken:  "test-fcm-token-wrong-user",
		DeviceID:  "test-device-wrong-user",
	}

	regRes, err := registerUC.RegisterAlarm(ctx, user1ID, regReq)
	assert.NoError(t, err)
	alarmID := regRes.AlarmID

	// User 2가 삭제 시도
	deleteRepo := repository.NewDeleteAlarmWeatherRepository(db)
	deleteUC := NewDeleteAlarmWeatherUseCase(deleteRepo, 10*time.Second)

	user2ID := 2
	delReq := &request.ReqDeleteAlarm{
		AlarmID: alarmID,
	}

	err = deleteUC.DeleteAlarm(ctx, user2ID, delReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission")

	t.Logf("Expected permission error: %v", err)
}
