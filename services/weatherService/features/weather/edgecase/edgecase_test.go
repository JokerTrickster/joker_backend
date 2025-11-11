package edgecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/cache"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/crawler"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/interface"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/repository"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/scheduler"
	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockNotifierEdgeCase implements notifier for edge case testing
type MockNotifierEdgeCase struct {
	shouldFail     bool
	errorMessage   string
	sendCount      int
	partialFailure bool
}

func (m *MockNotifierEdgeCase) SendWeatherNotification(ctx context.Context, tokens []string, data *entity.WeatherData, region string) error {
	if m.shouldFail {
		return errors.New(m.errorMessage)
	}

	m.sendCount++

	if m.partialFailure {
		// Simulate partial failure (some tokens fail)
		return errors.New("some tokens failed")
	}

	return nil
}

// setupEdgeCaseDB creates test database
func setupEdgeCaseDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&entity.UserAlarm{}, &entity.WeatherServiceToken{})
	require.NoError(t, err)

	return db
}

// TestEdgeCase_InvalidWeatherData tests handling of invalid weather data
func TestEdgeCase_InvalidWeatherData(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	weatherCache, err := cache.NewWeatherCache(mr.Addr(), "", logger)
	require.NoError(t, err)
	defer weatherCache.Close()

	// Cache invalid weather data (nil values)
	invalidData := &entity.WeatherData{
		Temperature:   0.0, // Could be valid, but check nil handling
		Humidity:      0.0,
		Precipitation: 0.0,
		WindSpeed:     0.0,
		CachedAt:      time.Now(),
	}

	ctx := context.Background()
	err = weatherCache.Set(ctx, "서울시 강남구", invalidData)
	require.NoError(t, err)

	// Verify retrieval
	retrieved, err := weatherCache.Get(ctx, "서울시 강남구")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, 0.0, retrieved.Temperature)
}

// TestEdgeCase_RedisConnectionLost tests handling of Redis connection loss
func TestEdgeCase_RedisConnectionLost(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mr, err := miniredis.Run()
	require.NoError(t, err)

	weatherCache, err := cache.NewWeatherCache(mr.Addr(), "", logger)
	require.NoError(t, err)

	// Close Redis mid-operation
	mr.Close()

	// Try to get from cache (should fail gracefully)
	ctx := context.Background()
	data, err := weatherCache.Get(ctx, "서울시 강남구")

	// Should handle error gracefully
	assert.Error(t, err)
	assert.Nil(t, data)

	// Try to set (should also fail gracefully)
	weatherData := &entity.WeatherData{
		Temperature: 25.5,
		CachedAt:    time.Now(),
	}
	err = weatherCache.Set(ctx, "서울시 강남구", weatherData)
	assert.Error(t, err)
}

// TestEdgeCase_MalformedRegionNames tests handling of unusual region names
func TestEdgeCase_MalformedRegionNames(t *testing.T) {
	testCases := []struct {
		name         string
		region       string
		shouldHandle bool
	}{
		{
			name:         "Empty string",
			region:       "",
			shouldHandle: true,
		},
		{
			name:         "Very long region name",
			region:       "서울특별시 강남구 역삼동 테헤란로 152",
			shouldHandle: true,
		},
		{
			name:         "Special characters",
			region:       "서울시 !@#$%",
			shouldHandle: true,
		},
		{
			name:         "Unicode emoji",
			region:       "서울시😀",
			shouldHandle: true,
		},
		{
			name:         "SQL injection attempt",
			region:       "서울시'; DROP TABLE user_alarms;",
			shouldHandle: true,
		},
	}

	db := setupEdgeCaseDB(t)
	logger, _ := zap.NewDevelopment()

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	weatherCache, err := cache.NewWeatherCache(mr.Addr(), "", logger)
	require.NoError(t, err)
	defer weatherCache.Close()

	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test cache operations with malformed region
			weatherData := &entity.WeatherData{
				Temperature: 25.5,
				Humidity:    60.0,
				CachedAt:    time.Now(),
			}

			// Set should not panic
			_ = weatherCache.Set(ctx, tc.region, weatherData)
			if tc.shouldHandle {
				// Should handle gracefully (may succeed or fail, but not panic)
				assert.NotPanics(t, func() {
					weatherCache.Get(ctx, tc.region)
				})
			}

			// Test database operations
			alarm := entity.UserAlarm{
				UserID:    1,
				AlarmTime: "09:00:00",
				Region:    tc.region,
				IsEnabled: true,
			}

			_ = db.Create(&alarm).Error
			if tc.shouldHandle {
				// Should handle gracefully
				assert.NotPanics(t, func() {
					db.Where("region = ?", tc.region).First(&entity.UserAlarm{})
				})
			}
		})
	}
}

