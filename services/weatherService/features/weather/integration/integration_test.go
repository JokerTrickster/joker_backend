package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"firebase.google.com/go/v4/messaging"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/cache"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/crawler"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/interface"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/notifier"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/repository"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/scheduler"
	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MockFCMClient for testing
type MockFCMClient struct {
	mu         sync.Mutex
	sendCount  int
	lastTokens []string
	lastRegion string
	shouldFail bool
}

func (m *MockFCMClient) SendMulticast(ctx context.Context, message *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFail {
		return nil, fmt.Errorf("mock FCM send failed")
	}

	m.sendCount++

	// Store for verification
	if message != nil {
		m.lastTokens = message.Tokens
	}

	// Create success responses
	successCount := len(message.Tokens)
	responses := make([]*messaging.SendResponse, successCount)
	for i := range responses {
		responses[i] = &messaging.SendResponse{
			Success: true,
		}
	}

	return &messaging.BatchResponse{
		SuccessCount: successCount,
		FailureCount: 0,
		Responses:    responses,
	}, nil
}

func (m *MockFCMClient) SendCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sendCount
}

func (m *MockFCMClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sendCount = 0
	m.lastTokens = nil
	m.lastRegion = ""
}

// TestDatabase manages test database lifecycle
type TestDatabase struct {
	DB     *gorm.DB
	DSN    string
	logger *zap.Logger
}

// setupTestDB initializes test database
func setupTestDB(t *testing.T) *TestDatabase {
	logger, _ := zap.NewDevelopment()

	// Use test database
	dsn := "test_user:test_password@tcp(localhost:3307)/joker_test?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	require.NoError(t, err, "Failed to connect to test database")

	// Auto-migrate tables
	err = db.AutoMigrate(&entity.UserAlarm{}, &entity.WeatherServiceToken{})
	require.NoError(t, err, "Failed to migrate tables")

	testDB := &TestDatabase{
		DB:     db,
		DSN:    dsn,
		logger: logger,
	}

	// Clean database
	testDB.Clean(t)

	return testDB
}

// Clean removes all test data
func (td *TestDatabase) Clean(t *testing.T) {
	td.DB.Exec("DELETE FROM user_alarms")
	td.DB.Exec("DELETE FROM weather_service_tokens")
	td.DB.Exec("ALTER TABLE user_alarms AUTO_INCREMENT = 1")
	td.DB.Exec("ALTER TABLE weather_service_tokens AUTO_INCREMENT = 1")
}

// Close closes database connection
func (td *TestDatabase) Close() {
	sqlDB, _ := td.DB.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
}

// setupTestRedis initializes test Redis using miniredis
func setupTestRedis(t *testing.T) (*cache.WeatherCache, *miniredis.Miniredis) {
	logger, _ := zap.NewDevelopment()

	// Start miniredis server
	mr, err := miniredis.Run()
	require.NoError(t, err, "Failed to start miniredis")

	// Create cache client
	weatherCache, err := cache.NewWeatherCache(mr.Addr(), "", logger)
	require.NoError(t, err, "Failed to create weather cache")

	return weatherCache, mr
}

// createTestAlarm creates a test alarm
func createTestAlarm(t *testing.T, db *gorm.DB, alarmTime time.Time, region string, userID int) *entity.UserAlarm {
	alarm := &entity.UserAlarm{
		UserID:    userID,
		AlarmTime: alarmTime.Format("15:04:05"),
		Region:    region,
		IsEnabled: true,
		LastSent:  nil,
	}

	err := db.Create(alarm).Error
	require.NoError(t, err, "Failed to create test alarm")

	return alarm
}

// createTestAlarmWithLastSent creates a test alarm with last_sent timestamp
func createTestAlarmWithLastSent(t *testing.T, db *gorm.DB, alarmTime time.Time, region string, userID int, lastSent time.Time) *entity.UserAlarm {
	alarm := &entity.UserAlarm{
		UserID:    userID,
		AlarmTime: alarmTime.Format("15:04:05"),
		Region:    region,
		IsEnabled: true,
		LastSent:  &lastSent,
	}

	err := db.Create(alarm).Error
	require.NoError(t, err, "Failed to create test alarm")

	return alarm
}

