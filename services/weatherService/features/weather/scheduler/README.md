# Weather Scheduler Service

Production-ready scheduler service for managing periodic weather alarm notifications.

## Quick Start

```go
import (
    "context"
    "time"
    "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/scheduler"
)

// Initialize scheduler
scheduler := scheduler.NewWeatherSchedulerService(
    repository,  // ISchedulerWeatherRepository
    crawler,     // IWeatherCrawler
    cache,       // IWeatherCache
    notifier,    // IFCMNotifier
    logger,      // *zap.Logger
    1*time.Minute,  // Interval
)

// Start in background
go scheduler.Start(context.Background())

// Graceful shutdown
scheduler.Stop()
```

## Features

- **1-Minute Ticker**: Processes alarms every minute
- **Graceful Shutdown**: 30-second timeout for in-flight operations
- **Cache-First**: Minimizes crawler requests with Redis cache
- **Concurrent Processing**: Parallel alarm processing
- **Error Resilience**: Continues on individual alarm failures
- **Thread-Safe**: Safe concurrent Start/Stop calls
- **Structured Logging**: Full observability with zap

## Architecture

```
Scheduler Tick (every 1 minute)
    ↓
GetAlarmsToNotify(target_time)
    ↓
For each alarm:
    ├─ Get weather data (cache → crawler)
    ├─ Get FCM tokens
    ├─ Send notification
    └─ Update last_sent
```

## Error Handling

| Error Type | Behavior | Recovery |
|------------|----------|----------|
| Crawler failure | Skip alarm | Retry next day |
| Cache failure | Fallback to crawler | Continue processing |
| FCM failure | Log & update last_sent | Prevent retry storm |
| Database failure | Log & skip alarm | Continue to next alarm |

## Testing

```bash
# Run all tests
go test ./scheduler/...

# Run specific test
go test ./scheduler/... -run TestProcessAlarms_Success

# Run with coverage
go test ./scheduler/... -cover
```

## Files

- `scheduler.go` - Core scheduler implementation
- `scheduler_test.go` - Comprehensive unit tests (17 test cases)
- `example_integration.go` - Integration examples
- `IMPLEMENTATION_SUMMARY.md` - Detailed technical documentation
- `README.md` - This file

## Dependencies

### Required
- `gorm.io/gorm` - Database ORM
- `github.com/redis/go-redis/v9` - Redis cache
- `go.uber.org/zap` - Structured logging

### Test Dependencies
- `github.com/stretchr/testify/mock` - Mocking framework
- `github.com/stretchr/testify/assert` - Test assertions

## Configuration

### Environment Variables
```bash
SCHEDULER_INTERVAL=1m      # Default: 1 minute
DB_HOST=localhost
DB_PORT=3306
REDIS_ADDR=localhost:6379
LOG_LEVEL=info
```

### Production Settings
```go
scheduler := scheduler.NewWeatherSchedulerService(
    repo,
    crawler,
    cache,
    notifier,
    logger,
    1*time.Minute,  // Production interval
)
```

## Deployment

### Docker Compose
See `example_integration.go` for full docker-compose.yaml

### Kubernetes
See `example_integration.go` for deployment manifest

### Important: Single Instance
Only run **one** scheduler instance per environment to prevent duplicate notifications.

## Monitoring

### Recommended Metrics
- `scheduler_alarms_processed_total{status="success|failure"}`
- `scheduler_processing_duration_seconds`
- `scheduler_cache_hit_ratio`
- `scheduler_active_alarms_gauge`

### Health Check
```go
func (s *WeatherSchedulerService) IsHealthy() bool {
    return s.running
}
```

## Next Steps

1. Implement FCM notifier (Task #7)
2. Add Prometheus metrics
3. Implement distributed locking for multi-instance support
4. Add circuit breaker for crawler protection

## License

Copyright (c) 2025 Joker Backend Team