// TestEdgeCase_ConcurrentAccessSameAlarm tests concurrent processing of same alarm
func TestEdgeCase_ConcurrentAccessSameAlarm(t *testing.T) {
	db := setupEdgeCaseDB(t)

	// Create single alarm
	alarm := entity.UserAlarm{
		UserID:    1,
		AlarmTime: "09:00:00",
		Region:    "서울시 강남구",
		IsEnabled: true,
		LastSent:  nil,
	}
	err := db.Create(&alarm).Error
	require.NoError(t, err)

	repo := repository.NewSchedulerWeatherRepository(db)
	ctx := context.Background()

	// Update last_sent concurrently multiple times
	const concurrency = 10
	errChan := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			err := repo.UpdateLastSent(ctx, alarm.ID, time.Now())
			errChan <- err
		}(i)
	}

	// Collect results
	var errors []error
	for i := 0; i < concurrency; i++ {
		err := <-errChan
		if err != nil {
			errors = append(errors, err)
		}
	}

	// Should handle concurrent updates gracefully
	assert.Empty(t, errors, "Concurrent updates should not cause errors")

	// Verify final state
	var updatedAlarm entity.UserAlarm
	err = db.First(&updatedAlarm, alarm.ID).Error
	require.NoError(t, err)
	assert.NotNil(t, updatedAlarm.LastSent, "LastSent should be updated")
}

// TestEdgeCase_TimeZoneHandling tests time zone edge cases
func TestEdgeCase_TimeZoneHandling(t *testing.T) {
	testCases := []struct {
		name       string
		alarmTime  string
		targetTime time.Time
		shouldFire bool
	}{
		{
			name:       "Midnight boundary",
			alarmTime:  "00:00:00",
			targetTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
			shouldFire: true,
		},
		{
			name:       "End of day",
			alarmTime:  "23:59:59",
			targetTime: time.Date(2024, 1, 1, 23, 59, 59, 0, time.Local),
			shouldFire: true,
		},
		{
			name:       "Morning time",
			alarmTime:  "09:00:00",
			targetTime: time.Date(2024, 1, 1, 9, 0, 0, 0, time.Local),
			shouldFire: true,
		},
	}

	db := setupEdgeCaseDB(t)
	repo := repository.NewSchedulerWeatherRepository(db)
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create alarm
			alarm := entity.UserAlarm{
				UserID:    1,
				AlarmTime: tc.alarmTime,
				Region:    "서울시 강남구",
				IsEnabled: true,
			}
			err := db.Create(&alarm).Error
			require.NoError(t, err)

			// Query alarms to notify
			alarms, err := repo.GetAlarmsToNotify(ctx, tc.targetTime)
			assert.NoError(t, err)

			if tc.shouldFire {
				assert.NotEmpty(t, alarms, "Alarm should fire at boundary time")
			}

			// Cleanup
			db.Delete(&alarm)
		})
	}
}

