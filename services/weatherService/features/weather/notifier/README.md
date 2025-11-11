# Weather Notification Package

This package provides FCM (Firebase Cloud Messaging) notification functionality for the weather service.

## Features

- **Batch Processing**: Automatically splits large token lists into batches (max 500 per batch as per FCM limits)
- **Retry Logic**: Network failures are automatically retried once
- **Partial Failure Handling**: Logs individual token failures but continues processing
- **Context Support**: Respects context cancellation and deadlines
- **Structured Logging**: Comprehensive logging with zap for debugging and monitoring
- **Testable Design**: Interface-based design with mock support for testing

## Installation

```bash
go get firebase.google.com/go/v4
go get firebase.google.com/go/v4/messaging
```

## Configuration

### Environment Variables

```bash
FCM_CREDENTIALS_PATH=/path/to/firebase-credentials.json
```

### Firebase Credentials

1. Go to [Firebase Console](https://console.firebase.google.com/)
2. Select your project
3. Go to Project Settings > Service Accounts
4. Click "Generate New Private Key"
5. Save the JSON file securely
6. Set `FCM_CREDENTIALS_PATH` to the file path

## Usage

### Basic Usage

```go
import (
    "context"
    "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/notifier"
    "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
    "go.uber.org/zap"
)

// Initialize logger
logger, _ := zap.NewProduction()

// Create FCM notifier
fcmNotifier, err := notifier.NewFCMNotifier("/path/to/credentials.json", logger)
if err != nil {
    log.Fatal(err)
}

// Prepare weather data
weatherData := &entity.WeatherData{
    Temperature:   15.5,
    Humidity:      60.0,
    Precipitation: 0.0,
    WindSpeed:     2.5,
    CachedAt:      time.Now(),
}

// FCM tokens
tokens := []string{"token1", "token2", "token3"}

// Send notifications
err = fcmNotifier.SendWeatherNotification(context.Background(), tokens, weatherData, "서울시 강남구")
if err != nil {
    logger.Error("Failed to send notifications", zap.Error(err))
}
```

### With Context Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err = fcmNotifier.SendWeatherNotification(ctx, tokens, weatherData, region)
```

### Integration with Scheduler

```go
// Get alarms to notify
alarms, err := schedulerRepo.GetAlarmsToNotify(ctx, targetTime)

for _, alarm := range alarms {
    // Get FCM tokens for user
    tokenEntities, err := schedulerRepo.GetFCMTokens(ctx, alarm.UserID)

    // Extract token strings
    tokens := make([]string, len(tokenEntities))
    for i, t := range tokenEntities {
        tokens[i] = t.FCMToken
    }

    // Get weather data from cache
    weatherData, err := weatherCache.Get(ctx, alarm.Region)

    // Send notifications
    err = fcmNotifier.SendWeatherNotification(ctx, tokens, weatherData, alarm.Region)
    if err != nil {
        logger.Error("Failed to send notification",
            zap.Int("alarm_id", alarm.ID),
            zap.Int("user_id", alarm.UserID),
            zap.Error(err),
        )
        continue
    }

    // Update last sent timestamp
    err = schedulerRepo.UpdateLastSent(ctx, alarm.ID, time.Now())
}
```

## Message Format

### Notification (displayed to user)

```
Title: "날씨 알림 - {region}"
Body:  "현재 {temperature}°C, 습도 {humidity}%, 강수 {precipitation}mm"
```

Example:
```
Title: "날씨 알림 - 서울시 강남구"
Body:  "현재 15.5°C, 습도 60%, 강수 0.0mm"
```

### Data Payload (for app logic)

```json
{
  "region": "서울시 강남구",
  "temperature": "15.5",
  "humidity": "60",
  "precipitation": "0.0",
  "wind_speed": "2.5",
  "timestamp": "2025-11-11T12:00:00Z"
}
```

## Batch Processing

The notifier automatically splits tokens into batches:

- **Max tokens per batch**: 500 (FCM limit)
- **Processing**: Sequential batch processing
- **Logging**: Progress logged for each batch
- **Failure handling**: Independent batch failures

Example with 1500 tokens:
- Batch 1: tokens 0-499
- Batch 2: tokens 500-999
- Batch 3: tokens 1000-1499

## Error Handling

### Return Conditions

| Scenario | Return Value | Behavior |
|----------|-------------|----------|
| All tokens sent successfully | `nil` | Normal success |
| Some tokens failed | `nil` | Logs failures, continues |
| All tokens failed | `error` | Returns error |
| Network error (no retry success) | `error` | Returns error |
| Empty tokens | `nil` | Logs warning |
| Nil weather data | `error` | Returns error |
| Empty region | `error` | Returns error |
| Context cancelled | `error` | Returns error |

### Retry Logic

- **Network failures**: Retried once automatically
- **Invalid tokens**: Not retried (logged as warnings)
- **Batch failures**: Each batch retried independently

## Testing

### Run Tests

```bash
go test ./features/weather/notifier/... -v
```

### Test Coverage

```bash
go test ./features/weather/notifier/... -cover
```

### Mock Client

For testing, use the provided interface:

```go
type MockFCMClient struct {
    SendMulticastFunc func(ctx context.Context, message *messaging.MulticastMessage) (*messaging.BatchResponse, error)
}

func (m *MockFCMClient) SendMulticast(ctx context.Context, message *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
    if m.SendMulticastFunc != nil {
        return m.SendMulticastFunc(ctx, message)
    }
    return &messaging.BatchResponse{
        SuccessCount: len(message.Tokens),
        FailureCount: 0,
    }, nil
}

// Create notifier with mock
notifier := notifier.NewFCMNotifierWithClient(mockClient, logger)
```

## Logging

The notifier provides comprehensive logging:

### Info Level
- Batch processing start/complete
- Success/failure counts
- Retry attempts

### Warn Level
- Empty token lists
- Individual token failures

### Error Level
- Network failures
- Batch send failures
- Retry failures

### Debug Level
- Batch details (number, size)
- Individual batch processing

Example logs:
```
INFO  Starting weather notification send  tokens=1500 batches=3
DEBUG Processing batch  batch=1 total_batches=3 tokens_in_batch=500
INFO  Batch completed  batch=1 success=498 failure=2
WARN  Token send failed  batch=1 token_index=45 error="invalid token"
INFO  Weather notification send completed  total_success=1495 total_failure=5 region="서울시 강남구"
```

## Performance Considerations

- **Batch size**: 500 tokens per batch (FCM maximum)
- **Sequential processing**: Batches processed one at a time
- **Memory efficiency**: Tokens not duplicated during batching
- **Context timeout**: Recommended 5-10 seconds per batch
- **Concurrent safety**: Thread-safe, can be called concurrently

## Security Considerations

- **Credentials**: Store Firebase credentials securely, never commit to git
- **Token validation**: Invalid tokens are logged but don't stop processing
- **Error messages**: Sensitive information not exposed in logs
- **Context cancellation**: Respects cancellation to prevent resource leaks

## API Reference

### `NewFCMNotifier(credentialsPath string, logger *zap.Logger) (*FCMNotifier, error)`

Creates a new FCM notifier with Firebase credentials.

**Parameters:**
- `credentialsPath`: Path to Firebase service account credentials JSON
- `logger`: Zap logger instance

**Returns:**
- `*FCMNotifier`: Initialized notifier
- `error`: Error if initialization fails

### `NewFCMNotifierWithClient(client IFCMClient, logger *zap.Logger) *FCMNotifier`

Creates a new FCM notifier with a custom client (for testing).

**Parameters:**
- `client`: Custom FCM client implementation
- `logger`: Zap logger instance

**Returns:**
- `*FCMNotifier`: Initialized notifier

### `SendWeatherNotification(ctx context.Context, tokens []string, data *entity.WeatherData, region string) error`

Sends weather notifications to multiple FCM tokens.

**Parameters:**
- `ctx`: Context for cancellation and timeout
- `tokens`: List of FCM tokens to send to
- `data`: Weather data to include in notification
- `region`: Region name for notification

**Returns:**
- `error`: Error only if all sends fail, nil for partial success

## Troubleshooting

### Issue: "failed to initialize Firebase app"
**Solution**: Verify credentials file path and format

### Issue: "all notification sends failed"
**Solution**: Check network connectivity and Firebase project configuration

### Issue: Context deadline exceeded
**Solution**: Increase context timeout or reduce batch size

### Issue: Invalid tokens
**Solution**: Clean up expired tokens from database using GetFCMTokens filter

## Future Enhancements

- [ ] Configurable retry count
- [ ] Exponential backoff for retries
- [ ] Token validation before sending
- [ ] Metrics collection (success/failure rates)
- [ ] Rate limiting support
- [ ] Custom notification templates
