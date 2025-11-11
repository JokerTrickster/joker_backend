# Weather Cache

Redis-based caching layer for weather data with automatic TTL management and graceful fallback.

## Features

- **Connection Pooling**: Efficient Redis connection management with configurable pool size
- **Hash Storage**: Weather data stored as Redis hashes for efficient field access
- **30-Minute TTL**: Automatic expiration to ensure data freshness
- **Graceful Fallback**: Handle Redis unavailability without breaking the service
- **Structured Logging**: Integration with zap logger for observability
- **Context-Aware**: All operations support context for timeout and cancellation
- **Concurrent Safe**: Thread-safe operations with connection pooling

## Architecture

```
┌─────────────────┐
│  Weather API    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     Cache Hit     ┌─────────────────┐
│ WeatherCache    │◄──────────────────│  Redis          │
│  (Get/Set)      │                    │  Hash Store     │
└────────┬────────┘     Cache Miss     └─────────────────┘
         │                              TTL: 30 minutes
         ▼
┌─────────────────┐
│  Weather        │
│  Crawler        │
└─────────────────┘
```

## Installation

```bash
cd /Users/luxrobo/project/joker_backend/services/weatherService
go get github.com/redis/go-redis/v9
go get github.com/alicebob/miniredis/v2  # for testing
```

## Usage

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/cache"
    "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewProduction()

    // Initialize cache
    weatherCache, err := cache.NewWeatherCache("localhost:6379", "", logger)
    if err != nil {
        log.Fatalf("Failed to connect to Redis: %v", err)
    }
    defer weatherCache.Close()

    ctx := context.Background()
    region := "서울시 강남구"

    // Try to get from cache
    cachedData, err := weatherCache.Get(ctx, region)
    if err != nil {
        log.Printf("Cache error: %v", err)
    }

    if cachedData != nil {
        log.Printf("Cache hit: Temperature=%.1f°C", cachedData.Temperature)
    } else {
        // Fetch from crawler and cache
        freshData := fetchFromCrawler(region) // implement this

        if err := weatherCache.Set(ctx, region, freshData); err != nil {
            log.Printf("Failed to cache: %v", err)
        }
    }
}
```

### With Graceful Fallback

```go
func getWeatherData(ctx context.Context, region string) (*entity.WeatherData, error) {
    // Try cache first
    if weatherCache != nil {
        if data, err := weatherCache.Get(ctx, region); err == nil && data != nil {
            return data, nil
        }
    }

    // Fallback to direct crawl
    data, err := crawlWeatherData(region)
    if err != nil {
        return nil, err
    }

    // Try to cache (ignore errors)
    if weatherCache != nil {
        _ = weatherCache.Set(ctx, region, data)
    }

    return data, nil
}
```

### Integration with Use Case

```go
type WeatherUseCase struct {
    cache   _interface.IWeatherCache
    crawler WeatherCrawler
    logger  *zap.Logger
}

func (uc *WeatherUseCase) GetWeather(ctx context.Context, region string) (*entity.WeatherData, error) {
    // 1. Try cache first
    if uc.cache != nil {
        data, err := uc.cache.Get(ctx, region)
        if err == nil && data != nil {
            uc.logger.Debug("Cache hit", zap.String("region", region))
            return data, nil
        }
        uc.logger.Debug("Cache miss", zap.String("region", region))
    }

    // 2. Fetch from crawler
    data, err := uc.crawler.Fetch(ctx, region)
    if err != nil {
        return nil, fmt.Errorf("crawler failed: %w", err)
    }

    // 3. Store in cache for next time
    if uc.cache != nil {
        if err := uc.cache.Set(ctx, region, data); err != nil {
            uc.logger.Warn("Failed to cache data",
                zap.String("region", region),
                zap.Error(err))
        }
    }

    return data, nil
}
```

## Redis Key Format

Keys follow the pattern: `weather:도:시:구`

**Examples:**
- `weather:서울시:강남구` - Seoul Gangnam-gu
- `weather:부산시:해운대구` - Busan Haeundae-gu
- `weather:제주도:제주시` - Jeju Island Jeju City
- `weather:경기도:성남시:분당구` - Gyeonggi-do Seongnam-si Bundang-gu

**Region parsing:**
- Supports space, comma, dash, and slash delimiters
- Automatically normalizes extra spaces
- Example: `"서울시, 강남구"` → `"weather:서울시:강남구"`

## Data Structure

Weather data is stored as Redis Hash with these fields:

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `temperature` | string (float) | Temperature in Celsius | "25.50" |
| `humidity` | string (float) | Humidity percentage | "60.00" |
| `precipitation` | string (float) | Precipitation in mm | "0.00" |
| `wind_speed` | string (float) | Wind speed in m/s | "3.50" |
| `cached_at` | string (unix timestamp) | Cache timestamp | "1699574823" |

**TTL:** 30 minutes (1800 seconds)

## Configuration

### Connection Options

```go
opt := &redis.Options{
    Addr:         "localhost:6379",
    Password:     "",          // Redis password
    DB:           0,           // Database number
    PoolSize:     10,          // Max connections
    MinIdleConns: 5,           // Min idle connections
    MaxRetries:   3,           // Retry failed commands
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
}
```

### Environment Variables

```bash
# For weatherService
REDIS_ADDR="localhost:6379"
REDIS_PASSWORD=""
REDIS_DB="0"
```

## Testing

### Run Tests

```bash
cd /Users/luxrobo/project/joker_backend/services/weatherService
go test ./features/weather/cache/... -v -count=1
```

### Test Coverage

```bash
go test ./features/weather/cache/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Key Test Scenarios

