package loadtest

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"firebase.google.com/go/v4/messaging"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/cache"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/crawler"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/interface"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/repository"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/scheduler"
	"github.com/JokerTrickster/joker_backend/services/weatherService/pkg/metrics"
	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// LoadTestMetrics tracks load test performance metrics
type LoadTestMetrics struct {
	TotalAlarms       int64
	SuccessCount      int64
	FailureCount      int64
	TotalDuration     time.Duration
	AverageDuration   time.Duration
	MaxDuration       time.Duration
	MinDuration       time.Duration
	GoroutinesBefore  int
	GoroutinesAfter   int
	MemoryBefore      uint64
	MemoryAfter       uint64
	mu                sync.Mutex
	durations         []time.Duration
}

// MockFCMClientLoadTest implements FCM client for load testing
type MockFCMClientLoadTest struct {
	mu            sync.Mutex
	sendCount     int64
	failureCount  int64
	durations     []time.Duration
	shouldFail    bool
	failureRate   float64 // 0.0 to 1.0
	latencyMs     int
	callCounter   int64
}

func (m *MockFCMClientLoadTest) SendMulticast(ctx context.Context, message *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
	atomic.AddInt64(&m.callCounter, 1)

	// Simulate latency
	if m.latencyMs > 0 {
		time.Sleep(time.Duration(m.latencyMs) * time.Millisecond)
	}

	// Simulate failure rate
	currentCount := atomic.LoadInt64(&m.callCounter)
	if m.failureRate > 0 && float64(currentCount)/(1.0/m.failureRate) == float64(int(float64(currentCount)/(1.0/m.failureRate))) {
		atomic.AddInt64(&m.failureCount, 1)
		return nil, fmt.Errorf("simulated FCM failure")
	}

	atomic.AddInt64(&m.sendCount, 1)

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

// SendWeatherNotification implements IFCMNotifier interface for load testing
func (m *MockFCMClientLoadTest) SendWeatherNotification(ctx context.Context, tokens []string, data *entity.WeatherData, region string) error {
	atomic.AddInt64(&m.callCounter, 1)

	// Simulate latency
	if m.latencyMs > 0 {
		time.Sleep(time.Duration(m.latencyMs) * time.Millisecond)
	}

	// Simulate failure rate
	currentCount := atomic.LoadInt64(&m.callCounter)
	if m.failureRate > 0 && float64(currentCount)/(1.0/m.failureRate) == float64(int(float64(currentCount)/(1.0/m.failureRate))) {
		atomic.AddInt64(&m.failureCount, 1)
		return fmt.Errorf("simulated FCM failure")
	}

	atomic.AddInt64(&m.sendCount, 1)
	return nil
}

func (m *MockFCMClientLoadTest) GetMetrics() (sends int64, failures int64) {
	return atomic.LoadInt64(&m.sendCount), atomic.LoadInt64(&m.failureCount)
}

func (m *MockFCMClientLoadTest) Reset() {
	atomic.StoreInt64(&m.sendCount, 0)
	atomic.StoreInt64(&m.failureCount, 0)
	atomic.StoreInt64(&m.callCounter, 0)
}

// setupLoadTestDB creates an in-memory SQLite database for load testing
func setupLoadTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err, "Failed to create in-memory database")

	// Auto-migrate tables
	err = db.AutoMigrate(&entity.UserAlarm{}, &entity.WeatherServiceToken{})
	require.NoError(t, err, "Failed to migrate tables")

	return db
}

// createTestAlarms creates N alarms for load testing
func createTestAlarms(t *testing.T, db *gorm.DB, count int, targetTime time.Time) []entity.UserAlarm {
	alarms := make([]entity.UserAlarm, 0, count)
	regions := []string{
		"서울시 강남구",
		"서울시 서초구",
		"경기도 성남시",
		"부산시 해운대구",
		"대구시 중구",
		"인천시 남동구",
		"광주시 북구",
		"대전시 서구",
		"울산시 남구",
		"경기도 수원시",
	}

	for i := 0; i < count; i++ {
		userID := i + 1
		region := regions[i%len(regions)]

		alarm := entity.UserAlarm{
			UserID:    userID,
			AlarmTime: targetTime.Format("15:04:05"),
			Region:    region,
			IsEnabled: true,
			LastSent:  nil,
		}

		err := db.Create(&alarm).Error
		require.NoError(t, err, "Failed to create alarm")

		// Create FCM tokens for each user
		for j := 0; j < 2; j++ {
			token := entity.WeatherServiceToken{
				UserID:   userID,
				FCMToken: fmt.Sprintf("token_%d_%d", userID, j),
				DeviceID: fmt.Sprintf("device_%d_%d", userID, j),
			}
			err := db.Create(&token).Error
			require.NoError(t, err, "Failed to create FCM token")
		}

		alarms = append(alarms, alarm)
	}

	return alarms
}

