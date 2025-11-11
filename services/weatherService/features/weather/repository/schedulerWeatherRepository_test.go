package repository

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
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

func createTestUsers(t *testing.T, db *gorm.DB) {
	// Create test users for foreign key constraints
	testUsers := []struct {
		ID       int
		Name     string
		Email    string
		Password string
	}{
		{999, "Test User 999", "test999@example.com", "password999"},
		{1000, "Test User 1000", "test1000@example.com", "password1000"},
		{1001, "Test User 1001", "test1001@example.com", "password1001"},
	}

	for _, user := range testUsers {
		// Use INSERT IGNORE to prevent duplicate key errors
		db.Exec("INSERT IGNORE INTO users (id, name, email, password) VALUES (?, ?, ?, ?)",
			user.ID, user.Name, user.Email, user.Password)
	}
}

func cleanupTestData(t *testing.T, db *gorm.DB) {
	// Clean up test data
	db.Exec("DELETE FROM user_alarms WHERE user_id IN (999, 1000, 1001)")
	db.Exec("DELETE FROM weather_service_tokens WHERE user_id IN (999, 1000, 1001)")
}

func TestGetAlarmsToNotify_Success(t *testing.T) {
	db := setupTestDB(t)
	createTestUsers(t, db)
	defer cleanupTestData(t, db)

	repo := NewSchedulerWeatherRepository(db)
	ctx := context.Background()

	// Create test alarm for 09:00:00
	testAlarm := &entity.UserAlarm{
		UserID:    999,
		AlarmTime: "09:00:00",
		Region:    "서울시 강남구",
		IsEnabled: true,
		LastSent:  nil, // Never sent before
	}

	err := db.Create(testAlarm).Error
	assert.NoError(t, err)
	t.Logf("Created test alarm: ID=%d, AlarmTime=%s", testAlarm.ID, testAlarm.AlarmTime)

	// Query alarms for 09:00:00
	targetTime, _ := time.Parse("15:04:05", "09:00:00")
	alarms, err := repo.GetAlarmsToNotify(ctx, targetTime)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(alarms), 1)

	// Find our test alarm
	found := false
	for _, alarm := range alarms {
		if alarm.ID == testAlarm.ID {
			found = true
			assert.Equal(t, 999, alarm.UserID)
			assert.Equal(t, "09:00:00", alarm.AlarmTime)
			assert.Equal(t, "서울시 강남구", alarm.Region)
			assert.True(t, alarm.IsEnabled)
			assert.Nil(t, alarm.LastSent)
			t.Logf("Found test alarm: %+v", alarm)
		}
	}
	assert.True(t, found, "Test alarm should be found in results")
}

func TestGetAlarmsToNotify_OnlyEnabledAlarms(t *testing.T) {
	db := setupTestDB(t)
	createTestUsers(t, db)
	defer cleanupTestData(t, db)

	repo := NewSchedulerWeatherRepository(db)
	ctx := context.Background()

	// Create enabled alarm
	enabledAlarm := &entity.UserAlarm{
		UserID:    1000,
		AlarmTime: "10:00:00",
		Region:    "서울시 서초구",
		IsEnabled: true,
		LastSent:  nil,
	}
	err := db.Create(enabledAlarm).Error
	assert.NoError(t, err)
	t.Logf("Created enabled alarm: ID=%d", enabledAlarm.ID)

	// Create disabled alarm using raw SQL to ensure is_enabled=false
	result := db.Exec(`INSERT INTO user_alarms (user_id, alarm_time, region, is_enabled, last_sent, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())`, 1000, "10:00:00", "서울시 관악구", false, nil)
	assert.NoError(t, result.Error)

	// Get the created alarm ID
	var disabledAlarm entity.UserAlarm
	db.Where("user_id = ? AND alarm_time = ? AND region = ?", 1000, "10:00:00", "서울시 관악구").
		Order("id DESC").First(&disabledAlarm)
	t.Logf("Created disabled alarm: ID=%d, IsEnabled=%v", disabledAlarm.ID, disabledAlarm.IsEnabled)

	// Query alarms for 10:00:00
	targetTime, _ := time.Parse("15:04:05", "10:00:00")
	alarms, err := repo.GetAlarmsToNotify(ctx, targetTime)

	assert.NoError(t, err)

	// Check that only enabled alarm is returned
	enabledFound := false
	disabledFound := false

	for _, alarm := range alarms {
		if alarm.ID == enabledAlarm.ID {
			enabledFound = true
			assert.True(t, alarm.IsEnabled)
			t.Logf("Enabled alarm found: %+v", alarm)
		}
		if alarm.ID == disabledAlarm.ID {
			disabledFound = true
			t.Logf("Disabled alarm found (should not happen): %+v", alarm)
		}
	}

	assert.True(t, enabledFound, "Enabled alarm should be found")
	assert.False(t, disabledFound, "Disabled alarm should NOT be found")
}