// createTestFCMTokens creates FCM tokens for a user
func createTestFCMTokens(t *testing.T, db *gorm.DB, userID int, count int) []entity.WeatherServiceToken {
	tokens := make([]entity.WeatherServiceToken, count)

	for i := 0; i < count; i++ {
		token := entity.WeatherServiceToken{
			UserID:   userID,
			FCMToken: fmt.Sprintf("test_token_%d_%d", userID, i),
			DeviceID: fmt.Sprintf("device_%d_%d", userID, i),
		}

		err := db.Create(&token).Error
		require.NoError(t, err, "Failed to create test FCM token")

		tokens[i] = token
	}

	return tokens
}

// getAlarm retrieves an alarm by ID
func getAlarm(t *testing.T, db *gorm.DB, alarmID int) *entity.UserAlarm {
	var alarm entity.UserAlarm
	err := db.First(&alarm, alarmID).Error
	require.NoError(t, err, "Failed to get alarm")
	return &alarm
}

// waitForSchedulerTick waits for scheduler to process alarms
func waitForSchedulerTick(t *testing.T, duration time.Duration) {
	t.Logf("Waiting %v for scheduler tick...", duration)
	time.Sleep(duration)
}

// TestEndToEnd_HappyPath tests the complete flow
func TestEndToEnd_HappyPath(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	testDB := setupTestDB(t)
	defer testDB.Close()
	defer testDB.Clean(t)

	weatherCache, mr := setupTestRedis(t)
	defer mr.Close()
	defer weatherCache.Close()

	mockFCM := &MockFCMClient{}
	logger, _ := zap.NewDevelopment()

	// Create dependencies
	repo := repository.NewSchedulerWeatherRepository(testDB.DB)
	weatherCrawler := crawler.NewNaverWeatherCrawler(10*time.Second, 3)
	fcmNotifier := notifier.NewFCMNotifierWithClient(mockFCM, logger)

	// Create test data - alarm for current_time + 10 seconds
	targetTime := time.Now().Add(10 * time.Second)
	userID := 1
	region := "서울시 강남구"

	alarm := createTestAlarm(t, testDB.DB, targetTime, region, userID)
	tokens := createTestFCMTokens(t, testDB.DB, userID, 2)

	t.Logf("Created alarm ID=%d for time=%s, user=%d, tokens=%d",
		alarm.ID, targetTime.Format("15:04:05"), userID, len(tokens))

	// Create scheduler with short interval for testing
	schedulerService := scheduler.NewWeatherSchedulerService(
		repo,
		weatherCrawler,
		weatherCache,
		fcmNotifier,
		logger,
		5*time.Second, // tick every 5 seconds
	)

	// Start scheduler
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go schedulerService.Start(ctx)
	defer schedulerService.Stop()

	// Wait for processing (target time + buffer)
	waitTime := time.Until(targetTime) + 10*time.Second
	if waitTime < 15*time.Second {
		waitTime = 15 * time.Second
	}
	waitForSchedulerTick(t, waitTime)

	// Verify notification was sent
	assert.Greater(t, mockFCM.SendCount(), 0, "FCM notification should be sent")

	// Verify last_sent updated
	updatedAlarm := getAlarm(t, testDB.DB, alarm.ID)
	assert.NotNil(t, updatedAlarm.LastSent, "LastSent should be updated")
	if updatedAlarm.LastSent != nil {
		t.Logf("LastSent updated to: %v", updatedAlarm.LastSent)
	}

	t.Logf("Test completed: FCM sends=%d", mockFCM.SendCount())
}