// getMemoryUsage returns current memory usage in bytes
func getMemoryUsage() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

// TestLoadTest_1000Alarms tests processing 1000 alarms
func TestLoadTest_1000Alarms(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	alarmCount := 1000
	t.Logf("Starting load test with %d alarms", alarmCount)

	// Initialize metrics
	metrics.InitMetrics()

	// Setup
	db := setupLoadTestDB(t)
	logger, _ := zap.NewDevelopment()

	// Setup Redis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	weatherCache, err := cache.NewWeatherCache(mr.Addr(), "", logger)
	require.NoError(t, err)
	defer weatherCache.Close()

	// Pre-populate cache to reduce crawler load
	regions := []string{
		"서울시 강남구", "서울시 서초구", "경기도 성남시", "부산시 해운대구", "대구시 중구",
		"인천시 남동구", "광주시 북구", "대전시 서구", "울산시 남구", "경기도 수원시",
	}
	for _, region := range regions {
		weatherData := &entity.WeatherData{
			Temperature:   25.5,
			Humidity:      60.0,
			Precipitation: 0.0,
			WindSpeed:     3.2,
			CachedAt:      time.Now(),
		}
		err := weatherCache.Set(context.Background(), region, weatherData)
		require.NoError(t, err)
	}

	// Create mock FCM client
	mockFCM := &MockFCMClientLoadTest{
		latencyMs: 10, // Simulate 10ms FCM latency
	}

	// Create dependencies
	repo := repository.NewSchedulerWeatherRepository(db)
	weatherCrawler := crawler.NewNaverWeatherCrawler(10*time.Second, 3)

	// Measure baseline metrics
	runtime.GC()
	initialGoroutines := runtime.NumGoroutine()
	initialMemory := getMemoryUsage()

	t.Logf("Initial state: goroutines=%d, memory=%d MB", initialGoroutines, initialMemory/(1024*1024))

	// Create alarms - target time is 5 seconds from now
	targetTime := time.Now().Add(5 * time.Second)
	alarms := createTestAlarms(t, db, alarmCount, targetTime)
	require.Len(t, alarms, alarmCount)

	t.Logf("Created %d alarms for time %s", alarmCount, targetTime.Format("15:04:05"))

	// Create scheduler with 1 second tick interval
	schedulerService := scheduler.NewWeatherSchedulerService(
		repo,
		weatherCrawler,
		weatherCache,
		mockFCM,
		logger,
		1*time.Second,
	)

	// Start scheduler
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	startTime := time.Now()
	go schedulerService.Start(ctx)
	defer schedulerService.Stop()

	// Wait for processing - wait until target time + buffer
	waitDuration := time.Until(targetTime) + 30*time.Second
	if waitDuration < 35*time.Second {
		waitDuration = 35 * time.Second
	}
	t.Logf("Waiting %v for alarm processing...", waitDuration)
	time.Sleep(waitDuration)

	totalDuration := time.Since(startTime)

	// Collect metrics
	runtime.GC()
	finalGoroutines := runtime.NumGoroutine()
	finalMemory := getMemoryUsage()

	sends, failures := mockFCM.GetMetrics()

	t.Logf("Load test completed in %v", totalDuration)
	t.Logf("FCM sends: %d, failures: %d", sends, failures)
	t.Logf("Final state: goroutines=%d, memory=%d MB", finalGoroutines, finalMemory/(1024*1024))
	t.Logf("Goroutine delta: %d", finalGoroutines-initialGoroutines)
	t.Logf("Memory delta: %d MB", (finalMemory-initialMemory)/(1024*1024))

	// Verify all alarms processed
	var updatedAlarms []entity.UserAlarm
	err = db.Where("last_sent IS NOT NULL").Find(&updatedAlarms).Error
	require.NoError(t, err)

	successRate := float64(len(updatedAlarms)) / float64(alarmCount) * 100
	t.Logf("Alarms processed: %d/%d (%.2f%%)", len(updatedAlarms), alarmCount, successRate)

	// Calculate average latency (approximation)
	avgLatency := totalDuration / time.Duration(alarmCount)
	t.Logf("Average latency per alarm: %v", avgLatency)

	// Assertions
	assert.GreaterOrEqual(t, len(updatedAlarms), int(float64(alarmCount)*0.95),
		"At least 95%% of alarms should be processed")
	assert.Less(t, avgLatency, 500*time.Millisecond,
		"Average latency should be less than 500ms")
	assert.GreaterOrEqual(t, sends, int64(float64(alarmCount)*0.95),
		"At least 95%% of notifications should be sent")

	// Check for resource leaks
	goroutineLeak := finalGoroutines - initialGoroutines
	assert.Less(t, goroutineLeak, 10,
		"Should not have significant goroutine leak")

	memoryLeakMB := (finalMemory - initialMemory) / (1024 * 1024)
	assert.Less(t, memoryLeakMB, uint64(50),
		"Memory increase should be less than 50MB")

	t.Logf("✅ Load test passed: %d alarms processed with %.2f%% success rate",
		len(updatedAlarms), successRate)
}