// TestEdgeCase_NotifierServiceUnavailable tests notifier service unavailability
func TestEdgeCase_NotifierServiceUnavailable(t *testing.T) {
	db := setupEdgeCaseDB(t)
	logger, _ := zap.NewDevelopment()

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	weatherCache, err := cache.NewWeatherCache(mr.Addr(), "", logger)
	require.NoError(t, err)
	defer weatherCache.Close()

	// Mock notifier with failure
	mockNotifier := &MockNotifierEdgeCase{
		shouldFail:   true,
		errorMessage: "service unavailable",
	}

	repo := repository.NewSchedulerWeatherRepository(db)
	weatherCrawler := crawler.NewNaverWeatherCrawler(10*time.Second, 3)

	// Pre-cache weather data
	weatherData := &entity.WeatherData{
		Temperature: 25.5,
		Humidity:    60.0,
		CachedAt:    time.Now(),
	}
	ctx := context.Background()
	err = weatherCache.Set(ctx, "서울시 강남구", weatherData)
	require.NoError(t, err)

	// Create alarm and tokens
	targetTime := time.Now().Add(5 * time.Second)
	alarm := entity.UserAlarm{
		UserID:    1,
		AlarmTime: targetTime.Format("15:04:05"),
		Region:    "서울시 강남구",
		IsEnabled: true,
	}
	err = db.Create(&alarm).Error
	require.NoError(t, err)

	token := entity.WeatherServiceToken{
		UserID:   1,
		FCMToken: "test_token",
		DeviceID: "test_device",
	}
	err = db.Create(&token).Error
	require.NoError(t, err)

	// Create scheduler
	schedulerService := scheduler.NewWeatherSchedulerService(
		repo,
		weatherCrawler,
		weatherCache,
		mockNotifier,
		logger,
		1*time.Second,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go schedulerService.Start(ctx)
	defer schedulerService.Stop()

	// Wait for processing
	waitDuration := time.Until(targetTime) + 10*time.Second
	if waitDuration < 15*time.Second {
		waitDuration = 15 * time.Second
	}
	time.Sleep(waitDuration)

	// Verify last_sent was still updated despite notifier failure
	var updatedAlarm entity.UserAlarm
	err = db.First(&updatedAlarm, alarm.ID).Error
	require.NoError(t, err)

	assert.NotNil(t, updatedAlarm.LastSent,
		"LastSent should be updated even when notifier fails to prevent retry storm")
	assert.Equal(t, 0, mockNotifier.sendCount,
		"No successful sends when notifier is unavailable")
}

// TestEdgeCase_PartialNotifierFailure tests partial notifier token failures
func TestEdgeCase_PartialNotifierFailure(t *testing.T) {
	db := setupEdgeCaseDB(t)
	logger, _ := zap.NewDevelopment()

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	weatherCache, err := cache.NewWeatherCache(mr.Addr(), "", logger)
	require.NoError(t, err)
	defer weatherCache.Close()

	// Mock notifier with partial failure
	mockNotifier := &MockNotifierEdgeCase{
		partialFailure: true,
	}

	repo := repository.NewSchedulerWeatherRepository(db)
	weatherCrawler := crawler.NewNaverWeatherCrawler(10*time.Second, 3)

	// Pre-cache weather data
	weatherData := &entity.WeatherData{
		Temperature: 25.5,
		CachedAt:    time.Now(),
	}
	ctx := context.Background()
	err = weatherCache.Set(ctx, "서울시 강남구", weatherData)
	require.NoError(t, err)

	// Create alarm with multiple tokens
	targetTime := time.Now().Add(5 * time.Second)
	alarm := entity.UserAlarm{
		UserID:    1,
		AlarmTime: targetTime.Format("15:04:05"),
		Region:    "서울시 강남구",
		IsEnabled: true,
	}
	err = db.Create(&alarm).Error
	require.NoError(t, err)

	// Create 4 tokens
	for i := 0; i < 4; i++ {
		token := entity.WeatherServiceToken{
			UserID:   1,
			FCMToken: "test_token_" + string(rune('a'+i)),
			DeviceID: "test_device_" + string(rune('a'+i)),
		}
		err = db.Create(&token).Error
		require.NoError(t, err)
	}

	// Create scheduler
	schedulerService := scheduler.NewWeatherSchedulerService(
		repo,
		weatherCrawler,
		weatherCache,
		mockNotifier,
		logger,
		1*time.Second,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go schedulerService.Start(ctx)
	defer schedulerService.Stop()

	// Wait for processing
	waitDuration := time.Until(targetTime) + 10*time.Second
	if waitDuration < 15*time.Second {
		waitDuration = 15 * time.Second
	}
	time.Sleep(waitDuration)

	// Verify alarm was processed despite partial failure
	var updatedAlarm entity.UserAlarm
	err = db.First(&updatedAlarm, alarm.ID).Error
	require.NoError(t, err)

	assert.NotNil(t, updatedAlarm.LastSent,
		"LastSent should be updated even with partial notifier failures")
	assert.Equal(t, 1, mockNotifier.sendCount,
		"Should attempt to send despite partial failures")
}

// TestEdgeCase_EmptyTokensList tests alarm with empty tokens list
func TestEdgeCase_EmptyTokensList(t *testing.T) {
	db := setupEdgeCaseDB(t)
	logger, _ := zap.NewDevelopment()

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	weatherCache, err := cache.NewWeatherCache(mr.Addr(), "", logger)
	require.NoError(t, err)
	defer weatherCache.Close()

	mockNotifier := &MockNotifierEdgeCase{}
	repo := repository.NewSchedulerWeatherRepository(db)
	weatherCrawler := crawler.NewNaverWeatherCrawler(10*time.Second, 3)

	// Pre-cache weather data
	weatherData := &entity.WeatherData{
		Temperature: 25.5,
		CachedAt:    time.Now(),
	}
	ctx := context.Background()
	err = weatherCache.Set(ctx, "서울시 강남구", weatherData)
	require.NoError(t, err)

	// Create alarm WITHOUT tokens
	targetTime := time.Now().Add(5 * time.Second)
	alarm := entity.UserAlarm{
		UserID:    999, // User with no tokens
		AlarmTime: targetTime.Format("15:04:05"),
		Region:    "서울시 강남구",
		IsEnabled: true,
	}
	err = db.Create(&alarm).Error
	require.NoError(t, err)

	// Create scheduler
	schedulerService := scheduler.NewWeatherSchedulerService(
		repo,
		weatherCrawler,
		weatherCache,
		mockNotifier,
		logger,
		1*time.Second,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go schedulerService.Start(ctx)
	defer schedulerService.Stop()

	// Wait for processing
	waitDuration := time.Until(targetTime) + 10*time.Second
	if waitDuration < 15*time.Second {
		waitDuration = 15 * time.Second
	}
	time.Sleep(waitDuration)

	// Verify last_sent was updated despite no tokens
	var updatedAlarm entity.UserAlarm
	err = db.First(&updatedAlarm, alarm.ID).Error
	require.NoError(t, err)

	assert.NotNil(t, updatedAlarm.LastSent,
		"LastSent should be updated even when user has no tokens")
	assert.Equal(t, 0, mockNotifier.sendCount,
		"Should not send when no tokens available")
}

// Ensure interface compliance
var _ _interface.IFCMNotifier = (*MockNotifierEdgeCase)(nil)
