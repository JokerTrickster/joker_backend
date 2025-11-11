package cache

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// setupTestCache creates a test cache with miniredis
func setupTestCache(t *testing.T) (*WeatherCache, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	cache := &WeatherCache{
		client: redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		}),
		logger: logger,
	}

	return cache, mr
}

func TestNewWeatherCache(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	tests := []struct {
		name        string
		redisAddr   string
		password    string
		expectError bool
	}{
		{
			name:        "successful connection",
			redisAddr:   mr.Addr(),
			password:    "",
			expectError: false,
		},
		{
			name:        "invalid address",
			redisAddr:   "invalid:99999",
			password:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := NewWeatherCache(tt.redisAddr, tt.password, logger)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, cache)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cache)
				if cache != nil {
					cache.Close()
				}
			}
		})
	}
}

func TestGenerateKey(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()
	defer cache.Close()

	tests := []struct {
		name     string
		region   string
		expected string
	}{
		{
			name:     "simple region",
			region:   "서울시 강남구",
			expected: "weather:서울시:강남구",
		},
		{
			name:     "region with extra spaces",
			region:   "서울시  강남구",
			expected: "weather:서울시:강남구",
		},
		{
			name:     "region with comma",
			region:   "서울시,강남구",
			expected: "weather:서울시:강남구",
		},
		{
			name:     "region with dash",
			region:   "서울시-강남구",
			expected: "weather:서울시:강남구",
		},
		{
			name:     "region with slash",
			region:   "서울시/강남구",
			expected: "weather:서울시:강남구",
		},
		{
			name:     "single part region",
			region:   "서울시",
			expected: "weather:서울시",
		},
		{
			name:     "three part region",
			region:   "경기도 성남시 분당구",
			expected: "weather:경기도:성남시:분당구",
		},
		{
			name:     "region with leading/trailing spaces",
			region:   "  서울시 강남구  ",
			expected: "weather:서울시:강남구",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cache.generateKey(tt.region)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSetAndGet(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()
	defer cache.Close()

	ctx := context.Background()

	tests := []struct {
		name   string
		region string
		data   *entity.WeatherData
	}{
		{
			name:   "basic weather data",
			region: "서울시 강남구",
			data: &entity.WeatherData{
				Temperature:   25.5,
				Humidity:      60.0,
				Precipitation: 0.0,
				WindSpeed:     3.5,
				CachedAt:      time.Now().Truncate(time.Second),
			},
		},
		{
			name:   "weather with precipitation",
			region: "부산시 해운대구",
			data: &entity.WeatherData{
				Temperature:   18.3,
				Humidity:      85.5,
				Precipitation: 12.5,
				WindSpeed:     8.2,
				CachedAt:      time.Now().Truncate(time.Second),
			},
		},
		{
			name:   "negative temperature",
			region: "강원도 평창군",
			data: &entity.WeatherData{
				Temperature:   -5.0,
				Humidity:      45.0,
				Precipitation: 0.0,
				WindSpeed:     2.1,
				CachedAt:      time.Now().Truncate(time.Second),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set data
			err := cache.Set(ctx, tt.region, tt.data)
			require.NoError(t, err)

			// Get data
			result, err := cache.Get(ctx, tt.region)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Verify data
			assert.InDelta(t, tt.data.Temperature, result.Temperature, 0.01)
			assert.InDelta(t, tt.data.Humidity, result.Humidity, 0.01)
			assert.InDelta(t, tt.data.Precipitation, result.Precipitation, 0.01)
			assert.InDelta(t, tt.data.WindSpeed, result.WindSpeed, 0.01)
			assert.WithinDuration(t, tt.data.CachedAt, result.CachedAt, time.Second)
		})
	}
}

func TestGet_CacheMiss(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()
	defer cache.Close()

	ctx := context.Background()

	result, err := cache.Get(ctx, "존재하지않는 지역")
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestSet_NilData(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()
	defer cache.Close()

	ctx := context.Background()

	err := cache.Set(ctx, "서울시 강남구", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

func TestSet_AutoSetCachedAt(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()
	defer cache.Close()

	ctx := context.Background()

	data := &entity.WeatherData{
		Temperature:   20.0,
		Humidity:      50.0,
		Precipitation: 0.0,
		WindSpeed:     2.0,
		// CachedAt is not set (zero value)
	}

	err := cache.Set(ctx, "서울시 강남구", data)
	require.NoError(t, err)

	result, err := cache.Get(ctx, "서울시 강남구")
	require.NoError(t, err)
	require.NotNil(t, result)

	// CachedAt should be set automatically
	assert.False(t, result.CachedAt.IsZero())
	assert.WithinDuration(t, time.Now(), result.CachedAt, 2*time.Second)
}

func TestDelete(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()
	defer cache.Close()

	ctx := context.Background()
	region := "서울시 강남구"

	// Set data first
	data := &entity.WeatherData{
		Temperature:   25.5,
		Humidity:      60.0,
		Precipitation: 0.0,
		WindSpeed:     3.5,
		CachedAt:      time.Now(),
	}

	err := cache.Set(ctx, region, data)
	require.NoError(t, err)

	// Verify data exists
	result, err := cache.Get(ctx, region)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Delete data
	err = cache.Delete(ctx, region)
	require.NoError(t, err)

	// Verify data is deleted
	result, err = cache.Get(ctx, region)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestTTL(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()
	defer cache.Close()

	ctx := context.Background()
	region := "서울시 강남구"

	// Set data
	data := &entity.WeatherData{
		Temperature:   25.5,
		Humidity:      60.0,
		Precipitation: 0.0,
		WindSpeed:     3.5,
		CachedAt:      time.Now(),
	}

	err := cache.Set(ctx, region, data)
	require.NoError(t, err)

	// Check TTL
	ttl, err := cache.GetTTL(ctx, region)
	require.NoError(t, err)
	assert.Greater(t, ttl, time.Duration(0))
	assert.LessOrEqual(t, ttl, CacheTTL)
}

func TestTTL_Expiration(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()
	defer cache.Close()

	ctx := context.Background()
	region := "서울시 강남구"

	// Set data
	data := &entity.WeatherData{
		Temperature:   25.5,
		Humidity:      60.0,
		Precipitation: 0.0,
		WindSpeed:     3.5,
		CachedAt:      time.Now(),
	}

	err := cache.Set(ctx, region, data)
	require.NoError(t, err)

	// Fast forward time in miniredis
	mr.FastForward(31 * time.Minute)

	// Data should be expired
	result, err := cache.Get(ctx, region)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestPing(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()
	defer cache.Close()

	ctx := context.Background()

	err := cache.Ping(ctx)
	assert.NoError(t, err)
}

func TestClose(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()

	err := cache.Close()
	assert.NoError(t, err)

	// After close, operations should fail
	ctx := context.Background()
	err = cache.Ping(ctx)
	assert.Error(t, err)
}

func TestConcurrentAccess(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()
	defer cache.Close()

	ctx := context.Background()
	region := "서울시 강남구"

	// Test concurrent writes and reads
	done := make(chan bool)
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			data := &entity.WeatherData{
				Temperature:   float64(20 + idx),
				Humidity:      50.0,
				Precipitation: 0.0,
				WindSpeed:     2.0,
				CachedAt:      time.Now(),
			}

			// Write
			err := cache.Set(ctx, region, data)
			assert.NoError(t, err)

			// Read
			result, err := cache.Get(ctx, region)
			assert.NoError(t, err)
			assert.NotNil(t, result)

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestMultipleRegions(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()
	defer cache.Close()

	ctx := context.Background()

	regions := []string{
		"서울시 강남구",
		"서울시 서초구",
		"부산시 해운대구",
		"제주도 제주시",
	}

	// Set different data for each region
	for i, region := range regions {
		data := &entity.WeatherData{
			Temperature:   float64(20 + i),
			Humidity:      float64(50 + i*5),
			Precipitation: float64(i),
			WindSpeed:     float64(2 + i),
			CachedAt:      time.Now(),
		}

		err := cache.Set(ctx, region, data)
		require.NoError(t, err)
	}

	// Verify each region has correct data
	for i, region := range regions {
		result, err := cache.Get(ctx, region)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.InDelta(t, float64(20+i), result.Temperature, 0.01)
		assert.InDelta(t, float64(50+i*5), result.Humidity, 0.01)
	}
}

func TestUpdate_ExistingData(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()
	defer cache.Close()

	ctx := context.Background()
	region := "서울시 강남구"

	// Set initial data
	initialData := &entity.WeatherData{
		Temperature:   20.0,
		Humidity:      50.0,
		Precipitation: 0.0,
		WindSpeed:     2.0,
		CachedAt:      time.Now(),
	}

	err := cache.Set(ctx, region, initialData)
	require.NoError(t, err)

	// Update with new data
	updatedData := &entity.WeatherData{
		Temperature:   25.0,
		Humidity:      60.0,
		Precipitation: 5.0,
		WindSpeed:     4.0,
		CachedAt:      time.Now(),
	}

	err = cache.Set(ctx, region, updatedData)
	require.NoError(t, err)

	// Verify updated data
	result, err := cache.Get(ctx, region)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.InDelta(t, updatedData.Temperature, result.Temperature, 0.01)
	assert.InDelta(t, updatedData.Humidity, result.Humidity, 0.01)
	assert.InDelta(t, updatedData.Precipitation, result.Precipitation, 0.01)
	assert.InDelta(t, updatedData.WindSpeed, result.WindSpeed, 0.01)
}