// TestEndToEnd_CacheHit tests cache hit scenario
func TestEndToEnd_CacheHit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	testDB := setupTestDB(t)
	defer testDB.Close()
	defer testDB.Clean(t)

	weatherCache, mr := setupTestRedis(t)
	defer mr.Close()
	defer weatherCache.Close()

	mockFCM := &MockFCMClient{}
	logger, _ := zap.NewDevelopment()

	// Pre-populate cache
	region := "서울시 강남구"
	cachedData := &entity.WeatherData{
		Temperature:   25.5,
		Humidity:      60.0,
		Precipitation: 0.0,
		WindSpeed:     3.2,
		CachedAt:      time.Now(),
	}
	err := weatherCache.Set(context.Background(), region, cachedData)
	require.NoError(t, err, "Failed to cache weather data")

	t.Logf("Pre-populated cache with data: temp=%.1f", cachedData.Temperature)

	// Create dependencies
	repo := repository.NewSchedulerWeatherRepository(testDB.DB)
	weatherCrawler := crawler.NewNaverWeatherCrawler(10*time.Second, 3)
	fcmNotifier := notifier.NewFCMNotifierWithClient(mockFCM, logger)

	// Create test data
	targetTime := time.Now().Add(10 * time.Second)
	userID := 1

	alarm := createTestAlarm(t, testDB.DB, targetTime, region, userID)
	createTestFCMTokens(t, testDB.DB, userID, 1)

	// Create scheduler
	schedulerService := scheduler.NewWeatherSchedulerService(
		repo,
		weatherCrawler,
		weatherCache,
		fcmNotifier,
		logger,
		5*time.Second,
	)

	// Start scheduler
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go schedulerService.Start(ctx)
	defer schedulerService.Stop()

	// Wait for processing
	waitTime := time.Until(targetTime) + 10*time.Second
	if waitTime < 15*time.Second {
		waitTime = 15 * time.Second
	}
	waitForSchedulerTick(t, waitTime)

	// Verify notification sent (cache hit path)
	assert.Greater(t, mockFCM.SendCount(), 0, "FCM notification should be sent from cache")

	// Verify last_sent updated
	updatedAlarm := getAlarm(t, testDB.DB, alarm.ID)
	assert.NotNil(t, updatedAlarm.LastSent, "LastSent should be updated")

	t.Logf("Cache hit test completed: FCM sends=%d", mockFCM.SendCount())
}

// TestEndToEnd_DuplicatePrevention tests duplicate prevention
func TestEndToEnd_DuplicatePrevention(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	testDB := setupTestDB(t)
	defer testDB.Close()
	defer testDB.Clean(t)

	weatherCache, mr := setupTestRedis(t)
	defer mr.Close()
	defer weatherCache.Close()

	mockFCM := &MockFCMClient{}
	logger, _ := zap.NewDevelopment()

	// Create dependencies
	repo := repository.NewSchedulerWeatherRepository(testDB.DB)
	weatherCrawler := crawler.NewNaverWeatherCrawler(10*time.Second, 3)
	fcmNotifier := notifier.NewFCMNotifierWithClient(mockFCM, logger)

	// Create alarm with last_sent = today (should NOT process)
	targetTime := time.Now().Add(10 * time.Second)
	userID := 1
	region := "서울시 강남구"
	today := time.Now()

	alarm := createTestAlarmWithLastSent(t, testDB.DB, targetTime, region, userID, today)
	createTestFCMTokens(t, testDB.DB, userID, 1)

	t.Logf("Created alarm with last_sent=today (should NOT process)")

	// Create scheduler
	schedulerService := scheduler.NewWeatherSchedulerService(
		repo,
		weatherCrawler,
		weatherCache,
		fcmNotifier,
		logger,
		5*time.Second,
	)

	// Start scheduler
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	go schedulerService.Start(ctx)
	defer schedulerService.Stop()

	// Wait for tick
	waitTime := time.Until(targetTime) + 10*time.Second
	if waitTime < 15*time.Second {
		waitTime = 15 * time.Second
	}
	waitForSchedulerTick(t, waitTime)

	// Verify NO notification sent (duplicate prevention)
	assert.Equal(t, 0, mockFCM.SendCount(), "No FCM notification should be sent (already sent today)")

	// Now test with last_sent = yesterday (should process)
	mockFCM.Reset()
	yesterday := time.Now().AddDate(0, 0, -1)

	// Update alarm with last_sent = yesterday
	testDB.DB.Model(&entity.UserAlarm{}).
		Where("id = ?", alarm.ID).
		Update("last_sent", yesterday)

	t.Logf("Updated alarm with last_sent=yesterday (should process)")

	// Wait for next tick
	waitForSchedulerTick(t, 15*time.Second)

	// Verify notification sent
	assert.Greater(t, mockFCM.SendCount(), 0, "FCM notification should be sent (last_sent = yesterday)")

	t.Logf("Duplicate prevention test completed: initial sends=%d, after update sends=%d",
		0, mockFCM.SendCount())
}

