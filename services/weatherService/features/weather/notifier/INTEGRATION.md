# FCM Notifier Integration Guide

## Overview

This guide demonstrates how to integrate the FCM notifier with the weather data collector scheduler.

## Architecture

```
Scheduler (Cron/Timer)
    ↓
GetAlarmsToNotify (Repository)
    ↓
For each alarm:
    ├─ GetFCMTokens (Repository)
    ├─ Get Weather Data (Cache/Crawler)
    ├─ SendWeatherNotification (FCM Notifier)
    └─ UpdateLastSent (Repository)
```

## Complete Integration Example

```go
package scheduler

import (
    "context"
    "os"
    "time"

    "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/cache"
    "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/notifier"
    "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/repository"
    "go.uber.org/zap"
    "gorm.io/gorm"
)

type WeatherNotificationScheduler struct {
    schedulerRepo   *repository.SchedulerWeatherRepository
    weatherCache    *cache.WeatherCache
    fcmNotifier     *notifier.FCMNotifier
    logger          *zap.Logger
}

func NewWeatherNotificationScheduler(db *gorm.DB, logger *zap.Logger) (*WeatherNotificationScheduler, error) {
    // Initialize repository
    schedulerRepo := repository.NewSchedulerWeatherRepository(db).(*repository.SchedulerWeatherRepository)

    // Initialize weather cache (assuming Redis)
    weatherCache, err := cache.NewWeatherCache("localhost:6379", "", 0, 24*time.Hour, logger)
    if err != nil {
        return nil, err
    }

    // Initialize FCM notifier
    credentialsPath := os.Getenv("FCM_CREDENTIALS_PATH")
    if credentialsPath == "" {
        credentialsPath = "/etc/firebase/credentials.json"
    }

    fcmNotifier, err := notifier.NewFCMNotifier(credentialsPath, logger)
    if err != nil {
        return nil, err
    }

    return &WeatherNotificationScheduler{
        schedulerRepo: schedulerRepo,
        weatherCache:  weatherCache,
        fcmNotifier:   fcmNotifier,
        logger:        logger,
    }, nil
}

// ProcessWeatherAlarms processes all alarms scheduled for the target time
// Typically called every minute by a cron job
func (s *WeatherNotificationScheduler) ProcessWeatherAlarms(ctx context.Context, targetTime time.Time) error {
    s.logger.Info("Starting weather alarm processing",
        zap.Time("target_time", targetTime),
    )

    // Get all alarms scheduled for this time
    alarms, err := s.schedulerRepo.GetAlarmsToNotify(ctx, targetTime)
    if err != nil {
        s.logger.Error("Failed to get alarms to notify", zap.Error(err))
        return err
    }

    if len(alarms) == 0 {
        s.logger.Info("No alarms to process")
        return nil
    }

    s.logger.Info("Processing alarms",
        zap.Int("alarm_count", len(alarms)),
    )

    successCount := 0
    failureCount := 0

    // Process each alarm
    for _, alarm := range alarms {
        err := s.processAlarm(ctx, alarm)
        if err != nil {
            failureCount++
            s.logger.Error("Failed to process alarm",
                zap.Int("alarm_id", alarm.ID),
                zap.Int("user_id", alarm.UserID),
                zap.String("region", alarm.Region),
                zap.Error(err),
            )
            continue
        }
        successCount++
    }

    s.logger.Info("Alarm processing completed",
        zap.Int("success", successCount),
        zap.Int("failure", failureCount),
    )

    return nil
}

// processAlarm processes a single alarm
func (s *WeatherNotificationScheduler) processAlarm(ctx context.Context, alarm entity.UserAlarm) error {
    // Get FCM tokens for user
    tokenEntities, err := s.schedulerRepo.GetFCMTokens(ctx, alarm.UserID)
    if err != nil {
        return fmt.Errorf("failed to get FCM tokens: %w", err)
    }

    if len(tokenEntities) == 0 {
        s.logger.Warn("No FCM tokens found for user",
            zap.Int("user_id", alarm.UserID),
        )
        // Still update last_sent to prevent retry
        return s.schedulerRepo.UpdateLastSent(ctx, alarm.ID, time.Now())
    }

    // Extract token strings
    tokens := make([]string, len(tokenEntities))
    for i, t := range tokenEntities {
        tokens[i] = t.FCMToken
    }

    // Get weather data from cache
    weatherData, err := s.weatherCache.Get(ctx, alarm.Region)
    if err != nil {
        return fmt.Errorf("failed to get weather data: %w", err)
    }

    // Send notifications
    err = s.fcmNotifier.SendWeatherNotification(ctx, tokens, weatherData, alarm.Region)
    if err != nil {
        return fmt.Errorf("failed to send notifications: %w", err)
    }

    // Update last sent timestamp
    err = s.schedulerRepo.UpdateLastSent(ctx, alarm.ID, time.Now())
    if err != nil {
        s.logger.Error("Failed to update last_sent, notification may be duplicated",
            zap.Int("alarm_id", alarm.ID),
            zap.Error(err),
        )
        // Don't return error since notification was successful
    }

    s.logger.Info("Alarm processed successfully",
        zap.Int("alarm_id", alarm.ID),
        zap.Int("user_id", alarm.UserID),
        zap.String("region", alarm.Region),
        zap.Int("tokens_sent", len(tokens)),
    )

    return nil
}

// Start begins the scheduler (example using time.Ticker)
func (s *WeatherNotificationScheduler) Start(ctx context.Context) {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()

    s.logger.Info("Weather notification scheduler started")

    for {
        select {
        case <-ctx.Done():
            s.logger.Info("Scheduler stopped")
            return
        case t := <-ticker.C:
            // Round to next minute
            targetTime := t.Add(1 * time.Minute).Truncate(1 * time.Minute)

            // Process alarms with timeout
            processCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
            err := s.ProcessWeatherAlarms(processCtx, targetTime)
            cancel()

            if err != nil {
                s.logger.Error("Error processing weather alarms",
                    zap.Error(err),
                )
            }
        }
    }
}
```