- Connection pooling and retry logic
- Key generation with various region formats
- Set/Get operations with data validation
- Cache miss handling
- TTL expiration behavior
- Nil data validation
- Concurrent access safety
- Multiple regions isolation
- Update existing data
- Delete operations

## Error Handling

### Redis Connection Errors

```go
cache, err := cache.NewWeatherCache(addr, password, logger)
if err != nil {
    // Handle connection failure
    // Option 1: Fail fast (critical services)
    return fmt.Errorf("redis required: %w", err)

    // Option 2: Graceful degradation (recommended)
    logger.Warn("Redis unavailable, continuing without cache", zap.Error(err))
    cache = nil  // Proceed with nil cache
}
```

### Cache Operation Errors

```go
data, err := cache.Get(ctx, region)
if err != nil {
    // Log and fallback to crawler
    logger.Warn("Cache get failed", zap.Error(err))
    // Continue with crawler fetch
}

if err := cache.Set(ctx, region, data); err != nil {
    // Non-critical: log and continue
    logger.Warn("Cache set failed", zap.Error(err))
    // Data is already available, cache failure is not critical
}
```

## Performance Considerations

### Connection Pooling

- Pool size: 10 connections (adjust based on concurrency)
- Min idle: 5 connections (warm pool for low latency)
- Retry: 3 attempts (handle transient failures)

### TTL Strategy

- 30 minutes: Balance between freshness and crawler load
- Consider: Weather data doesn't change frequently
- Trade-off: Longer TTL = less accurate but fewer crawler requests

### Memory Usage

Per cached entry (approximate):
- Key: 50 bytes (average)
- Hash fields: 150 bytes (5 fields × 30 bytes)
- Total: ~200 bytes per region

For 1000 regions: ~200KB memory footprint

## Monitoring

### Key Metrics

1. **Cache Hit Rate**: `hits / (hits + misses)`
2. **Cache Latency**: P50, P95, P99 for Get/Set operations
3. **Error Rate**: Failed operations / total operations
4. **TTL Distribution**: Remaining TTL when accessed
5. **Pool Utilization**: Active connections / pool size

### Logging

```go
// Debug level
cache.Get() → "Cache hit" / "Cache miss"
cache.Set() → "Successfully cached weather data"
cache.Delete() → "Successfully deleted cached weather data"

// Error level
cache.Get() → "Failed to get weather data from cache"
cache.Set() → "Failed to set weather data in cache"
```

## Future Enhancements

1. **Cache Warming**: Pre-populate cache for popular regions
2. **Stale-While-Revalidate**: Serve stale data while fetching fresh data
3. **Multi-Region Support**: Cache multiple regions in parallel
4. **Compression**: Compress large weather payloads
5. **Metrics**: Export cache metrics to monitoring system
6. **Circuit Breaker**: Prevent Redis overload during outages

## Related Files

- `/Users/luxrobo/project/joker_backend/services/weatherService/features/weather/cache/weather.go` - Implementation
- `/Users/luxrobo/project/joker_backend/services/weatherService/features/weather/cache/weather_test.go` - Tests
- `/Users/luxrobo/project/joker_backend/services/weatherService/features/weather/model/entity/weatherData.go` - Data model
- `/Users/luxrobo/project/joker_backend/services/weatherService/features/weather/model/interface/IWeatherCache.go` - Interface

## References

- Redis Go Client: https://github.com/redis/go-redis
- Miniredis (Testing): https://github.com/alicebob/miniredis
- Zap Logger: https://github.com/uber-go/zap