func TestGetAlarmsToNotify_DuplicatePrevention(t *testing.T) {
	db := setupTestDB(t)
	createTestUsers(t, db)
	defer cleanupTestData(t, db)

	repo := NewSchedulerWeatherRepository(db)
	ctx := context.Background()

	// Create alarm that was sent today
	today := time.Now()
	sentTodayAlarm := &entity.UserAlarm{
		UserID:    1001,
		AlarmTime: "11:00:00",
		Region:    "서울시 송파구",
		IsEnabled: true,
		LastSent:  &today,
	}
	err := db.Create(sentTodayAlarm).Error
	assert.NoError(t, err)
	t.Logf("Created alarm sent today: ID=%d, LastSent=%v", sentTodayAlarm.ID, sentTodayAlarm.LastSent)

	// Create alarm that was sent yesterday
	yesterday := time.Now().AddDate(0, 0, -1)
	sentYesterdayAlarm := &entity.UserAlarm{
		UserID:    1001,
		AlarmTime: "11:00:00",
		Region:    "서울시 강동구",
		IsEnabled: true,
		LastSent:  &yesterday,
	}
	err = db.Create(sentYesterdayAlarm).Error
	assert.NoError(t, err)
	t.Logf("Created alarm sent yesterday: ID=%d, LastSent=%v", sentYesterdayAlarm.ID, sentYesterdayAlarm.LastSent)

	// Create alarm never sent
	neverSentAlarm := &entity.UserAlarm{
		UserID:    1001,
		AlarmTime: "11:00:00",
		Region:    "서울시 강서구",
		IsEnabled: true,
		LastSent:  nil,
	}
	err = db.Create(neverSentAlarm).Error
	assert.NoError(t, err)
	t.Logf("Created alarm never sent: ID=%d", neverSentAlarm.ID)

	// Query alarms for 11:00:00
	targetTime, _ := time.Parse("15:04:05", "11:00:00")
	alarms, err := repo.GetAlarmsToNotify(ctx, targetTime)

	assert.NoError(t, err)

	// Check results
	sentTodayFound := false
	sentYesterdayFound := false
	neverSentFound := false

	for _, alarm := range alarms {
		switch alarm.ID {
		case sentTodayAlarm.ID:
			sentTodayFound = true
			t.Logf("Found alarm sent today (should be filtered): %+v", alarm)
		case sentYesterdayAlarm.ID:
			sentYesterdayFound = true
			t.Logf("Found alarm sent yesterday (should be included): %+v", alarm)
		case neverSentAlarm.ID:
			neverSentFound = true
			t.Logf("Found alarm never sent (should be included): %+v", alarm)
		}
	}

	assert.False(t, sentTodayFound, "Alarm sent today should be filtered out")
	assert.True(t, sentYesterdayFound, "Alarm sent yesterday should be included")
	assert.True(t, neverSentFound, "Alarm never sent should be included")
}

func TestGetAlarmsToNotify_NoMatchingTime(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSchedulerWeatherRepository(db)
	ctx := context.Background()

	// Query for a time that doesn't exist (23:59:59)
	targetTime, _ := time.Parse("15:04:05", "23:59:59")
	alarms, err := repo.GetAlarmsToNotify(ctx, targetTime)

	assert.NoError(t, err)
	assert.NotNil(t, alarms)
	t.Logf("Alarms found for non-existent time: %d", len(alarms))
}

func TestUpdateLastSent_Success(t *testing.T) {
	db := setupTestDB(t)
	createTestUsers(t, db)
	defer cleanupTestData(t, db)

	repo := NewSchedulerWeatherRepository(db)
	ctx := context.Background()

	// Create test alarm
	testAlarm := &entity.UserAlarm{
		UserID:    999,
		AlarmTime: "08:00:00",
		Region:    "서울시 종로구",
		IsEnabled: true,
		LastSent:  nil,
	}
	err := db.Create(testAlarm).Error
	assert.NoError(t, err)
	t.Logf("Created test alarm: ID=%d, LastSent=%v", testAlarm.ID, testAlarm.LastSent)

	// Update last_sent
	sentTime := time.Now()
	err = repo.UpdateLastSent(ctx, testAlarm.ID, sentTime)
	assert.NoError(t, err)

	// Verify update
	var updatedAlarm entity.UserAlarm
	err = db.First(&updatedAlarm, testAlarm.ID).Error
	assert.NoError(t, err)
	assert.NotNil(t, updatedAlarm.LastSent)
	t.Logf("Updated alarm: ID=%d, LastSent=%v", updatedAlarm.ID, updatedAlarm.LastSent)

	// Check that the timestamp is approximately correct (within 1 second)
	timeDiff := updatedAlarm.LastSent.Sub(sentTime).Abs()
	assert.Less(t, timeDiff, time.Second, "LastSent timestamp should match sent time")
}

