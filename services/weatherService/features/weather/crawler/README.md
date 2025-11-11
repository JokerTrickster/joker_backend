# Weather Crawler

A robust, thread-safe weather data crawler for Naver weather service with retry logic, structured logging, and comprehensive error handling.

## Features

- **HTTP Client with Timeout**: Configurable timeout (default 10 seconds)
- **Retry Logic**: Configurable retries with exponential backoff (1s, 2s, 4s)
- **Context-Aware**: Full support for context cancellation and timeouts
- **Thread-Safe**: Safe for concurrent use across multiple goroutines
- **Flexible HTML Parsing**: Multiple CSS selectors for resilient data extraction
- **Structured Logging**: zap logger for debugging and monitoring
- **Numeric Parsing**: Automatic conversion from strings ("15°") to float64 (15.0)
- **Dependency Injection**: Testable design with configurable HTTP client

## Installation

```bash
go get github.com/PuerkitoBio/goquery
go get go.uber.org/zap
```

## Usage

### Basic Usage

```go
import (
    "context"
    "fmt"
    "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/crawler"
)

func main() {
    ctx := context.Background()

    // Crawl weather data for Seoul
    data, err := crawler.CrawlWeather(ctx, "서울")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Temperature: %.1f°C\n", data.Temperature)
    fmt.Printf("Humidity: %.0f%%\n", data.Humidity)
    fmt.Printf("Precipitation: %.1fmm\n", data.Precipitation)
    fmt.Printf("Wind Speed: %.1fm/s\n", data.WindSpeed)
}
```

### With Custom Configuration

```go
// Create crawler with 15-second timeout and 5 retries
crawler := crawler.NewNaverWeatherCrawler(15*time.Second, 5)

ctx := context.Background()
data, err := crawler.Fetch(ctx, "부산")
if err != nil {
    log.Fatal(err)
}
```

### With Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

data, err := crawler.CrawlWeather(ctx, "부산")
if err != nil {
    log.Fatal(err)
}
```

## Data Structure

Uses the standard `entity.WeatherData` structure:

```go
type WeatherData struct {
    Temperature   float64   // Temperature in Celsius
    Humidity      float64   // Humidity percentage
    Precipitation float64   // Precipitation in mm
    WindSpeed     float64   // Wind speed in m/s
    CachedAt      time.Time // Time when data was cached
}
```

## Configuration

### Constants

- `naverWeatherBaseURL`: `https://search.naver.com/search.naver`
- `defaultTimeout`: 10 seconds
- `maxRetries`: 3 attempts
- `initialBackoff`: 1 second

### URL Pattern

Weather data is fetched from:
```
https://search.naver.com/search.naver?query=날씨+{region}
```

## Error Handling

The crawler implements comprehensive error handling:

1. **Empty Region**: Returns error if region parameter is empty
2. **Context Cancellation**: Respects context cancellation and timeouts
3. **HTTP Errors**: Handles non-200 status codes
4. **Parsing Errors**: Returns error if temperature cannot be extracted
5. **Retry Logic**: Automatically retries failed requests up to 3 times

## Testing

### Run Tests

```bash
go test ./features/weather/crawler/... -v -count=1
```

### Run Benchmarks

```bash
go test -bench=. -benchmem ./features/weather/crawler/...
```

### Test Coverage

- Success scenarios with mock servers
- Error handling (HTTP errors, parsing errors)
- Context cancellation and timeouts
- Retry logic with exponential backoff
- Thread safety with concurrent requests
- URL encoding for Korean characters
- Multiple CSS selector fallbacks

## Performance

Benchmark results on Intel i7-9750H @ 2.60GHz:

```
BenchmarkParseWeatherHTML-12    4723    222438 ns/op    492735 B/op    336 allocs/op
BenchmarkCrawlWeather-12        3       385594 ms/op   9119957 B/op   56157 allocs/op
```

Note: CrawlWeather benchmark performs actual network requests to Naver.

## Architecture

### Main Components

```go
type NaverWeatherCrawler struct {
    client  HTTPClient
    baseURL string
    logger  *zap.Logger
    timeout time.Duration
    retries int
}

func NewNaverWeatherCrawler(timeout time.Duration, retries int) *NaverWeatherCrawler
func (c *NaverWeatherCrawler) Fetch(ctx context.Context, region string) (*entity.WeatherData, error)
func (c *NaverWeatherCrawler) fetchWithRetry(ctx context.Context, region string, attempt int) (*entity.WeatherData, error)
func (c *NaverWeatherCrawler) parseWeatherData(doc *goquery.Document) (*entity.WeatherData, error)
```

### Dependency Injection

The crawler uses dependency injection for testability:

```go
type HTTPClient interface {
    Do(req *http.Request) (*http.Response, error)
}
```

This allows for:
- Easy mocking in tests
- Custom HTTP client configuration
- Testing without hitting real Naver servers

### HTML Parsing Strategy

Multiple CSS selectors are attempted in order:

1. `.temperature_text strong`
2. `.weather_graphic .temperature_text`
3. `.temperature_text`
4. `.temperature .num`

This ensures resilience against HTML structure changes.

### Numeric Parsing

The crawler automatically extracts numeric values from formatted strings:

```go
"15°" → 15.0
"60%" → 60.0
"2.5m/s" → 2.5
"0mm" → 0.0
"-5°" → -5.0
```

Uses regex pattern: `-?\d+\.?\d*` to extract numeric values including decimals and negatives.

## Implementation Details

### Retry Logic

```
Attempt 1: Immediate request
├─ Success → Return data
└─ Failure → Wait 1s

Attempt 2: After 1s backoff
├─ Success → Return data
└─ Failure → Wait 2s

Attempt 3: After 2s backoff
├─ Success → Return data
└─ Failure → Return error
```

Total retry time: ~3 seconds maximum

### Thread Safety

The crawler is designed to be thread-safe:
- No shared mutable state
- Each request creates its own HTTP request
- Safe for concurrent use from multiple goroutines

## CSS Selectors Used

Based on Naver weather page structure:

- **Temperature**: `.temperature_text strong`, `.weather_graphic .temperature_text`, `.temperature_text`, `.temperature .num`
- **Humidity**: Look for label containing "습도" in `.info_list .sort .term`
- **Precipitation**: Look for label containing "강수" in `.info_list .sort .term`
- **Wind Speed**: Look for label containing "바람" in `.info_list .sort .term`

## Error Handling

The crawler provides specific error types:

1. **Empty Region**: `region cannot be empty`
2. **HTTP Errors**: `unexpected status code: X`
3. **Parse Errors**: `failed to extract temperature from HTML`
4. **Numeric Parse**: `failed to parse temperature 'X'`
5. **Network Failures**: Automatic retry with exponential backoff
6. **Context Timeout**: Respects context deadline

## Files

- `naver.go` (297 lines): Main implementation with numeric parsing
- `naver_test.go` (566 lines): Comprehensive test suite (23 tests)
- `example_test.go` (62 lines): Usage examples
- `README.md`: Documentation

## Dependencies

- `github.com/PuerkitoBio/goquery`: HTML parsing and CSS selector support
- `go.uber.org/zap`: Structured logging
- `github.com/stretchr/testify`: Testing assertions

## License

Part of the joker_backend project.
