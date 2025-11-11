package bench

import (
	"context"
	"fmt"
	"testing"
	"time"

	"firebase.google.com/go/v4/messaging"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/cache"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/crawler"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/interface"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/repository"
	"github.com/alicebob/miniredis/v2"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockFCMClientBench implements fast FCM client for benchmarking
type MockFCMClientBench struct{}

func (m *MockFCMClientBench) SendMulticast(ctx context.Context, message *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
	// Minimal overhead mock
	responses := make([]*messaging.SendResponse, len(message.Tokens))
	for i := range responses {
		responses[i] = &messaging.SendResponse{Success: true}
	}

	return &messaging.BatchResponse{
		SuccessCount: len(message.Tokens),
		FailureCount: 0,
		Responses:    responses,
	}, nil
}

// setupBenchDB creates test database for benchmarking
func setupBenchDB(b *testing.B) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		b.Fatalf("Failed to create database: %v", err)
	}

	err = db.AutoMigrate(&entity.UserAlarm{}, &entity.WeatherServiceToken{})
	if err != nil {
		b.Fatalf("Failed to migrate: %v", err)
	}

	return db
}

// setupBenchRedis creates Redis instance for benchmarking
func setupBenchRedis(b *testing.B) (*cache.WeatherCache, *miniredis.Miniredis) {
	logger := zap.NewNop()

	mr, err := miniredis.Run()
	if err != nil {
		b.Fatalf("Failed to start miniredis: %v", err)
	}

	weatherCache, err := cache.NewWeatherCache(mr.Addr(), "", logger)
	if err != nil {
		mr.Close()
		b.Fatalf("Failed to create cache: %v", err)
	}

	return weatherCache, mr
}

// BenchmarkCacheGet benchmarks cache retrieval performance
func BenchmarkCacheGet(b *testing.B) {
	weatherCache, mr := setupBenchRedis(b)
	defer mr.Close()
	defer weatherCache.Close()

	ctx := context.Background()

	// Pre-populate cache
	weatherData := &entity.WeatherData{
		Temperature:   25.5,
		Humidity:      60.0,
		Precipitation: 0.0,
		WindSpeed:     3.2,
		CachedAt:      time.Now(),
	}

	err := weatherCache.Set(ctx, "서울시 강남구", weatherData)
	if err != nil {
		b.Fatalf("Failed to set cache: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := weatherCache.Get(ctx, "서울시 강남구")
		if err != nil {
			b.Fatalf("Cache get failed: %v", err)
		}
	}
}

// BenchmarkCacheSet benchmarks cache storage performance
func BenchmarkCacheSet(b *testing.B) {
	weatherCache, mr := setupBenchRedis(b)
	defer mr.Close()
	defer weatherCache.Close()

	ctx := context.Background()

	weatherData := &entity.WeatherData{
		Temperature:   25.5,
		Humidity:      60.0,
		Precipitation: 0.0,
		WindSpeed:     3.2,
		CachedAt:      time.Now(),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		region := fmt.Sprintf("region_%d", i)
		err := weatherCache.Set(ctx, region, weatherData)
		if err != nil {
			b.Fatalf("Cache set failed: %v", err)
		}
	}
}

// BenchmarkCacheGetMiss benchmarks cache miss performance
func BenchmarkCacheGetMiss(b *testing.B) {
	weatherCache, mr := setupBenchRedis(b)
	defer mr.Close()
	defer weatherCache.Close()

	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		region := fmt.Sprintf("nonexistent_%d", i)
		_, err := weatherCache.Get(ctx, region)
		if err != nil {
			// Expected for cache miss
		}
	}
}

// BenchmarkRepositoryGetAlarmsToNotify benchmarks alarm query performance
func BenchmarkRepositoryGetAlarmsToNotify(b *testing.B) {
	db := setupBenchDB(b)

	repo := repository.NewSchedulerWeatherRepository(db)
	ctx := context.Background()

	// Create test alarms
	targetTime := time.Now()
	for i := 0; i < 100; i++ {
		alarm := entity.UserAlarm{
			UserID:    i + 1,
			AlarmTime: targetTime.Format("15:04:05"),
			Region:    fmt.Sprintf("region_%d", i),
			IsEnabled: true,
		}
		db.Create(&alarm)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := repo.GetAlarmsToNotify(ctx, targetTime)
		if err != nil {
			b.Fatalf("GetAlarmsToNotify failed: %v", err)
		}
	}
}

// BenchmarkRepositoryUpdateLastSent benchmarks last_sent update performance
func BenchmarkRepositoryUpdateLastSent(b *testing.B) {
	db := setupBenchDB(b)

	repo := repository.NewSchedulerWeatherRepository(db)
	ctx := context.Background()

	// Create test alarm
	alarm := entity.UserAlarm{
		UserID:    1,
		AlarmTime: "09:00:00",
		Region:    "서울시 강남구",
		IsEnabled: true,
	}
	db.Create(&alarm)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := repo.UpdateLastSent(ctx, alarm.ID, time.Now())
		if err != nil {
			b.Fatalf("UpdateLastSent failed: %v", err)
		}
	}
}

