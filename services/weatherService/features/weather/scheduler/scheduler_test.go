package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
	"github.com/JokerTrickster/joker_backend/services/weatherService/pkg/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Initialize metrics once for all tests
func init() {
	metrics.InitMetrics()
}

// Mock implementations

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetAlarmsToNotify(ctx context.Context, targetTime time.Time) ([]entity.UserAlarm, error) {
	args := m.Called(ctx, targetTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.UserAlarm), args.Error(1)
}

func (m *MockRepository) UpdateLastSent(ctx context.Context, alarmID int, sentTime time.Time) error {
	args := m.Called(ctx, alarmID, mock.Anything)
	return args.Error(0)
}

func (m *MockRepository) GetFCMTokens(ctx context.Context, userID int) ([]entity.WeatherServiceToken, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.WeatherServiceToken), args.Error(1)
}

type MockCrawler struct {
	mock.Mock
}

func (m *MockCrawler) Fetch(ctx context.Context, region string) (*entity.WeatherData, error) {
	args := m.Called(ctx, region)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WeatherData), args.Error(1)
}

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, region string) (*entity.WeatherData, error) {
	args := m.Called(ctx, region)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WeatherData), args.Error(1)
}

func (m *MockCache) Set(ctx context.Context, region string, data *entity.WeatherData) error {
	args := m.Called(ctx, region, data)
	return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, region string) error {
	args := m.Called(ctx, region)
	return args.Error(0)
}

func (m *MockCache) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCache) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCache) GetTTL(ctx context.Context, region string) (time.Duration, error) {
	args := m.Called(ctx, region)
	return args.Get(0).(time.Duration), args.Error(1)
}

type MockNotifier struct {
	mock.Mock
	mu              sync.Mutex
	notificationLog []NotificationRecord
}

type NotificationRecord struct {
	Tokens []string
	Data   *entity.WeatherData
	Region string
	Time   time.Time
}

func (m *MockNotifier) SendWeatherNotification(ctx context.Context, tokens []string, data *entity.WeatherData, region string) error {
	m.mu.Lock()
	m.notificationLog = append(m.notificationLog, NotificationRecord{
		Tokens: tokens,
		Data:   data,
		Region: region,
		Time:   time.Now(),
	})
	m.mu.Unlock()

	args := m.Called(ctx, tokens, data, region)
	return args.Error(0)
}

func (m *MockNotifier) GetNotificationLog() []NotificationRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]NotificationRecord{}, m.notificationLog...)
}

// Test helper functions

func createTestLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

func createTestScheduler(repo *MockRepository, crawler *MockCrawler, cache *MockCache, notifier *MockNotifier) *WeatherSchedulerService {
	return NewWeatherSchedulerService(
		repo,
		crawler,
		cache,
		notifier,
		createTestLogger(),
		100*time.Millisecond, // Short interval for testing
	)
}

// Tests

func TestNewWeatherSchedulerService(t *testing.T) {
	repo := new(MockRepository)
	crawler := new(MockCrawler)
	cache := new(MockCache)
	notifier := new(MockNotifier)

	scheduler := NewWeatherSchedulerService(
		repo,
		crawler,
		cache,
		notifier,
		nil,
		1*time.Minute,
	)

	assert.NotNil(t, scheduler)
	assert.NotNil(t, scheduler.logger)
	assert.Equal(t, 1*time.Minute, scheduler.interval)
	assert.False(t, scheduler.running)
}

func TestStartStop_Lifecycle(t *testing.T) {
	repo := new(MockRepository)
	crawler := new(MockCrawler)
	cache := new(MockCache)
	notifier := new(MockNotifier)

	// Mock empty alarm list
	repo.On("GetAlarmsToNotify", mock.Anything, mock.Anything).Return([]entity.UserAlarm{}, nil)

	scheduler := createTestScheduler(repo, crawler, cache, notifier)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start scheduler
	go func() {
		err := scheduler.Start(ctx)
		assert.NoError(t, err)
	}()

	// Wait for scheduler to start
	time.Sleep(50 * time.Millisecond)
	assert.True(t, scheduler.running)

	// Stop scheduler
	err := scheduler.Stop()
	assert.NoError(t, err)
	assert.False(t, scheduler.running)

	// Verify idempotent stop
	err = scheduler.Stop()
	assert.NoError(t, err)
}

