package notifier_test

import (
	"context"
	"fmt"
	"time"

	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/notifier"
	"go.uber.org/zap"
)

// Example demonstrates basic usage of FCM notifier
func Example_basicUsage() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Create FCM notifier with credentials
	fcmNotifier, err := notifier.NewFCMNotifier("/path/to/firebase-credentials.json", logger)
	if err != nil {
		logger.Fatal("Failed to create FCM notifier", zap.Error(err))
	}

	// Prepare weather data
	weatherData := &entity.WeatherData{
		Temperature:   15.5,
		Humidity:      60.0,
		Precipitation: 0.0,
		WindSpeed:     2.5,
		CachedAt:      time.Now(),
	}

	// FCM tokens from database
	tokens := []string{
		"fcm_token_1",
		"fcm_token_2",
		"fcm_token_3",
	}

	// Send notifications
	ctx := context.Background()
	region := "서울시 강남구"

	err = fcmNotifier.SendWeatherNotification(ctx, tokens, weatherData, region)
	if err != nil {
		logger.Error("Failed to send notifications", zap.Error(err))
	} else {
		logger.Info("Notifications sent successfully")
	}
}

// Example_withTimeout demonstrates using context timeout
func Example_withTimeout() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	fcmNotifier, err := notifier.NewFCMNotifier("/path/to/firebase-credentials.json", logger)
	if err != nil {
		logger.Fatal("Failed to create FCM notifier", zap.Error(err))
	}

	weatherData := &entity.WeatherData{
		Temperature:   20.0,
		Humidity:      65.0,
		Precipitation: 1.5,
		WindSpeed:     3.0,
		CachedAt:      time.Now(),
	}

	tokens := []string{"fcm_token_1", "fcm_token_2"}

	// Create context with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = fcmNotifier.SendWeatherNotification(ctx, tokens, weatherData, "부산시 해운대구")
	if err != nil {
		logger.Error("Failed to send notifications", zap.Error(err))
	}
}

// Example_largeBatch demonstrates batch processing for many tokens
func Example_largeBatch() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	fcmNotifier, err := notifier.NewFCMNotifier("/path/to/firebase-credentials.json", logger)
	if err != nil {
		logger.Fatal("Failed to create FCM notifier", zap.Error(err))
	}

	weatherData := &entity.WeatherData{
		Temperature:   18.0,
		Humidity:      70.0,
		Precipitation: 2.0,
		WindSpeed:     4.5,
		CachedAt:      time.Now(),
	}

	// Simulate 1500 tokens (will be split into 3 batches of 500)
	tokens := make([]string, 1500)
	for i := range tokens {
		tokens[i] = fmt.Sprintf("fcm_token_%d", i)
	}

	ctx := context.Background()

	// FCM notifier will automatically split into batches
	err = fcmNotifier.SendWeatherNotification(ctx, tokens, weatherData, "인천시 중구")
	if err != nil {
		logger.Error("Some or all notifications failed", zap.Error(err))
	} else {
		logger.Info("All notifications sent successfully")
	}
}

// Example_schedulerIntegration demonstrates integration with scheduler repository
func Example_schedulerIntegration() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Initialize FCM notifier
	fcmNotifier, err := notifier.NewFCMNotifier("/path/to/firebase-credentials.json", logger)
	if err != nil {
		logger.Fatal("Failed to create FCM notifier", zap.Error(err))
	}

	// Example: Processing alarms from scheduler
	// In real implementation, these would come from SchedulerWeatherRepository

	// Alarm data
	userID := 123
	region := "대전시 서구"

	// Tokens from GetFCMTokens
	tokensFromDB := []entity.WeatherServiceToken{
		{FCMToken: "token_1"},
		{FCMToken: "token_2"},
		{FCMToken: "token_3"},
	}

	// Extract token strings
	tokens := make([]string, len(tokensFromDB))
	for i, t := range tokensFromDB {
		tokens[i] = t.FCMToken
	}

	// Weather data from cache/crawler
	weatherData := &entity.WeatherData{
		Temperature:   22.5,
		Humidity:      58.0,
		Precipitation: 0.0,
		WindSpeed:     2.0,
		CachedAt:      time.Now(),
	}

	ctx := context.Background()

	// Send notifications
	err = fcmNotifier.SendWeatherNotification(ctx, tokens, weatherData, region)
	if err != nil {
		logger.Error("Notification failed",
			zap.Int("user_id", userID),
			zap.String("region", region),
			zap.Error(err),
		)
	} else {
		logger.Info("Notification sent",
			zap.Int("user_id", userID),
			zap.String("region", region),
			zap.Int("tokens_sent", len(tokens)),
		)
	}
}
