# FCM Notifier Quick Start Guide

## 5-Minute Setup

### 1. Get Firebase Credentials

```bash
# Download from Firebase Console
# Project Settings > Service Accounts > Generate New Private Key
# Save as firebase-credentials.json
```

### 2. Set Environment Variable

```bash
export FCM_CREDENTIALS_PATH=/path/to/firebase-credentials.json
```

### 3. Basic Usage

```go
package main

import (
    "context"
    "time"

    "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/notifier"
    "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
    "go.uber.org/zap"
)

func main() {
    // Initialize
    logger, _ := zap.NewProduction()
    fcm, err := notifier.NewFCMNotifier(
        "/path/to/firebase-credentials.json",
        logger,
    )
    if err != nil {
        panic(err)
    }

    // Send notification
    err = fcm.SendWeatherNotification(
        context.Background(),
        []string{"fcm_token_1", "fcm_token_2"},
        &entity.WeatherData{
            Temperature:   15.5,
            Humidity:      60.0,
            Precipitation: 0.0,
            WindSpeed:     2.5,
            CachedAt:      time.Now(),
        },
        "서울시 강남구",
    )

    if err != nil {
        logger.Error("Failed", zap.Error(err))
    }
}
```

## Common Patterns

### Pattern 1: With Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := fcm.SendWeatherNotification(ctx, tokens, data, region)
```

### Pattern 2: Error Handling

```go
err := fcm.SendWeatherNotification(ctx, tokens, data, region)
switch {
case err == nil:
    // All or some succeeded
    logger.Info("Notifications sent")
case err != nil:
    // All failed
    logger.Error("All notifications failed", zap.Error(err))
}
```

### Pattern 3: Large Token Lists

```go
// Automatically batched (500 per batch)
tokens := make([]string, 1500) // Will send in 3 batches
err := fcm.SendWeatherNotification(ctx, tokens, data, region)
```

## Message Received by Client

### Android (Kotlin)

```kotlin
// In FirebaseMessagingService
override fun onMessageReceived(remoteMessage: RemoteMessage) {
    // Notification
    remoteMessage.notification?.let {
        val title = it.title  // "날씨 알림 - 서울시 강남구"
        val body = it.body    // "현재 15.5°C, 습도 60%, 강수 0.0mm"
    }

    // Data
    remoteMessage.data?.let {
        val region = it["region"]         // "서울시 강남구"
        val temp = it["temperature"]      // "15.5"
        val humidity = it["humidity"]     // "60"
        val precip = it["precipitation"]  // "0.0"
        val wind = it["wind_speed"]       // "2.5"
        val timestamp = it["timestamp"]   // "2025-11-11T12:00:00Z"
    }
}
```

### iOS (Swift)

```swift
// In AppDelegate
func application(_ application: UIApplication,
                 didReceiveRemoteNotification userInfo: [AnyHashable: Any]) {
    // Notification
    if let aps = userInfo["aps"] as? [String: Any],
       let alert = aps["alert"] as? [String: String] {
        let title = alert["title"]  // "날씨 알림 - 서울시 강남구"
        let body = alert["body"]    // "현재 15.5°C, 습도 60%, 강수 0.0mm"
    }

    // Data
    let region = userInfo["region"] as? String         // "서울시 강남구"
    let temp = userInfo["temperature"] as? String      // "15.5"
    let humidity = userInfo["humidity"] as? String     // "60"
    let precip = userInfo["precipitation"] as? String  // "0.0"
    let wind = userInfo["wind_speed"] as? String       // "2.5"
    let timestamp = userInfo["timestamp"] as? String   // "2025-11-11T12:00:00Z"
}
```

## Testing

### Unit Test with Mock

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "firebase.google.com/go/v4/messaging"
)

func TestMyFunction(t *testing.T) {
    // Create mock
    mockClient := &notifier.MockFCMClient{
        SendMulticastFunc: func(ctx context.Context, msg *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
            return &messaging.BatchResponse{
                SuccessCount: len(msg.Tokens),
                FailureCount: 0,
            }, nil
        },
    }

    // Create notifier with mock
    fcm := notifier.NewFCMNotifierWithClient(mockClient, zap.NewNop())

    // Test your code
    err := fcm.SendWeatherNotification(ctx, tokens, data, region)
    assert.NoError(t, err)
}
```

## Troubleshooting

### Problem: "failed to initialize Firebase app"

```bash
# Check file exists
ls -la /path/to/firebase-credentials.json

# Check file permissions
chmod 600 /path/to/firebase-credentials.json

# Verify JSON format
cat /path/to/firebase-credentials.json | python -m json.tool
```

### Problem: "all notification sends failed"

```go
// Add timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

// Check logs
logger.Error("Details", zap.Error(err))
```

### Problem: Some tokens fail

```
// Expected - invalid tokens are logged but don't stop processing
// Check logs for: "Token send failed"
// Clean up invalid tokens from database
```

## Performance Tips

1. **Reuse notifier instance** (don't create per request)
2. **Use context timeouts** (5-10 seconds recommended)
3. **Batch operations automatically handled** (500 per batch)
4. **Monitor logs** for invalid tokens to clean up

## Security Checklist

- [ ] Credentials file not in git (add to .gitignore)
- [ ] File permissions restricted (chmod 600)
- [ ] Environment variable set correctly
- [ ] Credentials rotated periodically
- [ ] No credentials in logs or error messages

## Next Steps

1. Read [README.md](./README.md) for comprehensive documentation
2. Read [INTEGRATION.md](./INTEGRATION.md) for scheduler integration
3. See [example_test.go](./example_test.go) for more examples
4. Check [TASK_SUMMARY.md](./TASK_SUMMARY.md) for implementation details

## Support

- Firebase Console: https://console.firebase.google.com/
- FCM Documentation: https://firebase.google.com/docs/cloud-messaging
- Go Admin SDK: https://firebase.google.com/docs/admin/setup#go
