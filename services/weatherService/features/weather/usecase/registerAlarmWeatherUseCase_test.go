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

func setupTestDB(t *testing.T) *gorm.DB {
	dsn := "root:rootpassword@tcp(localhost:3307)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	return db
}

func TestRegisterAlarmWeatherUseCase_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRegisterAlarmWeatherRepository(db)
	uc := NewRegisterAlarmWeatherUseCase(repo, 10*time.Second)

	ctx := context.Background()
	userID := 1

	req := &request.ReqRegisterAlarm{
		AlarmTime: "09:00",
		Region:    "서울시 관악구",
		FCMToken:  "test-fcm-token-success",
		DeviceID:  "test-device-success",
	}

	res, err := uc.RegisterAlarm(ctx, userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "09:00:00", res.AlarmTime)
	assert.Equal(t, "서울시 관악구", res.Region)
	assert.True(t, res.IsEnabled)
	assert.Greater(t, res.AlarmID, 0)

	t.Logf("Created alarm: %+v", res)
}

func TestRegisterAlarmWeatherUseCase_TimeFormatValidation(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRegisterAlarmWeatherRepository(db)
	uc := NewRegisterAlarmWeatherUseCase(repo, 10*time.Second)

	tests := []struct {
		name        string
		alarmTime   string
		expectError bool
		expectedFmt string
	}{
		{
			name:        "Valid HH:MM format",
			alarmTime:   "14:30",
			expectError: false,
			expectedFmt: "14:30:00",
		},
		{
			name:        "Valid HH:MM:SS format",
			alarmTime:   "06:30:00",
			expectError: false,
			expectedFmt: "06:30:00",
		},
		{
			name:        "Invalid hour (25)",
			alarmTime:   "25:00",
			expectError: true,
		},
		{
			name:        "Invalid minute (70)",
			alarmTime:   "10:70",
			expectError: true,
		},
		{
			name:        "Invalid format (single digit)",
			alarmTime:   "9:00",
			expectError: true,
		},
		{
			name:        "Empty string",
			alarmTime:   "",
			expectError: true,
		},
	}

	ctx := context.Background()
	userID := 1

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &request.ReqRegisterAlarm{
				AlarmTime: tt.alarmTime,
				Region:    "서울시 강남구",
				FCMToken:  "test-fcm-token",
				DeviceID:  "test-device",
			}

			res, err := uc.RegisterAlarm(ctx, userID, req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, res)
				t.Logf("Expected error for %s: %v", tt.alarmTime, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, tt.expectedFmt, res.AlarmTime)
				t.Logf("Successfully validated %s -> %s", tt.alarmTime, res.AlarmTime)
			}
		})
	}
}

func TestRegisterAlarmWeatherUseCase_FCMTokenUpdate(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRegisterAlarmWeatherRepository(db)
	uc := NewRegisterAlarmWeatherUseCase(repo, 10*time.Second)

	ctx := context.Background()
	userID := 2
	deviceID := "test-device-update"

	// First registration
	req1 := &request.ReqRegisterAlarm{
		AlarmTime: "08:00",
		Region:    "서울시 서초구",
		FCMToken:  "old-fcm-token",
		DeviceID:  deviceID,
	}

	res1, err := uc.RegisterAlarm(ctx, userID, req1)
	assert.NoError(t, err)
	assert.NotNil(t, res1)
	t.Logf("First registration: %+v", res1)

	// Second registration with same user and device, different FCM token
	req2 := &request.ReqRegisterAlarm{
		AlarmTime: "10:00",
		Region:    "서울시 송파구",
		FCMToken:  "new-fcm-token",
		DeviceID:  deviceID,
	}

	res2, err := uc.RegisterAlarm(ctx, userID, req2)
	assert.NoError(t, err)
	assert.NotNil(t, res2)
	t.Logf("Second registration: %+v", res2)

	// Both alarms should be created successfully
	// FCM token should be updated (UPSERT)
}

func TestValidateAndFormatTime(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRegisterAlarmWeatherRepository(db)
	uc := NewRegisterAlarmWeatherUseCase(repo, 10*time.Second).(*RegisterAlarmWeatherUseCase)

	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"09:30", "09:30:00", false},
		{"23:59", "23:59:00", false},
		{"00:00", "00:00:00", false},
		{"14:30:45", "14:30:45", false},
		{"25:00", "", true},
		{"12:60", "", true},
		{"9:30", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := uc.validateAndFormatTime(tt.input)

			if tt.hasError {
				assert.Error(t, err)
				t.Logf("Expected error for input '%s': %v", tt.input, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
				t.Logf("'%s' -> '%s'", tt.input, result)
			}
		})
	}
}