// TestLoadTest_ConcurrentRegions tests concurrent processing of different regions
func TestLoadTest_ConcurrentRegions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	alarmCount := 500
	t.Logf("Starting concurrent regions test with %d alarms", alarmCount)

	// Setup
	db := setupLoadTestDB(t)
	logger, _ := zap.NewDevelopment()

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	weatherCache, err := cache.NewWeatherCache(mr.Addr(), "", logger)
	require.NoError(t, err)
	defer weatherCache.Close()

	mockFCM := &MockFCMClientLoadTest{
		latencyMs: 20,
	}

	repo := repository.NewSchedulerWeatherRepository(db)
	weatherCrawler := crawler.NewNaverWeatherCrawler(10*time.Second, 3)

	// Create alarms
	targetTime := time.Now().Add(5 * time.Second)
	alarms := createTestAlarms(t, db, alarmCount, targetTime)
	require.Len(t, alarms, alarmCount)

	// Create scheduler
	schedulerService := scheduler.NewWeatherSchedulerService(
		repo,
		weatherCrawler,
		weatherCache,
		mockFCM,
		logger,
		1*time.Second,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	startTime := time.Now()
	go schedulerService.Start(ctx)
	defer schedulerService.Stop()

	// Wait for processing
	waitDuration := time.Until(targetTime) + 25*time.Second
	if waitDuration < 30*time.Second {
		waitDuration = 30 * time.Second
	}
	time.Sleep(waitDuration)

	totalDuration := time.Since(startTime)

	// Verify results
	var updatedAlarms []entity.UserAlarm
	err = db.Where("last_sent IS NOT NULL").Find(&updatedAlarms).Error
	require.NoError(t, err)

	sends, _ := mockFCM.GetMetrics()
	successRate := float64(len(updatedAlarms)) / float64(alarmCount) * 100

	t.Logf("Concurrent test completed in %v", totalDuration)
	t.Logf("Alarms processed: %d/%d (%.2f%%)", len(updatedAlarms), alarmCount, successRate)
	t.Logf("FCM sends: %d", sends)

	assert.GreaterOrEqual(t, successRate, 95.0, "Should process at least 95% of alarms")
}

// TestLoadTest_HighFailureRate tests system resilience under high failure rate
func TestLoadTest_HighFailureRate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	alarmCount := 200
	t.Logf("Starting high failure rate test with %d alarms", alarmCount)

	// Setup
	db := setupLoadTestDB(t)
	logger, _ := zap.NewDevelopment()

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	weatherCache, err := cache.NewWeatherCache(mr.Addr(), "", logger)
	require.NoError(t, err)
	defer weatherCache.Close()

	// Mock with 20% failure rate
	mockFCM := &MockFCMClientLoadTest{
		latencyMs:   15,
		failureRate: 0.2, // 20% failure rate
	}

	repo := repository.NewSchedulerWeatherRepository(db)
	weatherCrawler := crawler.NewNaverWeatherCrawler(10*time.Second, 3)

	// Create alarms
	targetTime := time.Now().Add(5 * time.Second)
	alarms := createTestAlarms(t, db, alarmCount, targetTime)
	require.Len(t, alarms, alarmCount)

	// Create scheduler
	schedulerService := scheduler.NewWeatherSchedulerService(
		repo,
		weatherCrawler,
		weatherCache,
		mockFCM,
		logger,
		1*time.Second,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	go schedulerService.Start(ctx)
	defer schedulerService.Stop()

	// Wait for processing
	waitDuration := time.Until(targetTime) + 20*time.Second
	if waitDuration < 25*time.Second {
		waitDuration = 25 * time.Second
	}
	time.Sleep(waitDuration)

	// Verify system continued despite failures
	var updatedAlarms []entity.UserAlarm
	err = db.Where("last_sent IS NOT NULL").Find(&updatedAlarms).Error
	require.NoError(t, err)

	sends, failures := mockFCM.GetMetrics()

	t.Logf("High failure test completed")
	t.Logf("Alarms processed: %d/%d", len(updatedAlarms), alarmCount)
	t.Logf("FCM sends: %d, failures: %d", sends, failures)
	t.Logf("Actual failure rate: %.2f%%", float64(failures)/float64(sends+failures)*100)

	// Even with failures, last_sent should be updated to prevent retry storm
	assert.GreaterOrEqual(t, len(updatedAlarms), int(float64(alarmCount)*0.90),
		"At least 90% of alarms should have last_sent updated despite FCM failures")
}

// Ensure interface compliance
var _ _interface.IFCMClient = (*MockFCMClientLoadTest)(nil)
