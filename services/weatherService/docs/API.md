# API Documentation

Internal service interfaces for the Weather Data Collector Service.

**Version:** 1.0.0
**Last Updated:** 2025-11-11

## Table of Contents

- [Overview](#overview)
- [Repository Interface](#repository-interface)
- [Cache Interface](#cache-interface)
- [Crawler Interface](#crawler-interface)
- [Notifier Interface](#notifier-interface)
- [Data Models](#data-models)

## Overview

The Weather Data Collector Service uses interface-based design for modularity, testability, and maintainability. All components depend on abstractions (interfaces) rather than concrete implementations.

### Interface Philosophy

- **Dependency Inversion**: High-level modules don't depend on low-level modules
- **Testability**: Interfaces enable mock implementations for testing
- **Flexibility**: Swap implementations without changing business logic
- **Contract Definition**: Clear API contracts between components

## Repository Interface

### ISchedulerWeatherRepository

Database access layer for alarm and token management.

**Package:** `features/weather/model/interface`

```go
type ISchedulerWeatherRepository interface {
    GetAlarmsToNotify(ctx context.Context, targetTime time.Time) ([]entity.UserAlarm, error)
    UpdateLastSent(ctx context.Context, alarmID int, sentTime time.Time) error
    GetFCMTokens(ctx context.Context, userID int) ([]entity.WeatherServiceToken, error)
}
```

---

#### GetAlarmsToNotify

Retrieves alarms scheduled for notification at the target time.

**Signature:**
```go
GetAlarmsToNotify(ctx context.Context, targetTime time.Time) ([]entity.UserAlarm, error)
```

**Purpose:**
Query database for enabled alarms matching the target time that haven't been sent today.

**Parameters:**
- `ctx` (context.Context): Request context for cancellation and timeouts
- `targetTime` (time.Time): Target notification time (e.g., current time + 1 minute)

**Returns:**
- `[]entity.UserAlarm`: Slice of matching alarms, empty if none found
- `error`: Database error or nil on success

**Behavior:**
- Filters: `is_enabled = true`, `alarm_time = HH:MM`, `last_sent != today`
- Truncates `targetTime` to minute precision (HH:MM:00)
- Returns empty slice (not error) if no alarms match

**Error Conditions:**
- Database connection failure
- Query syntax error
- Context cancellation

**Example Usage:**
```go
targetTime := time.Now().Add(1 * time.Minute)
alarms, err := repo.GetAlarmsToNotify(ctx, targetTime)
if err != nil {
    return fmt.Errorf("failed to get alarms: %w", err)
}

for _, alarm := range alarms {
    log.Printf("Processing alarm %d for user %d", alarm.ID, alarm.UserID)
}
```

**Performance:**
- Query uses composite index: `(alarm_time, is_enabled, last_sent)`
- Expected latency: 5-20ms for 10k alarms
- Scans: ~100 rows per query (1% of total)

**SQL Query:**
```sql
SELECT id, user_id, region, alarm_time, last_sent
FROM weather_alarms
WHERE alarm_time = ?
  AND is_enabled = true
  AND (last_sent IS NULL OR DATE(last_sent) != CURDATE())
```

---

#### UpdateLastSent

Updates the last notification timestamp for an alarm.

**Signature:**
```go
UpdateLastSent(ctx context.Context, alarmID int, sentTime time.Time) error
```

**Purpose:**
Mark alarm as sent to prevent duplicate notifications within the same day.

**Parameters:**
- `ctx` (context.Context): Request context
- `alarmID` (int): Alarm ID to update
- `sentTime` (time.Time): Notification timestamp

**Returns:**
- `error`: Database error or nil on success

**Behavior:**
- Updates `last_sent` column with provided timestamp
- No effect if `alarmID` doesn't exist (idempotent)
- Uses optimistic locking (no row-level locks)

**Error Conditions:**
- Database connection failure
- Invalid `alarmID` (< 1)
- Context cancellation

**Example Usage:**
```go
err := repo.UpdateLastSent(ctx, alarm.ID, time.Now())
if err != nil {
    return fmt.Errorf("failed to update last_sent: %w", err)
}
```

**Performance:**
- Primary key lookup: O(log n)
- Expected latency: 2-5ms
- No index scan required

**SQL Query:**
```sql
UPDATE weather_alarms
SET last_sent = ?
WHERE id = ?
```

---

#### GetFCMTokens

Retrieves FCM registration tokens for a user.

**Signature:**
```go
GetFCMTokens(ctx context.Context, userID int) ([]entity.WeatherServiceToken, error)
```

**Purpose:**
Fetch all valid FCM tokens for sending notifications to a user.

**Parameters:**
- `ctx` (context.Context): Request context
- `userID` (int): User ID

**Returns:**
- `[]entity.WeatherServiceToken`: Slice of token entities, empty if none found
- `error`: Database error or nil on success

**Behavior:**
- Returns all tokens for the user (no active/inactive filtering)
- Returns empty slice (not error) if user has no tokens
- Ordered by creation date (newest first)

**Error Conditions:**
- Database connection failure
- Invalid `userID` (< 1)
- Context cancellation

**Example Usage:**
```go
tokenEntities, err := repo.GetFCMTokens(ctx, alarm.UserID)
if err != nil {
    return fmt.Errorf("failed to get tokens: %w", err)
}

if len(tokenEntities) == 0 {
    log.Printf("No tokens for user %d", alarm.UserID)
    return nil
}

tokens := make([]string, len(tokenEntities))
for i, entity := range tokenEntities {
    tokens[i] = entity.FCMToken
}
```

**Performance:**
- Query uses index: `user_id`
- Expected latency: 3-10ms per user
- Average tokens per user: 1-3

**SQL Query:**
```sql
SELECT id, user_id, fcm_token, created_at
FROM weather_service_tokens
WHERE user_id = ?
ORDER BY created_at DESC
```

---

## Cache Interface

### IWeatherCache

Redis-based weather data caching layer.

**Package:** `features/weather/model/interface`

```go
type IWeatherCache interface {
    Get(ctx context.Context, region string) (*entity.WeatherData, error)
    Set(ctx context.Context, region string, data *entity.WeatherData) error
    Delete(ctx context.Context, region string) error
    Close() error
    Ping(ctx context.Context) error
    GetTTL(ctx context.Context, region string) (time.Duration, error)
}
```

---

#### Get

Retrieves cached weather data for a region.

**Signature:**
```go
Get(ctx context.Context, region string) (*entity.WeatherData, error)
```

**Purpose:**
Fetch weather data from cache to avoid crawler API calls.

**Parameters:**
- `ctx` (context.Context): Request context
- `region` (string): Region name (e.g., "서울", "부산")

**Returns:**
- `*entity.WeatherData`: Weather data or nil if cache miss
- `error`: Redis error or nil (cache miss returns nil, nil)

**Behavior:**
- Returns `nil, nil` on cache miss (not error)
- Deserializes JSON from Redis
- Returns data even if stale (TTL expired)

**Error Conditions:**
- Redis connection failure
- JSON deserialization error
- Context cancellation

**Example Usage:**
```go
data, err := cache.Get(ctx, "서울")
if err != nil {
    return fmt.Errorf("cache error: %w", err)
}

if data == nil {
    // Cache miss - fetch from crawler
    data, err = crawler.Fetch(ctx, "서울")
}
```

**Performance:**
- Expected latency: 1-3ms (local Redis)
- Cache key: `weather:{region}`
- Serialization: JSON (300-500 bytes per entry)

---

#### Set

Stores weather data in cache with TTL.

**Signature:**
```go
Set(ctx context.Context, region string, data *entity.WeatherData) error
```

**Purpose:**
Cache weather data to reduce crawler API calls.

**Parameters:**
- `ctx` (context.Context): Request context
- `region` (string): Region name
- `data` (*entity.WeatherData): Weather data to cache

**Returns:**
- `error`: Redis error or nil on success

**Behavior:**
- Serializes data to JSON
- Sets TTL to 10 minutes
- Overwrites existing entry

**Error Conditions:**
- Redis connection failure
- JSON serialization error
- Context cancellation

**Example Usage:**
```go
data := &entity.WeatherData{
    Temperature: 15.5,
    Humidity: 60,
    Precipitation: 0,
    WindSpeed: 2.5,
    CachedAt: time.Now(),
}

err := cache.Set(ctx, "서울", data)
if err != nil {
    log.Printf("Failed to cache data: %v", err)
    // Continue even if caching fails
}
```

**Performance:**
- Expected latency: 2-4ms
- TTL: 10 minutes
- Auto-expiration: Redis eviction

---

#### Delete

Removes weather data from cache.

**Signature:**
```go
Delete(ctx context.Context, region string) error
```

**Purpose:**
Invalidate cached data (e.g., after crawler failure).

**Parameters:**
- `ctx` (context.Context): Request context
- `region` (string): Region name

**Returns:**
- `error`: Redis error or nil on success

**Example Usage:**
```go
err := cache.Delete(ctx, "서울")
if err != nil {
    return fmt.Errorf("failed to delete cache: %w", err)
}
```

**Performance:**
- Expected latency: 1-2ms

---

#### Ping

Checks cache connection health.

**Signature:**
```go
Ping(ctx context.Context) error
```

**Purpose:**
Verify Redis connection for health checks.

**Returns:**
- `error`: Connection error or nil if healthy

**Example Usage:**
```go
if err := cache.Ping(ctx); err != nil {
    log.Printf("Cache unhealthy: %v", err)
}
```

---

#### GetTTL

Gets remaining TTL for cached entry.

**Signature:**
```go
GetTTL(ctx context.Context, region string) (time.Duration, error)
```

**Purpose:**
Check cache freshness.

**Returns:**
- `time.Duration`: Remaining TTL (0 if expired/missing)
- `error`: Redis error or nil

**Example Usage:**
```go
ttl, err := cache.GetTTL(ctx, "서울")
if err != nil {
    return err
}

if ttl < 2*time.Minute {
    // Refresh cache proactively
}
```

---

## Crawler Interface

### IWeatherCrawler

Weather data fetcher interface.

**Package:** `features/weather/model/interface`

```go
type IWeatherCrawler interface {
    Fetch(ctx context.Context, region string) (*entity.WeatherData, error)
}
```

---

#### Fetch

Fetches real-time weather data from Naver Weather.

**Signature:**
```go
Fetch(ctx context.Context, region string) (*entity.WeatherData, error)
```

**Purpose:**
Retrieve current weather data when cache misses.

**Parameters:**
- `ctx` (context.Context): Request context
- `region` (string): Region name (Korean)

**Returns:**
- `*entity.WeatherData`: Current weather data
- `error`: Fetch/parse error

**Behavior:**
- Retries 3 times with exponential backoff (1s, 2s, 4s)
- Sets User-Agent to avoid blocking
- Parses HTML with goquery
- Returns error after max retries

**Error Conditions:**
- Network timeout
- HTTP status != 200
- HTML parsing failure
- Invalid region name
- Context cancellation

**Example Usage:**
```go
data, err := crawler.Fetch(ctx, "서울")
if err != nil {
    return fmt.Errorf("failed to fetch weather: %w", err)
}

log.Printf("Temperature: %.1f°C", data.Temperature)
```

**Performance:**
- Expected latency: 500-2000ms (network)
- Timeout: 10 seconds per attempt
- Max retries: 3
- Total max time: ~30 seconds

---

## Notifier Interface

### IFCMNotifier

FCM notification sender interface.

**Package:** `features/weather/model/interface`

```go
type IFCMNotifier interface {
    SendWeatherNotification(ctx context.Context, tokens []string, data *entity.WeatherData, region string) error
}
```

---

#### SendWeatherNotification

Sends weather notifications via FCM.

**Signature:**
```go
SendWeatherNotification(ctx context.Context, tokens []string, data *entity.WeatherData, region string) error
```

**Purpose:**
Deliver weather alerts to user devices.

**Parameters:**
- `ctx` (context.Context): Request context
- `tokens` ([]string): FCM registration tokens
- `data` (*entity.WeatherData): Weather data
- `region` (string): Region name

**Returns:**
- `error`: Error only if ALL sends fail

**Behavior:**
- Splits tokens into batches of 500 (FCM limit)
- Retries failed batches once
- Logs individual token failures
- Returns nil if any send succeeds

**Error Conditions:**
- All batches fail
- Invalid FCM credentials
- Network failure
- Invalid tokens (logged, not returned)

**Example Usage:**
```go
tokens := []string{"token1", "token2", "token3"}
data := &entity.WeatherData{
    Temperature: 15.5,
    Humidity: 60,
    Precipitation: 0,
    WindSpeed: 2.5,
}

err := notifier.SendWeatherNotification(ctx, tokens, data, "서울")
if err != nil {
    return fmt.Errorf("all sends failed: %w", err)
}
```

**Performance:**
- Batch size: 500 tokens
- Expected latency: 200-800ms per batch
- Success rate: >95%
- Retry: 1 attempt per batch

**Notification Format:**
```json
{
  "notification": {
    "title": "날씨 알림 - 서울",
    "body": "현재 15.5°C, 습도 60%, 강수 0.0mm"
  },
  "data": {
    "region": "서울",
    "temperature": "15.5",
    "humidity": "60",
    "precipitation": "0.0",
    "wind_speed": "2.5",
    "timestamp": "2025-11-11T10:30:00+09:00"
  }
}
```

---

## Data Models

### UserAlarm

**Package:** `features/weather/model/entity`

```go
type UserAlarm struct {
    ID        int       // Primary key
    UserID    int       // User ID
    Region    string    // Region name (e.g., "서울")
    AlarmTime string    // Time in HH:MM format (e.g., "09:00")
    IsEnabled bool      // Enabled flag
    LastSent  time.Time // Last notification timestamp
    CreatedAt time.Time // Creation timestamp
    UpdatedAt time.Time // Update timestamp
}
```

**Usage:**
```go
alarm := entity.UserAlarm{
    ID:        1,
    UserID:    100,
    Region:    "서울",
    AlarmTime: "09:00",
    IsEnabled: true,
}
```

---

### WeatherData

**Package:** `features/weather/model/entity`

```go
type WeatherData struct {
    Temperature   float64   // Temperature in Celsius
    Humidity      float64   // Humidity percentage (0-100)
    Precipitation float64   // Precipitation in mm
    WindSpeed     float64   // Wind speed in m/s
    CachedAt      time.Time // Cache timestamp
}
```

**Usage:**
```go
data := &entity.WeatherData{
    Temperature:   15.5,
    Humidity:      60.0,
    Precipitation: 0.0,
    WindSpeed:     2.5,
    CachedAt:      time.Now(),
}
```

---

### WeatherServiceToken

**Package:** `features/weather/model/entity`

```go
type WeatherServiceToken struct {
    ID        int       // Primary key
    UserID    int       // User ID
    FCMToken  string    // FCM registration token
    CreatedAt time.Time // Registration timestamp
}
```

**Usage:**
```go
token := entity.WeatherServiceToken{
    ID:       1,
    UserID:   100,
    FCMToken: "fcm_token_abc123...",
}
```

---

## Error Handling

### Error Types

**Database Errors:**
```go
// Connection failure
err := repo.GetAlarmsToNotify(ctx, time.Now())
// Returns: failed to connect to database: dial tcp: connection refused

// Query timeout
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
// Returns: context deadline exceeded
```

**Cache Errors:**
```go
// Redis connection failure
err := cache.Get(ctx, "서울")
// Returns: redis: connection pool exhausted

// Serialization failure
err := cache.Set(ctx, "서울", invalidData)
// Returns: json: unsupported type: chan int
```

**Crawler Errors:**
```go
// Network timeout
err := crawler.Fetch(ctx, "서울")
// Returns: failed after 3 attempts: context deadline exceeded

// Parse failure
err := crawler.Fetch(ctx, "InvalidRegion")
// Returns: failed to extract temperature from HTML
```

**Notifier Errors:**
```go
// All sends failed
err := notifier.SendWeatherNotification(ctx, tokens, data, "서울")
// Returns: all notification sends failed: 150 failures

// Invalid credentials
err := notifier.SendWeatherNotification(ctx, tokens, data, "서울")
// Returns: failed to initialize Firebase app: credentials error
```

### Error Handling Best Practices

1. **Always check errors:**
   ```go
   if err != nil {
       return fmt.Errorf("operation failed: %w", err)
   }
   ```

2. **Distinguish cache miss from error:**
   ```go
   data, err := cache.Get(ctx, region)
   if err != nil {
       log.Printf("Cache error: %v", err)
   }
   if data == nil {
       // Cache miss - fetch from crawler
   }
   ```

3. **Continue on non-critical failures:**
   ```go
   if err := cache.Set(ctx, region, data); err != nil {
       log.Printf("Failed to cache: %v", err)
       // Continue processing
   }
   ```

4. **Use context for timeouts:**
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
   defer cancel()
   ```

---

## Version History

- **v1.0.0** (2025-11-11): Initial release