func TestStart_AlreadyRunning(t *testing.T) {
	repo := new(MockRepository)
	crawler := new(MockCrawler)
	cache := new(MockCache)
	notifier := new(MockNotifier)

	repo.On("GetAlarmsToNotify", mock.Anything, mock.Anything).Return([]entity.UserAlarm{}, nil)

	scheduler := createTestScheduler(repo, crawler, cache, notifier)

	ctx := context.Background()
	go scheduler.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	// Try to start again
	err := scheduler.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	scheduler.Stop()
}

func TestProcessAlarms_NoAlarms(t *testing.T) {
	repo := new(MockRepository)
	crawler := new(MockCrawler)
	cache := new(MockCache)
	notifier := new(MockNotifier)

	targetTime := time.Now()
	repo.On("GetAlarmsToNotify", mock.Anything, targetTime).Return([]entity.UserAlarm{}, nil)

	scheduler := createTestScheduler(repo, crawler, cache, notifier)
	ctx := context.Background()

	err := scheduler.processAlarms(ctx, targetTime)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	crawler.AssertNotCalled(t, "Fetch")
	notifier.AssertNotCalled(t, "SendWeatherNotification")
}

func TestProcessAlarms_RepositoryError(t *testing.T) {
	repo := new(MockRepository)
	crawler := new(MockCrawler)
	cache := new(MockCache)
	notifier := new(MockNotifier)

	targetTime := time.Now()
	expectedErr := errors.New("database error")
	repo.On("GetAlarmsToNotify", mock.Anything, targetTime).Return(nil, expectedErr)

	scheduler := createTestScheduler(repo, crawler, cache, notifier)
	ctx := context.Background()

	err := scheduler.processAlarms(ctx, targetTime)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get alarms to notify")
	repo.AssertExpectations(t)
}

func TestProcessAlarms_Success_CacheHit(t *testing.T) {
	repo := new(MockRepository)
	crawler := new(MockCrawler)
	cache := new(MockCache)
	notifier := new(MockNotifier)

	targetTime := time.Now()
	alarms := []entity.UserAlarm{
		{
			ID:        1,
			UserID:    100,
			AlarmTime: "09:00:00",
			Region:    "서울시 강남구",
			IsEnabled: true,
		},
	}

	weatherData := &entity.WeatherData{
		Temperature:   25.5,
		Humidity:      60.0,
		Precipitation: 0.0,
		WindSpeed:     2.5,
		CachedAt:      time.Now(),
	}

	tokens := []entity.WeatherServiceToken{
		{ID: 1, UserID: 100, FCMToken: "token1"},
		{ID: 2, UserID: 100, FCMToken: "token2"},
	}

	repo.On("GetAlarmsToNotify", mock.Anything, targetTime).Return(alarms, nil)
	cache.On("Get", mock.Anything, "서울시 강남구").Return(weatherData, nil)
	repo.On("GetFCMTokens", mock.Anything, 100).Return(tokens, nil)
	notifier.On("SendWeatherNotification", mock.Anything, []string{"token1", "token2"}, weatherData, "서울시 강남구").Return(nil)
	repo.On("UpdateLastSent", mock.Anything, 1, mock.Anything).Return(nil)

	scheduler := createTestScheduler(repo, crawler, cache, notifier)
	ctx := context.Background()

	err := scheduler.processAlarms(ctx, targetTime)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
	notifier.AssertExpectations(t)
	crawler.AssertNotCalled(t, "Fetch") // Should not call crawler on cache hit
}