func TestUpdateLastSent_AlarmNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSchedulerWeatherRepository(db)
	ctx := context.Background()

	// Try to update non-existent alarm
	nonExistentID := 999999
	sentTime := time.Now()
	err := repo.UpdateLastSent(ctx, nonExistentID, sentTime)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	t.Logf("Expected error for non-existent alarm: %v", err)
}

func TestGetFCMTokens_Success(t *testing.T) {
	db := setupTestDB(t)
	createTestUsers(t, db)
	defer cleanupTestData(t, db)

	repo := NewSchedulerWeatherRepository(db)
	ctx := context.Background()

	// Create test tokens
	token1 := &entity.WeatherServiceToken{
		UserID:   999,
		FCMToken: "test-token-1",
		DeviceID: "device-1",
	}
	err := db.Create(token1).Error
	assert.NoError(t, err)
	t.Logf("Created token 1: ID=%d, FCMToken=%s", token1.ID, token1.FCMToken)

	token2 := &entity.WeatherServiceToken{
		UserID:   999,
		FCMToken: "test-token-2",
		DeviceID: "device-2",
	}
	err = db.Create(token2).Error
	assert.NoError(t, err)
	t.Logf("Created token 2: ID=%d, FCMToken=%s", token2.ID, token2.FCMToken)

	// Query tokens
	tokens, err := repo.GetFCMTokens(ctx, 999)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(tokens), 2)

	// Verify tokens are in results
	token1Found := false
	token2Found := false

	for _, token := range tokens {
		if token.ID == token1.ID {
			token1Found = true
			assert.Equal(t, "test-token-1", token.FCMToken)
			assert.Equal(t, "device-1", token.DeviceID)
		}
		if token.ID == token2.ID {
			token2Found = true
			assert.Equal(t, "test-token-2", token.FCMToken)
			assert.Equal(t, "device-2", token.DeviceID)
		}
	}

	assert.True(t, token1Found, "Token 1 should be found")
	assert.True(t, token2Found, "Token 2 should be found")
	t.Logf("Found %d tokens for user 999", len(tokens))
}

func TestGetFCMTokens_NoTokens(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSchedulerWeatherRepository(db)
	ctx := context.Background()

	// Query for user with no tokens
	nonExistentUserID := 888888
	tokens, err := repo.GetFCMTokens(ctx, nonExistentUserID)

	assert.NoError(t, err)
	assert.NotNil(t, tokens)
	assert.Equal(t, 0, len(tokens))
	t.Logf("No tokens found for user %d (expected)", nonExistentUserID)
}

func TestGetFCMTokens_OnlyActiveTokens(t *testing.T) {
	db := setupTestDB(t)
	createTestUsers(t, db)
	defer cleanupTestData(t, db)

	repo := NewSchedulerWeatherRepository(db)
	ctx := context.Background()

	// Create active token
	activeToken := &entity.WeatherServiceToken{
		UserID:   1000,
		FCMToken: "active-token",
		DeviceID: "active-device",
	}
	err := db.Create(activeToken).Error
	assert.NoError(t, err)
	t.Logf("Created active token: ID=%d", activeToken.ID)

	// Create deleted token
	deletedTime := time.Now()
	deletedToken := &entity.WeatherServiceToken{
		UserID:    1000,
		FCMToken:  "deleted-token",
		DeviceID:  "deleted-device",
		DeletedAt: &deletedTime,
	}
	err = db.Create(deletedToken).Error
	assert.NoError(t, err)
	t.Logf("Created deleted token: ID=%d", deletedToken.ID)

	// Query tokens
	tokens, err := repo.GetFCMTokens(ctx, 1000)

	assert.NoError(t, err)

	// Check that only active token is returned
	activeFound := false
	deletedFound := false

	for _, token := range tokens {
		if token.ID == activeToken.ID {
			activeFound = true
			assert.Nil(t, token.DeletedAt)
			t.Logf("Active token found: %+v", token)
		}
		if token.ID == deletedToken.ID {
			deletedFound = true
		}
	}

	assert.True(t, activeFound, "Active token should be found")
	assert.False(t, deletedFound, "Deleted token should NOT be found")
}