## Usage in Main Application

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"

    "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/scheduler"
    "go.uber.org/zap"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

func main() {
    // Initialize logger
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // Initialize database
    dsn := "user:password@tcp(localhost:3306)/weather_db?parseTime=true"
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        logger.Fatal("Failed to connect to database", zap.Error(err))
    }

    // Create scheduler
    weatherScheduler, err := scheduler.NewWeatherNotificationScheduler(db, logger)
    if err != nil {
        logger.Fatal("Failed to create scheduler", zap.Error(err))
    }

    // Setup graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Handle signals
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    // Start scheduler in goroutine
    go weatherScheduler.Start(ctx)

    // Wait for signal
    <-sigChan
    logger.Info("Shutdown signal received")

    // Cancel context to stop scheduler
    cancel()

    logger.Info("Application stopped")
}
```

## Environment Configuration

Create a `.env` file:

```bash
# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=weather_user
DB_PASSWORD=secret
DB_NAME=weather_db

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Firebase
FCM_CREDENTIALS_PATH=/etc/firebase/credentials.json

# Logging
LOG_LEVEL=info
```

## Cron Job Alternative

If using cron instead of a long-running process:

```go
package main

import (
    "context"
    "time"

    // ... imports
)

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        logger.Fatal("Failed to connect to database", zap.Error(err))
    }

    scheduler, err := scheduler.NewWeatherNotificationScheduler(db, logger)
    if err != nil {
        logger.Fatal("Failed to create scheduler", zap.Error(err))
    }

    // Process for current minute + 1
    targetTime := time.Now().Add(1 * time.Minute).Truncate(1 * time.Minute)

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    err = scheduler.ProcessWeatherAlarms(ctx, targetTime)
    if err != nil {
        logger.Fatal("Failed to process alarms", zap.Error(err))
    }
}
```

Crontab entry:
```cron
* * * * * /usr/local/bin/weather-notifier >> /var/log/weather-notifier.log 2>&1
```

## Error Handling Strategy

### Transient Errors (Retry)
- Network timeouts
- Temporary FCM service unavailability
- Redis connection issues

### Permanent Errors (Skip and Log)
- Invalid FCM tokens
- Malformed weather data
- Missing regions

### Critical Errors (Alert)
- Database connection loss
- All notifications failing
- Firebase authentication failure

## Monitoring and Metrics

Key metrics to track:

1. **Notification Success Rate**: `(successful_sends / total_attempts) * 100`
2. **Processing Time**: Time taken to process all alarms
3. **Token Validity**: Percentage of valid vs invalid tokens
4. **Cache Hit Rate**: Weather data cache effectiveness
5. **Alarm Count**: Number of alarms processed per interval

## Testing

Integration test example:

```go
func TestSchedulerIntegration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer teardownTestDB(t, db)

    // Create mock FCM client
    mockFCM := &notifier.MockFCMClient{
        SendMulticastFunc: func(ctx context.Context, msg *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
            return &messaging.BatchResponse{
                SuccessCount: len(msg.Tokens),
            }, nil
        },
    }

    // Create scheduler with mock
    scheduler := &WeatherNotificationScheduler{
        schedulerRepo: repository.NewSchedulerWeatherRepository(db),
        weatherCache:  setupTestCache(t),
        fcmNotifier:   notifier.NewFCMNotifierWithClient(mockFCM, zap.NewNop()),
        logger:        zap.NewNop(),
    }

    // Create test data
    createTestAlarm(t, db)
    createTestTokens(t, db)
    createTestWeatherData(t)

    // Process alarms
    err := scheduler.ProcessWeatherAlarms(context.Background(), time.Now())
    assert.NoError(t, err)

    // Verify last_sent updated
    verifyLastSentUpdated(t, db)
}
```

## Troubleshooting

### Issue: Duplicate Notifications
**Cause**: `UpdateLastSent` failing after successful send
**Solution**: Add retry logic for `UpdateLastSent`, monitor logs

### Issue: Missing Notifications
**Cause**: Scheduler not running or alarms not retrieved
**Solution**: Add health check endpoint, verify cron schedule

### Issue: High Memory Usage
**Cause**: Too many tokens processed at once
**Solution**: FCM notifier already batches, check alarm distribution

### Issue: Slow Processing
**Cause**: Sequential alarm processing
**Solution**: Consider parallel processing with goroutine pool
