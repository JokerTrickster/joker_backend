# Naver Weather Crawler Usage Example

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/crawler"
)

func main() {
    // Simple usage with default settings
    ctx := context.Background()

    data, err := crawler.CrawlWeather(ctx, "서울")
    if err != nil {
        log.Fatalf("Failed to fetch weather: %v", err)
    }

    fmt.Printf("Seoul Weather:\n")
    fmt.Printf("  Temperature: %.1f°C\n", data.Temperature)
    fmt.Printf("  Humidity: %.0f%%\n", data.Humidity)
    fmt.Printf("  Precipitation: %.1fmm\n", data.Precipitation)
    fmt.Printf("  Wind Speed: %.1fm/s\n", data.WindSpeed)
    fmt.Printf("  Cached At: %v\n", data.CachedAt.Format(time.RFC3339))
}
```

## Advanced Usage

### Custom Configuration

```go
// Create crawler with custom timeout and retries
crawler := crawler.NewNaverWeatherCrawler(
    15*time.Second, // timeout
    5,              // max retries
)

ctx := context.Background()
data, err := crawler.Fetch(ctx, "부산")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Temperature: %.1f°C\n", data.Temperature)
```

### With Context Timeout

```go
// Create context with 5-second timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

data, err := crawler.CrawlWeather(ctx, "인천")
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        log.Println("Request timed out after 5 seconds")
    } else {
        log.Printf("Error: %v", err)
    }
    return
}

fmt.Printf("Incheon Temperature: %.1f°C\n", data.Temperature)
```

### Concurrent Fetching for Multiple Regions

```go
package main

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/crawler"
    "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
)

func fetchMultipleRegions() {
    regions := []string{"서울", "부산", "인천", "대구", "광주"}

    type result struct {
        region string
        data   *entity.WeatherData
        err    error
    }

    results := make(chan result, len(regions))
    var wg sync.WaitGroup

    crawler := crawler.NewNaverWeatherCrawler(10*time.Second, 3)
    ctx := context.Background()

    // Launch concurrent fetches
    for _, region := range regions {
        wg.Add(1)
        go func(r string) {
            defer wg.Done()
            data, err := crawler.Fetch(ctx, r)
            results <- result{region: r, data: data, err: err}
        }(region)
    }

    // Wait and close
    go func() {
        wg.Wait()
        close(results)
    }()

    // Process results
    fmt.Println("Weather Data for Multiple Regions:")
    fmt.Println("=" * 50)

    for r := range results {
        if r.err != nil {
            fmt.Printf("%s: Error - %v\n", r.region, r.err)
            continue
        }
        fmt.Printf("%s: %.1f°C, %.0f%% humidity, %.1fm/s wind\n",
            r.region,
            r.data.Temperature,
            r.data.Humidity,
            r.data.WindSpeed,
        )
    }
}
```

### Error Handling

```go
data, err := crawler.CrawlWeather(ctx, "서울")
if err != nil {
    switch {
    case strings.Contains(err.Error(), "region cannot be empty"):
        log.Println("Invalid input: region is required")
    case strings.Contains(err.Error(), "failed after"):
        log.Println("All retry attempts failed")
    case strings.Contains(err.Error(), "unexpected status code"):
        log.Println("HTTP error from Naver")
    case strings.Contains(err.Error(), "failed to extract temperature"):
        log.Println("HTML parsing failed - Naver page structure may have changed")
    case ctx.Err() == context.DeadlineExceeded:
        log.Println("Request timed out")
    default:
        log.Printf("Unknown error: %v", err)
    }
    return
}

// Use data
fmt.Printf("Temperature: %.1f°C\n", data.Temperature)
```

### Integration with Cache

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/cache"
    "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/crawler"
)

func getWeatherWithCache(region string) (*entity.WeatherData, error) {
    ctx := context.Background()

    // Initialize cache and crawler
    weatherCache := cache.NewRedisWeatherCache(redisClient, 10*time.Minute)
    weatherCrawler := crawler.NewNaverWeatherCrawler(10*time.Second, 3)

    // Try cache first
    cachedData, err := weatherCache.Get(ctx, region)
    if err == nil && cachedData != nil {
        fmt.Println("Cache hit!")
        return cachedData, nil
    }

    // Cache miss - fetch from Naver
    fmt.Println("Cache miss - fetching from Naver...")
    data, err := weatherCrawler.Fetch(ctx, region)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch weather: %w", err)
    }

    // Store in cache
    if err := weatherCache.Set(ctx, region, data); err != nil {
        // Log warning but don't fail
        fmt.Printf("Warning: failed to cache data: %v\n", err)
    }

    return data, nil
}
```

## Expected Output

```
Seoul Weather:
  Temperature: 15.0°C
  Humidity: 60%
  Precipitation: 0.0mm
  Wind Speed: 2.5m/s
  Cached At: 2025-11-11T15:30:45+09:00
```

## Common Issues

### Issue 1: Empty Temperature
**Problem**: `failed to extract temperature from HTML`
**Cause**: Naver changed their HTML structure
**Solution**: Update CSS selectors in `parseWeatherData()`

### Issue 2: Timeout Errors
**Problem**: Requests consistently timing out
**Cause**: Network issues or slow response from Naver
**Solution**: Increase timeout or check network connectivity

```go
crawler := crawler.NewNaverWeatherCrawler(20*time.Second, 5)
```

### Issue 3: Parse Errors
**Problem**: `failed to parse temperature '데이터없음'`
**Cause**: Naver returning non-numeric data
**Solution**: Check if region name is valid

## Performance Considerations

1. **Caching**: Always use caching to avoid hammering Naver servers
2. **Concurrent Requests**: Crawler is thread-safe, use goroutines for multiple regions
3. **Timeout**: Set appropriate timeout based on your latency requirements
4. **Retries**: 3 retries is usually sufficient, increase for unreliable networks

## Best Practices

1. Always use context with timeout
2. Implement caching layer (Redis recommended)
3. Log all errors with structured logging
4. Handle partial data gracefully
5. Monitor fetch success rates
6. Respect Naver's rate limits
7. Set User-Agent header (already done internally)

## Testing

Run tests:
```bash
go test ./features/weather/crawler/... -v -count=1
```

Run benchmarks:
```bash
go test -bench=. -benchmem ./features/weather/crawler/...
```

Check coverage:
```bash
go test -cover ./features/weather/crawler/...
```