func TestProcessAlarms_Success_CacheMiss(t *testing.T) {
	repo := new(MockRepository)
	crawler := new(MockCrawler)
	cache := new(MockCache)
	notifier := new(MockNotifier)

	targetTime := time.Now()
	alarms := []entity.UserAlarm{
		{
			ID:        1,
			UserID:    100,
			AlarmTime: "09:00:00",
			Region:    "서울시 강남구",
			IsEnabled: true,
		},
	}

	weatherData := &entity.WeatherData{
		Temperature:   25.5,
		Humidity:      60.0,
		Precipitation: 0.0,
		WindSpeed:     2.5,
		CachedAt:      time.Now(),
	}

	tokens := []entity.WeatherServiceToken{
		{ID: 1, UserID: 100, FCMToken: "token1"},
	}

	repo.On("GetAlarmsToNotify", mock.Anything, targetTime).Return(alarms, nil)
	cache.On("Get", mock.Anything, "서울시 강남구").Return(nil, nil) // Cache miss
	crawler.On("Fetch", mock.Anything, "서울시 강남구").Return(weatherData, nil)
	cache.On("Set", mock.Anything, "서울시 강남구", weatherData).Return(nil)
	repo.On("GetFCMTokens", mock.Anything, 100).Return(tokens, nil)
	notifier.On("SendWeatherNotification", mock.Anything, []string{"token1"}, weatherData, "서울시 강남구").Return(nil)
	repo.On("UpdateLastSent", mock.Anything, 1, mock.Anything).Return(nil)

	scheduler := createTestScheduler(repo, crawler, cache, notifier)
	ctx := context.Background()

	err := scheduler.processAlarms(ctx, targetTime)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
	crawler.AssertExpectations(t) // Should call crawler on cache miss
	notifier.AssertExpectations(t)
}

func TestProcessAlarms_CrawlerFailure(t *testing.T) {
	repo := new(MockRepository)
	crawler := new(MockCrawler)
	cache := new(MockCache)
	notifier := new(MockNotifier)

	targetTime := time.Now()
	alarms := []entity.UserAlarm{
		{
			ID:        1,
			UserID:    100,
			AlarmTime: "09:00:00",
			Region:    "서울시 강남구",
			IsEnabled: true,
		},
	}

	repo.On("GetAlarmsToNotify", mock.Anything, targetTime).Return(alarms, nil)
	cache.On("Get", mock.Anything, "서울시 강남구").Return(nil, nil) // Cache miss
	crawler.On("Fetch", mock.Anything, "서울시 강남구").Return(nil, errors.New("crawler error"))

	scheduler := createTestScheduler(repo, crawler, cache, notifier)
	ctx := context.Background()

	err := scheduler.processAlarms(ctx, targetTime)

	// Should not return error, but log it and continue
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
	crawler.AssertExpectations(t)
	notifier.AssertNotCalled(t, "SendWeatherNotification") // Should not notify on crawler failure
}

func TestProcessAlarms_FCMFailure_StillUpdateLastSent(t *testing.T) {
	repo := new(MockRepository)
	crawler := new(MockCrawler)
	cache := new(MockCache)
	notifier := new(MockNotifier)

	targetTime := time.Now()
	alarms := []entity.UserAlarm{
		{
			ID:        1,
			UserID:    100,
			AlarmTime: "09:00:00",
			Region:    "서울시 강남구",
			IsEnabled: true,
		},
	}

	weatherData := &entity.WeatherData{
		Temperature: 25.5,
	}

	tokens := []entity.WeatherServiceToken{
		{ID: 1, UserID: 100, FCMToken: "token1"},
	}

	repo.On("GetAlarmsToNotify", mock.Anything, targetTime).Return(alarms, nil)
	cache.On("Get", mock.Anything, "서울시 강남구").Return(weatherData, nil)
	repo.On("GetFCMTokens", mock.Anything, 100).Return(tokens, nil)
	notifier.On("SendWeatherNotification", mock.Anything, []string{"token1"}, weatherData, "서울시 강남구").Return(errors.New("FCM error"))
	repo.On("UpdateLastSent", mock.Anything, 1, mock.Anything).Return(nil)

	scheduler := createTestScheduler(repo, crawler, cache, notifier)
	ctx := context.Background()

	err := scheduler.processAlarms(ctx, targetTime)

	// Should not return error, but still update last_sent to prevent retry storm
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	repo.AssertCalled(t, "UpdateLastSent", mock.Anything, 1, mock.Anything)
}