// BenchmarkRepositoryGetFCMTokens benchmarks FCM token retrieval
func BenchmarkRepositoryGetFCMTokens(b *testing.B) {
	db := setupBenchDB(b)

	repo := repository.NewSchedulerWeatherRepository(db)
	ctx := context.Background()

	userID := 1

	// Create test tokens
	for i := 0; i < 10; i++ {
		token := entity.WeatherServiceToken{
			UserID:   userID,
			FCMToken: fmt.Sprintf("token_%d", i),
			DeviceID: fmt.Sprintf("device_%d", i),
		}
		db.Create(&token)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := repo.GetFCMTokens(ctx, userID)
		if err != nil {
			b.Fatalf("GetFCMTokens failed: %v", err)
		}
	}
}

// BenchmarkFCMNotification benchmarks FCM notification sending
func BenchmarkFCMNotification(b *testing.B) {
	logger := zap.NewNop()
	mockFCM := &MockFCMClientBench{}

	// Note: Using real notifier would require refactoring
	// This benchmark focuses on the mock overhead
	ctx := context.Background()

	tokens := []string{"token1", "token2", "token3"}
	message := &messaging.MulticastMessage{
		Tokens: tokens,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := mockFCM.SendMulticast(ctx, message)
		if err != nil {
			b.Fatalf("SendMulticast failed: %v", err)
		}
	}

	_ = logger // Use logger to avoid unused warning
}

// BenchmarkCrawlerFetch benchmarks weather data crawler
// NOTE: This is commented out because it makes real HTTP requests
// Uncomment only for real-world performance testing
/*
func BenchmarkCrawlerFetch(b *testing.B) {
	weatherCrawler := crawler.NewNaverWeatherCrawler(10*time.Second, 3)
	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := weatherCrawler.Fetch(ctx, "서울시 강남구")
		if err != nil {
			b.Logf("Crawler fetch failed (expected for benchmark): %v", err)
		}
	}
}
*/

// BenchmarkDatabaseInsert benchmarks alarm insertion performance
func BenchmarkDatabaseInsert(b *testing.B) {
	db := setupBenchDB(b)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		alarm := entity.UserAlarm{
			UserID:    i + 1,
			AlarmTime: "09:00:00",
			Region:    "서울시 강남구",
			IsEnabled: true,
		}
		err := db.Create(&alarm).Error
		if err != nil {
			b.Fatalf("Insert failed: %v", err)
		}
	}
}

// BenchmarkDatabaseQuery benchmarks alarm query performance
func BenchmarkDatabaseQuery(b *testing.B) {
	db := setupBenchDB(b)

	// Pre-populate database
	for i := 0; i < 1000; i++ {
		alarm := entity.UserAlarm{
			UserID:    i + 1,
			AlarmTime: "09:00:00",
			Region:    "서울시 강남구",
			IsEnabled: true,
		}
		db.Create(&alarm)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var alarms []entity.UserAlarm
		err := db.Where("is_enabled = ? AND alarm_time = ?", true, "09:00:00").Find(&alarms).Error
		if err != nil {
			b.Fatalf("Query failed: %v", err)
		}
	}
}

// BenchmarkParallelCacheOperations benchmarks parallel cache operations
func BenchmarkParallelCacheOperations(b *testing.B) {
	weatherCache, mr := setupBenchRedis(b)
	defer mr.Close()
	defer weatherCache.Close()

	ctx := context.Background()

	weatherData := &entity.WeatherData{
		Temperature: 25.5,
		CachedAt:    time.Now(),
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			region := fmt.Sprintf("region_%d", i%10)
			if i%2 == 0 {
				weatherCache.Set(ctx, region, weatherData)
			} else {
				weatherCache.Get(ctx, region)
			}
			i++
		}
	})
}

// BenchmarkParallelDatabaseOperations benchmarks parallel database operations
func BenchmarkParallelDatabaseOperations(b *testing.B) {
	db := setupBenchDB(b)

	// Pre-populate
	for i := 0; i < 100; i++ {
		alarm := entity.UserAlarm{
			UserID:    i + 1,
			AlarmTime: "09:00:00",
			Region:    "서울시 강남구",
			IsEnabled: true,
		}
		db.Create(&alarm)
	}

	ctx := context.Background()
	repo := repository.NewSchedulerWeatherRepository(db)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			repo.GetAlarmsToNotify(ctx, time.Now())
		}
	})
}

// BenchmarkWeatherDataSerialization benchmarks weather data JSON serialization
func BenchmarkWeatherDataSerialization(b *testing.B) {
	weatherCache, mr := setupBenchRedis(b)
	defer mr.Close()
	defer weatherCache.Close()

	ctx := context.Background()

	weatherData := &entity.WeatherData{
		Temperature:   25.5,
		Humidity:      60.0,
		Precipitation: 0.0,
		WindSpeed:     3.2,
		CachedAt:      time.Now(),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Set involves serialization
		err := weatherCache.Set(ctx, "test_region", weatherData)
		if err != nil {
			b.Fatalf("Set failed: %v", err)
		}

		// Get involves deserialization
		_, err = weatherCache.Get(ctx, "test_region")
		if err != nil {
			b.Fatalf("Get failed: %v", err)
		}
	}
}

// Ensure interface compliance
var _ _interface.IFCMClient = (*MockFCMClientBench)(nil)

// Benchmark results reporting
func init() {
	_ = crawler.NewNaverWeatherCrawler(10*time.Second, 3)
}
