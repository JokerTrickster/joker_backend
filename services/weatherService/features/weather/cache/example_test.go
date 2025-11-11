package cache_test

import (
	"context"
	"fmt"
	"time"

	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/cache"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
	"go.uber.org/zap"
)

// ExampleWeatherCache demonstrates basic usage of the weather cache
func ExampleWeatherCache() {
	logger, _ := zap.NewDevelopment()

	// Initialize cache
	weatherCache, err := cache.NewWeatherCache("localhost:6379", "", logger)
	if err != nil {
		fmt.Printf("Failed to connect to Redis: %v\n", err)
		return
	}
	defer weatherCache.Close()

	ctx := context.Background()
	region := "서울시 강남구"

	// Check if data is cached
	cachedData, err := weatherCache.Get(ctx, region)
	if err != nil {
		fmt.Printf("Error getting cache: %v\n", err)
		return
	}

	if cachedData != nil {
		// Cache hit - use cached data
		fmt.Printf("Using cached data: Temperature=%.1f°C\n", cachedData.Temperature)
	} else {
		// Cache miss - fetch from crawler and cache it
		weatherData := &entity.WeatherData{
			Temperature:   25.5,
			Humidity:      60.0,
			Precipitation: 0.0,
			WindSpeed:     3.5,
			CachedAt:      time.Now(),
		}

		// Store in cache
		if err := weatherCache.Set(ctx, region, weatherData); err != nil {
			fmt.Printf("Error setting cache: %v\n", err)
			return
		}

		fmt.Printf("Cached new data: Temperature=%.1f°C\n", weatherData.Temperature)
	}
}

// ExampleWeatherCache_withFallback demonstrates cache with fallback to crawler
func ExampleWeatherCache_withFallback() {
	logger, _ := zap.NewDevelopment()

	weatherCache, err := cache.NewWeatherCache("localhost:6379", "", logger)
	if err != nil {
		// Fallback: If Redis is unavailable, proceed without cache
		fmt.Println("Redis unavailable, using direct crawl")
		return
	}
	defer weatherCache.Close()

	ctx := context.Background()
	region := "부산시 해운대구"

	// Try to get from cache first
	cachedData, err := weatherCache.Get(ctx, region)
	if err != nil || cachedData == nil {
		// Fallback to crawler
		fmt.Println("Cache miss or error, fetching from crawler")

		// Simulate crawler fetch
		freshData := &entity.WeatherData{
			Temperature:   18.3,
			Humidity:      85.5,
			Precipitation: 12.5,
			WindSpeed:     8.2,
			CachedAt:      time.Now(),
		}

		// Try to cache the fresh data (ignore errors)
		_ = weatherCache.Set(ctx, region, freshData)

		fmt.Printf("Using fresh data: Temperature=%.1f°C\n", freshData.Temperature)
	} else {
		fmt.Printf("Using cached data: Temperature=%.1f°C\n", cachedData.Temperature)
	}
}

// ExampleWeatherCache_ttl demonstrates TTL checking
func ExampleWeatherCache_ttl() {
	logger, _ := zap.NewDevelopment()

	weatherCache, err := cache.NewWeatherCache("localhost:6379", "", logger)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}
	defer weatherCache.Close()

	ctx := context.Background()
	region := "제주도 제주시"

	// Cache some data
	weatherData := &entity.WeatherData{
		Temperature:   22.0,
		Humidity:      70.0,
		Precipitation: 5.0,
		WindSpeed:     10.5,
	}

	weatherCache.Set(ctx, region, weatherData)

	// Check remaining TTL
	ttl, err := weatherCache.GetTTL(ctx, region)
	if err != nil {
		fmt.Printf("Error getting TTL: %v\n", err)
		return
	}

	fmt.Printf("Cache will expire in: %v\n", ttl.Round(time.Minute))
}
