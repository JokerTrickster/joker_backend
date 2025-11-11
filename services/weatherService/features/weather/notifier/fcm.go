package notifier

import (
	"context"
	"fmt"
	"strconv"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/interface"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

const (
	// MaxTokensPerBatch is the maximum number of tokens FCM allows per batch
	MaxTokensPerBatch = 500
)

// FCMNotifier handles sending weather notifications via Firebase Cloud Messaging
type FCMNotifier struct {
	client _interface.IFCMClient
	logger *zap.Logger
}

// NewFCMNotifier creates a new FCM notifier instance
// credentialsPath: path to Firebase service account credentials JSON file
// logger: zap logger for logging
// Returns error if Firebase app initialization fails
func NewFCMNotifier(credentialsPath string, logger *zap.Logger) (*FCMNotifier, error) {
	if credentialsPath == "" {
		return nil, fmt.Errorf("FCM credentials path is required")
	}

	ctx := context.Background()
	opt := option.WithCredentialsFile(credentialsPath)

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	messagingClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get messaging client: %w", err)
	}

	return &FCMNotifier{
		client: NewFCMClientWrapper(messagingClient),
		logger: logger,
	}, nil
}

// NewFCMNotifierWithClient creates a new FCM notifier with a custom client (for testing)
func NewFCMNotifierWithClient(client _interface.IFCMClient, logger *zap.Logger) *FCMNotifier {
	return &FCMNotifier{
		client: client,
		logger: logger,
	}
}

// SendWeatherNotification sends weather notifications to multiple FCM tokens
// Handles batch processing (max 500 tokens per batch)
// Returns error only if ALL sends fail
// Logs individual token failures but continues processing
func (n *FCMNotifier) SendWeatherNotification(ctx context.Context, tokens []string, data *entity.WeatherData, region string) error {
	if len(tokens) == 0 {
		n.logger.Warn("No FCM tokens provided for notification")
		return nil
	}

	if data == nil {
		return fmt.Errorf("weather data is required")
	}

	if region == "" {
		return fmt.Errorf("region is required")
	}

	// Split tokens into batches if necessary
	batches := splitTokensIntoBatches(tokens, MaxTokensPerBatch)
	totalBatches := len(batches)

	totalSuccess := 0
	totalFailure := 0

	n.logger.Info("Starting weather notification send",
		zap.Int("total_tokens", len(tokens)),
		zap.Int("batches", totalBatches),
	)

	for i, batch := range batches {
		batchNum := i + 1
		n.logger.Debug("Processing batch",
			zap.Int("batch", batchNum),
			zap.Int("total_batches", totalBatches),
			zap.Int("tokens_in_batch", len(batch)),
		)

		message := n.formatMessage(data, region, batch)

		response, err := n.client.SendMulticast(ctx, message)
		if err != nil {
			// Network error - log and continue with retry logic
			n.logger.Error("Failed to send batch",
				zap.Int("batch", batchNum),
				zap.Error(err),
			)

			// Retry once
			n.logger.Info("Retrying batch", zap.Int("batch", batchNum))
			response, err = n.client.SendMulticast(ctx, message)
			if err != nil {
				n.logger.Error("Batch retry failed",
					zap.Int("batch", batchNum),
					zap.Error(err),
				)
				totalFailure += len(batch)
				continue
			}
		}

		// Process batch response
		successCount := response.SuccessCount
		failureCount := response.FailureCount
		totalSuccess += successCount
		totalFailure += failureCount

		n.logger.Info("Batch completed",
			zap.Int("batch", batchNum),
			zap.Int("success", successCount),
			zap.Int("failure", failureCount),
		)

		// Log individual failures
		if failureCount > 0 {
			for idx, resp := range response.Responses {
				if !resp.Success {
					n.logger.Warn("Token send failed",
						zap.Int("batch", batchNum),
						zap.Int("token_index", idx),
						zap.String("error", resp.Error.Error()),
					)
				}
			}
		}
	}

	n.logger.Info("Weather notification send completed",
		zap.Int("total_success", totalSuccess),
		zap.Int("total_failure", totalFailure),
		zap.String("region", region),
	)

	// Return error only if all sends failed
	if totalSuccess == 0 && totalFailure > 0 {
		return fmt.Errorf("all notification sends failed: %d failures", totalFailure)
	}

	return nil
}

// formatMessage creates a multicast message for FCM
func (n *FCMNotifier) formatMessage(data *entity.WeatherData, region string, tokens []string) *messaging.MulticastMessage {
	// Format temperature with 1 decimal place
	tempStr := strconv.FormatFloat(data.Temperature, 'f', 1, 64)
	humidityStr := strconv.FormatFloat(data.Humidity, 'f', 0, 64)
	precipStr := strconv.FormatFloat(data.Precipitation, 'f', 1, 64)
	windStr := strconv.FormatFloat(data.WindSpeed, 'f', 1, 64)

	return &messaging.MulticastMessage{
		Notification: &messaging.Notification{
			Title: fmt.Sprintf("날씨 알림 - %s", region),
			Body:  fmt.Sprintf("현재 %s°C, 습도 %s%%, 강수 %smm", tempStr, humidityStr, precipStr),
		},
		Data: map[string]string{
			"region":        region,
			"temperature":   tempStr,
			"humidity":      humidityStr,
			"precipitation": precipStr,
			"wind_speed":    windStr,
			"timestamp":     data.CachedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		Tokens: tokens,
	}
}

// splitTokensIntoBatches splits a slice of tokens into batches of specified size
func splitTokensIntoBatches(tokens []string, batchSize int) [][]string {
	var batches [][]string

	for i := 0; i < len(tokens); i += batchSize {
		end := i + batchSize
		if end > len(tokens) {
			end = len(tokens)
		}
		batches = append(batches, tokens[i:end])
	}

	return batches
}