func TestProcessAlarms_NoFCMTokens(t *testing.T) {
	repo := new(MockRepository)
	crawler := new(MockCrawler)
	cache := new(MockCache)
	notifier := new(MockNotifier)

	targetTime := time.Now()
	alarms := []entity.UserAlarm{
		{
			ID:        1,
			UserID:    100,
			AlarmTime: "09:00:00",
			Region:    "서울시 강남구",
			IsEnabled: true,
		},
	}

	weatherData := &entity.WeatherData{
		Temperature: 25.5,
	}

	repo.On("GetAlarmsToNotify", mock.Anything, targetTime).Return(alarms, nil)
	cache.On("Get", mock.Anything, "서울시 강남구").Return(weatherData, nil)
	repo.On("GetFCMTokens", mock.Anything, 100).Return([]entity.WeatherServiceToken{}, nil)
	repo.On("UpdateLastSent", mock.Anything, 1, mock.Anything).Return(nil)

	scheduler := createTestScheduler(repo, crawler, cache, notifier)
	ctx := context.Background()

	err := scheduler.processAlarms(ctx, targetTime)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	notifier.AssertNotCalled(t, "SendWeatherNotification") // Should not notify if no tokens
	repo.AssertCalled(t, "UpdateLastSent", mock.Anything, 1, mock.Anything) // Still update last_sent
}

func TestProcessAlarms_MultipleAlarms(t *testing.T) {
	repo := new(MockRepository)
	crawler := new(MockCrawler)
	cache := new(MockCache)
	notifier := new(MockNotifier)

	targetTime := time.Now()
	alarms := []entity.UserAlarm{
		{ID: 1, UserID: 100, AlarmTime: "09:00:00", Region: "서울시 강남구", IsEnabled: true},
		{ID: 2, UserID: 101, AlarmTime: "09:00:00", Region: "서울시 서초구", IsEnabled: true},
		{ID: 3, UserID: 102, AlarmTime: "09:00:00", Region: "서울시 송파구", IsEnabled: true},
	}

	weatherData1 := &entity.WeatherData{Temperature: 25.5}
	weatherData2 := &entity.WeatherData{Temperature: 26.0}
	weatherData3 := &entity.WeatherData{Temperature: 24.5}

	tokens1 := []entity.WeatherServiceToken{{ID: 1, UserID: 100, FCMToken: "token1"}}
	tokens2 := []entity.WeatherServiceToken{{ID: 2, UserID: 101, FCMToken: "token2"}}
	tokens3 := []entity.WeatherServiceToken{{ID: 3, UserID: 102, FCMToken: "token3"}}

	repo.On("GetAlarmsToNotify", mock.Anything, targetTime).Return(alarms, nil)

	cache.On("Get", mock.Anything, "서울시 강남구").Return(weatherData1, nil)
	cache.On("Get", mock.Anything, "서울시 서초구").Return(weatherData2, nil)
	cache.On("Get", mock.Anything, "서울시 송파구").Return(weatherData3, nil)

	repo.On("GetFCMTokens", mock.Anything, 100).Return(tokens1, nil)
	repo.On("GetFCMTokens", mock.Anything, 101).Return(tokens2, nil)
	repo.On("GetFCMTokens", mock.Anything, 102).Return(tokens3, nil)

	notifier.On("SendWeatherNotification", mock.Anything, []string{"token1"}, weatherData1, "서울시 강남구").Return(nil)
	notifier.On("SendWeatherNotification", mock.Anything, []string{"token2"}, weatherData2, "서울시 서초구").Return(nil)
	notifier.On("SendWeatherNotification", mock.Anything, []string{"token3"}, weatherData3, "서울시 송파구").Return(nil)

	repo.On("UpdateLastSent", mock.Anything, 1, mock.Anything).Return(nil)
	repo.On("UpdateLastSent", mock.Anything, 2, mock.Anything).Return(nil)
	repo.On("UpdateLastSent", mock.Anything, 3, mock.Anything).Return(nil)

	scheduler := createTestScheduler(repo, crawler, cache, notifier)
	ctx := context.Background()

	err := scheduler.processAlarms(ctx, targetTime)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
	notifier.AssertExpectations(t)
	assert.Equal(t, 3, len(notifier.GetNotificationLog()))
}