// TestEndToEnd_MultipleAlarms tests concurrent alarm processing
func TestEndToEnd_MultipleAlarms(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	testDB := setupTestDB(t)
	defer testDB.Close()
	defer testDB.Clean(t)

	weatherCache, mr := setupTestRedis(t)
	defer mr.Close()
	defer weatherCache.Close()

	mockFCM := &MockFCMClient{}
	logger, _ := zap.NewDevelopment()

	// Create dependencies
	repo := repository.NewSchedulerWeatherRepository(testDB.DB)
	weatherCrawler := crawler.NewNaverWeatherCrawler(10*time.Second, 3)
	fcmNotifier := notifier.NewFCMNotifierWithClient(mockFCM, logger)

	// Create multiple alarms at same time
	targetTime := time.Now().Add(10 * time.Second)
	alarmCount := 10
	regions := []string{
		"서울시 강남구",
		"경기도 성남시",
		"부산시 해운대구",
		"대구시 중구",
		"인천시 남동구",
	}

	t.Logf("Creating %d alarms for time=%s", alarmCount, targetTime.Format("15:04:05"))

	for i := 0; i < alarmCount; i++ {
		userID := i + 1
		region := regions[i%len(regions)]

		createTestAlarm(t, testDB.DB, targetTime, region, userID)
		createTestFCMTokens(t, testDB.DB, userID, 1)
	}

	// Create scheduler
	schedulerService := scheduler.NewWeatherSchedulerService(
		repo,
		weatherCrawler,
		weatherCache,
		fcmNotifier,
		logger,
		5*time.Second,
	)

	// Start scheduler
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go schedulerService.Start(ctx)
	defer schedulerService.Stop()

	// Wait for processing
	waitTime := time.Until(targetTime) + 15*time.Second
	if waitTime < 20*time.Second {
		waitTime = 20 * time.Second
	}
	waitForSchedulerTick(t, waitTime)

	// Verify all notifications sent
	assert.GreaterOrEqual(t, mockFCM.SendCount(), alarmCount,
		"All alarms should be processed")

	// Verify all alarms have last_sent updated
	var alarms []entity.UserAlarm
	err := testDB.DB.Where("alarm_time = ?", targetTime.Format("15:04:05")).Find(&alarms).Error
	require.NoError(t, err)

	updatedCount := 0
	for _, alarm := range alarms {
		if alarm.LastSent != nil {
			updatedCount++
		}
	}

	t.Logf("Multiple alarms test completed: alarms=%d, processed=%d, updated=%d",
		alarmCount, mockFCM.SendCount(), updatedCount)

	assert.GreaterOrEqual(t, updatedCount, alarmCount,
		"All alarms should have last_sent updated")
}

// TestEndToEnd_NoTokens tests alarm with no FCM tokens
func TestEndToEnd_NoTokens(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	testDB := setupTestDB(t)
	defer testDB.Close()
	defer testDB.Clean(t)

	weatherCache, mr := setupTestRedis(t)
	defer mr.Close()
	defer weatherCache.Close()

	mockFCM := &MockFCMClient{}
	logger, _ := zap.NewDevelopment()

	// Create dependencies
	repo := repository.NewSchedulerWeatherRepository(testDB.DB)
	weatherCrawler := crawler.NewNaverWeatherCrawler(10*time.Second, 3)
	fcmNotifier := notifier.NewFCMNotifierWithClient(mockFCM, logger)

	// Create alarm WITHOUT FCM tokens
	targetTime := time.Now().Add(10 * time.Second)
	userID := 999
	region := "서울시 강남구"

	alarm := createTestAlarm(t, testDB.DB, targetTime, region, userID)
	// NOTE: No FCM tokens created

	t.Logf("Created alarm ID=%d WITHOUT FCM tokens", alarm.ID)

	// Create scheduler
	schedulerService := scheduler.NewWeatherSchedulerService(
		repo,
		weatherCrawler,
		weatherCache,
		fcmNotifier,
		logger,
		5*time.Second,
	)

	// Start scheduler
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go schedulerService.Start(ctx)
	defer schedulerService.Stop()

	// Wait for processing
	waitTime := time.Until(targetTime) + 10*time.Second
	if waitTime < 15*time.Second {
		waitTime = 15 * time.Second
	}
	waitForSchedulerTick(t, waitTime)

	// Verify NO notification sent (no tokens)
	assert.Equal(t, 0, mockFCM.SendCount(), "No FCM notification should be sent (no tokens)")

	// Verify last_sent still updated (prevent retry)
	updatedAlarm := getAlarm(t, testDB.DB, alarm.ID)
	assert.NotNil(t, updatedAlarm.LastSent, "LastSent should be updated even without tokens")

	t.Logf("No tokens test completed: last_sent updated=%v", updatedAlarm.LastSent != nil)
}

// MockFCMClientInterface ensures MockFCMClient implements the interface
var _ _interface.IFCMClient = (*MockFCMClient)(nil)
