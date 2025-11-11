package cache

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	// WeatherKeyPrefix is the prefix for weather cache keys
	WeatherKeyPrefix = "weather"
	// CacheTTL is the cache expiration time (30 minutes)
	CacheTTL = 30 * time.Minute
)

// WeatherCache manages Redis caching for weather data
type WeatherCache struct {
	client *redis.Client
	logger *zap.Logger
}

// NewWeatherCache creates a new WeatherCache instance with connection pooling
func NewWeatherCache(redisAddr, password string, logger *zap.Logger) (*WeatherCache, error) {
	if logger == nil {
		var err error
		logger, err = zap.NewProduction()
		if err != nil {
			return nil, fmt.Errorf("failed to create logger: %w", err)
		}
	}

	opt := &redis.Options{
		Addr:         redisAddr,
		Password:     password,
		DB:           0,
		PoolSize:     10,              // Connection pool size
		MinIdleConns: 5,               // Minimum idle connections
		MaxRetries:   3,               // Retry failed commands
		DialTimeout:  5 * time.Second, // Connection timeout
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Successfully connected to Redis for weather cache",
		zap.String("address", redisAddr))

	return &WeatherCache{
		client: client,
		logger: logger,
	}, nil
}

// generateKey generates a Redis key from region string
// Format: weather:도:시:구
// Example: weather:서울시:강남구
func (c *WeatherCache) generateKey(region string) string {
	// Remove extra spaces and normalize
	region = strings.TrimSpace(region)
	region = strings.ReplaceAll(region, "  ", " ")

	// Split by space or other common delimiters
	parts := strings.FieldsFunc(region, func(r rune) bool {
		return r == ' ' || r == ',' || r == '/' || r == '-'
	})

	// Join with colon
	key := fmt.Sprintf("%s:%s", WeatherKeyPrefix, strings.Join(parts, ":"))

	return key
}

// Get retrieves weather data from Redis cache
func (c *WeatherCache) Get(ctx context.Context, region string) (*entity.WeatherData, error) {
	key := c.generateKey(region)

	// Get all hash fields
	result, err := c.client.HGetAll(ctx, key).Result()
	if err != nil {
		c.logger.Error("Failed to get weather data from cache",
			zap.String("key", key),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}

	// Check if data exists
	if len(result) == 0 {
		c.logger.Debug("Cache miss",
			zap.String("key", key))
		return nil, nil // Cache miss
	}

	// Parse cached data
	data := &entity.WeatherData{}

	if temp, ok := result["temperature"]; ok {
		data.Temperature, _ = strconv.ParseFloat(temp, 64)
	}
	if humidity, ok := result["humidity"]; ok {
		data.Humidity, _ = strconv.ParseFloat(humidity, 64)
	}
	if precip, ok := result["precipitation"]; ok {
		data.Precipitation, _ = strconv.ParseFloat(precip, 64)
	}
	if wind, ok := result["wind_speed"]; ok {
		data.WindSpeed, _ = strconv.ParseFloat(wind, 64)
	}
	if cachedAt, ok := result["cached_at"]; ok {
		if timestamp, err := strconv.ParseInt(cachedAt, 10, 64); err == nil {
			data.CachedAt = time.Unix(timestamp, 0)
		}
	}

	c.logger.Debug("Cache hit",
		zap.String("key", key),
		zap.Time("cached_at", data.CachedAt))

	return data, nil
}

// Set stores weather data in Redis cache with TTL
func (c *WeatherCache) Set(ctx context.Context, region string, data *entity.WeatherData) error {
	if data == nil {
		return fmt.Errorf("weather data cannot be nil")
	}

	key := c.generateKey(region)

	// Set cached_at timestamp if not already set
	if data.CachedAt.IsZero() {
		data.CachedAt = time.Now()
	}

	// Prepare hash fields
	fields := map[string]interface{}{
		"temperature":   fmt.Sprintf("%.2f", data.Temperature),
		"humidity":      fmt.Sprintf("%.2f", data.Humidity),
		"precipitation": fmt.Sprintf("%.2f", data.Precipitation),
		"wind_speed":    fmt.Sprintf("%.2f", data.WindSpeed),
		"cached_at":     strconv.FormatInt(data.CachedAt.Unix(), 10),
	}

	// Use pipeline for atomic operations
	pipe := c.client.Pipeline()
	pipe.HSet(ctx, key, fields)
	pipe.Expire(ctx, key, CacheTTL)

	_, err := pipe.Exec(ctx)
	if err != nil {
		c.logger.Error("Failed to set weather data in cache",
			zap.String("key", key),
			zap.Error(err))
		return fmt.Errorf("failed to set cache: %w", err)
	}

	c.logger.Debug("Successfully cached weather data",
		zap.String("key", key),
		zap.Duration("ttl", CacheTTL))

	return nil
}

// Delete removes weather data from cache
func (c *WeatherCache) Delete(ctx context.Context, region string) error {
	key := c.generateKey(region)

	err := c.client.Del(ctx, key).Err()
	if err != nil {
		c.logger.Error("Failed to delete weather data from cache",
			zap.String("key", key),
			zap.Error(err))
		return fmt.Errorf("failed to delete cache: %w", err)
	}

	c.logger.Debug("Successfully deleted cached weather data",
		zap.String("key", key))

	return nil
}

// Close closes the Redis connection
func (c *WeatherCache) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// Ping checks if Redis connection is alive
func (c *WeatherCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// GetTTL returns the remaining TTL for a cached entry
func (c *WeatherCache) GetTTL(ctx context.Context, region string) (time.Duration, error) {
	key := c.generateKey(region)
	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}
	return ttl, nil
}