func TestProcessAlarms_PartialFailures(t *testing.T) {
	repo := new(MockRepository)
	crawler := new(MockCrawler)
	cache := new(MockCache)
	notifier := new(MockNotifier)

	targetTime := time.Now()
	alarms := []entity.UserAlarm{
		{ID: 1, UserID: 100, AlarmTime: "09:00:00", Region: "서울시 강남구", IsEnabled: true},
		{ID: 2, UserID: 101, AlarmTime: "09:00:00", Region: "서울시 서초구", IsEnabled: true}, // This will fail
		{ID: 3, UserID: 102, AlarmTime: "09:00:00", Region: "서울시 송파구", IsEnabled: true},
	}

	weatherData1 := &entity.WeatherData{Temperature: 25.5}
	weatherData3 := &entity.WeatherData{Temperature: 24.5}

	tokens1 := []entity.WeatherServiceToken{{ID: 1, UserID: 100, FCMToken: "token1"}}
	tokens3 := []entity.WeatherServiceToken{{ID: 3, UserID: 102, FCMToken: "token3"}}

	repo.On("GetAlarmsToNotify", mock.Anything, targetTime).Return(alarms, nil)

	cache.On("Get", mock.Anything, "서울시 강남구").Return(weatherData1, nil)
	cache.On("Get", mock.Anything, "서울시 서초구").Return(nil, nil) // Cache miss
	crawler.On("Fetch", mock.Anything, "서울시 서초구").Return(nil, errors.New("crawler error")) // Crawler fails
	cache.On("Get", mock.Anything, "서울시 송파구").Return(weatherData3, nil)

	repo.On("GetFCMTokens", mock.Anything, 100).Return(tokens1, nil)
	repo.On("GetFCMTokens", mock.Anything, 102).Return(tokens3, nil)

	notifier.On("SendWeatherNotification", mock.Anything, []string{"token1"}, weatherData1, "서울시 강남구").Return(nil)
	notifier.On("SendWeatherNotification", mock.Anything, []string{"token3"}, weatherData3, "서울시 송파구").Return(nil)

	repo.On("UpdateLastSent", mock.Anything, 1, mock.Anything).Return(nil)
	repo.On("UpdateLastSent", mock.Anything, 3, mock.Anything).Return(nil)

	scheduler := createTestScheduler(repo, crawler, cache, notifier)
	ctx := context.Background()

	err := scheduler.processAlarms(ctx, targetTime)

	// Should succeed even with partial failures
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
	crawler.AssertExpectations(t)
	notifier.AssertExpectations(t)
	assert.Equal(t, 2, len(notifier.GetNotificationLog())) // Only 2 successful notifications
}

func TestGracefulShutdown_Timeout(t *testing.T) {
	repo := new(MockRepository)
	crawler := new(MockCrawler)
	cache := new(MockCache)
	notifier := new(MockNotifier)

	// Mock long-running operation
	repo.On("GetAlarmsToNotify", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		time.Sleep(35 * time.Second) // Longer than shutdown timeout
	}).Return([]entity.UserAlarm{}, nil)

	scheduler := createTestScheduler(repo, crawler, cache, notifier)

	ctx := context.Background()
	go scheduler.Start(ctx)

	time.Sleep(50 * time.Millisecond)

	// Stop should timeout after 30 seconds
	start := time.Now()
	err := scheduler.Stop()
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.True(t, duration >= 30*time.Second, "Stop should wait for timeout")
	assert.True(t, duration < 35*time.Second, "Stop should not wait indefinitely")
}

func TestContextCancellation(t *testing.T) {
	repo := new(MockRepository)
	crawler := new(MockCrawler)
	cache := new(MockCache)
	notifier := new(MockNotifier)

	repo.On("GetAlarmsToNotify", mock.Anything, mock.Anything).Return([]entity.UserAlarm{}, nil)

	scheduler := createTestScheduler(repo, crawler, cache, notifier)

	ctx, cancel := context.WithCancel(context.Background())

	errChan := make(chan error)
	go func() {
		errChan <- scheduler.Start(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	// Cancel context
	cancel()

	err := <-errChan
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestTickerFiresAtIntervals(t *testing.T) {
	repo := new(MockRepository)
	crawler := new(MockCrawler)
	cache := new(MockCache)
	notifier := new(MockNotifier)

	callCount := 0
	var mu sync.Mutex
	repo.On("GetAlarmsToNotify", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		mu.Lock()
		callCount++
		mu.Unlock()
	}).Return([]entity.UserAlarm{}, nil)

	scheduler := NewWeatherSchedulerService(
		repo,
		crawler,
		cache,
		notifier,
		createTestLogger(),
		50*time.Millisecond, // Very short interval for testing
	)

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	go scheduler.Start(ctx)

	<-ctx.Done()
	scheduler.Stop()

	mu.Lock()
	count := callCount
	mu.Unlock()

	// Should fire multiple times within 250ms with 50ms interval
	// At least 3 times: initial + 2 ticks
	assert.GreaterOrEqual(t, count, 3, fmt.Sprintf("Expected at least 3 calls, got %d", count))
}

func TestConcurrentSafety(t *testing.T) {
	repo := new(MockRepository)
	crawler := new(MockCrawler)
	cache := new(MockCache)
	notifier := new(MockNotifier)

	repo.On("GetAlarmsToNotify", mock.Anything, mock.Anything).Return([]entity.UserAlarm{}, nil)

	scheduler := createTestScheduler(repo, crawler, cache, notifier)

	ctx := context.Background()
	go scheduler.Start(ctx)

	time.Sleep(50 * time.Millisecond)

	// Call Stop multiple times concurrently
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := scheduler.Stop()
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
	assert.False(t, scheduler.running)
}

func TestCacheFailure_FallbackToCrawler(t *testing.T) {
	repo := new(MockRepository)
	crawler := new(MockCrawler)
	cache := new(MockCache)
	notifier := new(MockNotifier)

	targetTime := time.Now()
	alarms := []entity.UserAlarm{
		{ID: 1, UserID: 100, AlarmTime: "09:00:00", Region: "서울시 강남구", IsEnabled: true},
	}

	weatherData := &entity.WeatherData{Temperature: 25.5}
	tokens := []entity.WeatherServiceToken{{ID: 1, UserID: 100, FCMToken: "token1"}}

	repo.On("GetAlarmsToNotify", mock.Anything, targetTime).Return(alarms, nil)
	cache.On("Get", mock.Anything, "서울시 강남구").Return(nil, errors.New("cache error"))
	crawler.On("Fetch", mock.Anything, "서울시 강남구").Return(weatherData, nil)
	cache.On("Set", mock.Anything, "서울시 강남구", weatherData).Return(nil)
	repo.On("GetFCMTokens", mock.Anything, 100).Return(tokens, nil)
	notifier.On("SendWeatherNotification", mock.Anything, []string{"token1"}, weatherData, "서울시 강남구").Return(nil)
	repo.On("UpdateLastSent", mock.Anything, 1, mock.Anything).Return(nil)

	scheduler := createTestScheduler(repo, crawler, cache, notifier)
	ctx := context.Background()

	err := scheduler.processAlarms(ctx, targetTime)

	assert.NoError(t, err)
	crawler.AssertExpectations(t) // Should fallback to crawler on cache error
}

func TestCacheSetFailure_ContinuesProcessing(t *testing.T) {
	repo := new(MockRepository)
	crawler := new(MockCrawler)
	cache := new(MockCache)
	notifier := new(MockNotifier)

	targetTime := time.Now()
	alarms := []entity.UserAlarm{
		{ID: 1, UserID: 100, AlarmTime: "09:00:00", Region: "서울시 강남구", IsEnabled: true},
	}

	weatherData := &entity.WeatherData{Temperature: 25.5}
	tokens := []entity.WeatherServiceToken{{ID: 1, UserID: 100, FCMToken: "token1"}}

	repo.On("GetAlarmsToNotify", mock.Anything, targetTime).Return(alarms, nil)
	cache.On("Get", mock.Anything, "서울시 강남구").Return(nil, nil)
	crawler.On("Fetch", mock.Anything, "서울시 강남구").Return(weatherData, nil)
	cache.On("Set", mock.Anything, "서울시 강남구", weatherData).Return(errors.New("cache set error"))
	repo.On("GetFCMTokens", mock.Anything, 100).Return(tokens, nil)
	notifier.On("SendWeatherNotification", mock.Anything, []string{"token1"}, weatherData, "서울시 강남구").Return(nil)
	repo.On("UpdateLastSent", mock.Anything, 1, mock.Anything).Return(nil)

	scheduler := createTestScheduler(repo, crawler, cache, notifier)
	ctx := context.Background()

	err := scheduler.processAlarms(ctx, targetTime)

	// Should succeed even if cache set fails
	assert.NoError(t, err)
	notifier.AssertExpectations(t)
}
