package notifier

import (
	"context"
	"errors"
	"testing"
	"time"

	"firebase.google.com/go/v4/messaging"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// MockFCMClient is a mock implementation of IFCMClient for testing
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
		Responses:    make([]*messaging.SendResponse, len(message.Tokens)),
	}, nil
}

func TestNewFCMNotifier(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name            string
		credentialsPath string
		expectError     bool
	}{
		{
			name:            "Empty credentials path",
			credentialsPath: "",
			expectError:     true,
		},
		{
			name:            "Invalid credentials path",
			credentialsPath: "/invalid/path/to/credentials.json",
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFCMNotifier(tt.credentialsPath, logger)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFormatMessage(t *testing.T) {
	logger := zap.NewNop()
	mockClient := &MockFCMClient{}
	notifier := NewFCMNotifierWithClient(mockClient, logger)

	weatherData := &entity.WeatherData{
		Temperature:   15.5,
		Humidity:      60.0,
		Precipitation: 0.0,
		WindSpeed:     2.5,
		CachedAt:      time.Date(2025, 11, 11, 12, 0, 0, 0, time.UTC),
	}

	tokens := []string{"token1", "token2"}
	region := "서울시 강남구"

	message := notifier.formatMessage(weatherData, region, tokens)

	assert.NotNil(t, message)
	assert.NotNil(t, message.Notification)
	assert.Equal(t, "날씨 알림 - 서울시 강남구", message.Notification.Title)
	assert.Equal(t, "현재 15.5°C, 습도 60%, 강수 0.0mm", message.Notification.Body)
	assert.Equal(t, tokens, message.Tokens)

	assert.Equal(t, "서울시 강남구", message.Data["region"])
	assert.Equal(t, "15.5", message.Data["temperature"])
	assert.Equal(t, "60", message.Data["humidity"])
	assert.Equal(t, "0.0", message.Data["precipitation"])
	assert.Equal(t, "2.5", message.Data["wind_speed"])
	assert.Equal(t, "2025-11-11T12:00:00Z", message.Data["timestamp"])
}

func TestSendWeatherNotification_Success(t *testing.T) {
	logger := zap.NewNop()
	mockClient := &MockFCMClient{
		SendMulticastFunc: func(ctx context.Context, message *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
			responses := make([]*messaging.SendResponse, len(message.Tokens))
			for i := range responses {
				responses[i] = &messaging.SendResponse{
					Success: true,
				}
			}
			return &messaging.BatchResponse{
				SuccessCount: len(message.Tokens),
				FailureCount: 0,
				Responses:    responses,
			}, nil
		},
	}

	notifier := NewFCMNotifierWithClient(mockClient, logger)

	weatherData := &entity.WeatherData{
		Temperature:   20.0,
		Humidity:      55.0,
		Precipitation: 2.5,
		WindSpeed:     3.0,
		CachedAt:      time.Now(),
	}

	tokens := []string{"token1", "token2", "token3"}
	region := "부산시 해운대구"

	err := notifier.SendWeatherNotification(context.Background(), tokens, weatherData, region)
	assert.NoError(t, err)
}

func TestSendWeatherNotification_EmptyTokens(t *testing.T) {
	logger := zap.NewNop()
	mockClient := &MockFCMClient{}
	notifier := NewFCMNotifierWithClient(mockClient, logger)

	weatherData := &entity.WeatherData{
		Temperature:   20.0,
		Humidity:      55.0,
		Precipitation: 2.5,
		WindSpeed:     3.0,
		CachedAt:      time.Now(),
	}

	err := notifier.SendWeatherNotification(context.Background(), []string{}, weatherData, "서울")
	assert.NoError(t, err) // Empty tokens should not return error, just log warning
}

func TestSendWeatherNotification_NilWeatherData(t *testing.T) {
	logger := zap.NewNop()
	mockClient := &MockFCMClient{}
	notifier := NewFCMNotifierWithClient(mockClient, logger)

	tokens := []string{"token1"}

	err := notifier.SendWeatherNotification(context.Background(), tokens, nil, "서울")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "weather data is required")
}

func TestSendWeatherNotification_EmptyRegion(t *testing.T) {
	logger := zap.NewNop()
	mockClient := &MockFCMClient{}
	notifier := NewFCMNotifierWithClient(mockClient, logger)

	weatherData := &entity.WeatherData{
		Temperature:   20.0,
		Humidity:      55.0,
		Precipitation: 2.5,
		WindSpeed:     3.0,
		CachedAt:      time.Now(),
	}

	tokens := []string{"token1"}

	err := notifier.SendWeatherNotification(context.Background(), tokens, weatherData, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "region is required")
}

func TestSendWeatherNotification_PartialFailure(t *testing.T) {
	logger := zap.NewNop()
	mockClient := &MockFCMClient{
		SendMulticastFunc: func(ctx context.Context, message *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
			responses := []*messaging.SendResponse{
				{Success: true},
				{Success: false, Error: errors.New("invalid token")},
				{Success: true},
			}
			return &messaging.BatchResponse{
				SuccessCount: 2,
				FailureCount: 1,
				Responses:    responses,
			}, nil
		},
	}

	notifier := NewFCMNotifierWithClient(mockClient, logger)

	weatherData := &entity.WeatherData{
		Temperature:   18.5,
		Humidity:      70.0,
		Precipitation: 0.0,
		WindSpeed:     1.5,
		CachedAt:      time.Now(),
	}

	tokens := []string{"valid_token1", "invalid_token", "valid_token2"}
	region := "인천시 중구"

	err := notifier.SendWeatherNotification(context.Background(), tokens, weatherData, region)
	assert.NoError(t, err) // Partial success should not return error
}

func TestSendWeatherNotification_AllFailures(t *testing.T) {
	logger := zap.NewNop()
	mockClient := &MockFCMClient{
		SendMulticastFunc: func(ctx context.Context, message *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
			responses := make([]*messaging.SendResponse, len(message.Tokens))
			for i := range responses {
				responses[i] = &messaging.SendResponse{
					Success: false,
					Error:   errors.New("invalid token"),
				}
			}
			return &messaging.BatchResponse{
				SuccessCount: 0,
				FailureCount: len(message.Tokens),
				Responses:    responses,
			}, nil
		},
	}

	notifier := NewFCMNotifierWithClient(mockClient, logger)

	weatherData := &entity.WeatherData{
		Temperature:   18.5,
		Humidity:      70.0,
		Precipitation: 0.0,
		WindSpeed:     1.5,
		CachedAt:      time.Now(),
	}

	tokens := []string{"invalid_token1", "invalid_token2"}
	region := "대전시 서구"

	err := notifier.SendWeatherNotification(context.Background(), tokens, weatherData, region)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "all notification sends failed")
}

func TestSendWeatherNotification_NetworkErrorWithRetry(t *testing.T) {
	logger := zap.NewNop()
	callCount := 0
	mockClient := &MockFCMClient{
		SendMulticastFunc: func(ctx context.Context, message *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
			callCount++
			if callCount == 1 {
				// First call fails
				return nil, errors.New("network error")
			}
			// Retry succeeds
			responses := make([]*messaging.SendResponse, len(message.Tokens))
			for i := range responses {
				responses[i] = &messaging.SendResponse{
					Success: true,
				}
			}
			return &messaging.BatchResponse{
				SuccessCount: len(message.Tokens),
				FailureCount: 0,
				Responses:    responses,
			}, nil
		},
	}

	notifier := NewFCMNotifierWithClient(mockClient, logger)

	weatherData := &entity.WeatherData{
		Temperature:   22.0,
		Humidity:      65.0,
		Precipitation: 1.0,
		WindSpeed:     4.0,
		CachedAt:      time.Now(),
	}

	tokens := []string{"token1", "token2"}
	region := "광주시 북구"

	err := notifier.SendWeatherNotification(context.Background(), tokens, weatherData, region)
	assert.NoError(t, err)
	assert.Equal(t, 2, callCount) // Should have retried once
}

func TestSendWeatherNotification_NetworkErrorRetryFails(t *testing.T) {
	logger := zap.NewNop()
	mockClient := &MockFCMClient{
		SendMulticastFunc: func(ctx context.Context, message *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
			return nil, errors.New("persistent network error")
		},
	}

	notifier := NewFCMNotifierWithClient(mockClient, logger)

	weatherData := &entity.WeatherData{
		Temperature:   22.0,
		Humidity:      65.0,
		Precipitation: 1.0,
		WindSpeed:     4.0,
		CachedAt:      time.Now(),
	}

	tokens := []string{"token1", "token2"}
	region := "울산시 남구"

	err := notifier.SendWeatherNotification(context.Background(), tokens, weatherData, region)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "all notification sends failed")
}

func TestSendWeatherNotification_ContextCancellation(t *testing.T) {
	logger := zap.NewNop()
	mockClient := &MockFCMClient{
		SendMulticastFunc: func(ctx context.Context, message *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
			return nil, context.Canceled
		},
	}

	notifier := NewFCMNotifierWithClient(mockClient, logger)

	weatherData := &entity.WeatherData{
		Temperature:   19.0,
		Humidity:      58.0,
		Precipitation: 0.5,
		WindSpeed:     2.0,
		CachedAt:      time.Now(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	tokens := []string{"token1"}
	region := "세종시"

	err := notifier.SendWeatherNotification(ctx, tokens, weatherData, region)
	assert.Error(t, err)
}

func TestSendWeatherNotification_BatchProcessing(t *testing.T) {
	logger := zap.NewNop()
	batchCallCount := 0
	mockClient := &MockFCMClient{
		SendMulticastFunc: func(ctx context.Context, message *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
			batchCallCount++
			responses := make([]*messaging.SendResponse, len(message.Tokens))
			for i := range responses {
				responses[i] = &messaging.SendResponse{
					Success: true,
				}
			}
			return &messaging.BatchResponse{
				SuccessCount: len(message.Tokens),
				FailureCount: 0,
				Responses:    responses,
			}, nil
		},
	}

	notifier := NewFCMNotifierWithClient(mockClient, logger)

	weatherData := &entity.WeatherData{
		Temperature:   16.0,
		Humidity:      72.0,
		Precipitation: 3.0,
		WindSpeed:     5.5,
		CachedAt:      time.Now(),
	}

	// Create 1200 tokens (should split into 3 batches: 500 + 500 + 200)
	tokens := make([]string, 1200)
	for i := range tokens {
		tokens[i] = "token_" + string(rune(i))
	}

	region := "제주시"

	err := notifier.SendWeatherNotification(context.Background(), tokens, weatherData, region)
	assert.NoError(t, err)
	assert.Equal(t, 3, batchCallCount) // Should have split into 3 batches
}

func TestSplitTokensIntoBatches(t *testing.T) {
	tests := []struct {
		name           string
		tokens         []string
		batchSize      int
		expectedBatches int
	}{
		{
			name:            "Empty tokens",
			tokens:          []string{},
			batchSize:       500,
			expectedBatches: 0,
		},
		{
			name:            "Single batch",
			tokens:          make([]string, 100),
			batchSize:       500,
			expectedBatches: 1,
		},
		{
			name:            "Exact multiple batches",
			tokens:          make([]string, 1000),
			batchSize:       500,
			expectedBatches: 2,
		},
		{
			name:            "Partial last batch",
			tokens:          make([]string, 750),
			batchSize:       500,
			expectedBatches: 2,
		},
		{
			name:            "Large number of tokens",
			tokens:          make([]string, 2550),
			batchSize:       500,
			expectedBatches: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batches := splitTokensIntoBatches(tt.tokens, tt.batchSize)
			assert.Equal(t, tt.expectedBatches, len(batches))

			// Verify total count
			totalTokens := 0
			for _, batch := range batches {
				totalTokens += len(batch)
				// Verify each batch doesn't exceed max size
				assert.LessOrEqual(t, len(batch), tt.batchSize)
			}
			assert.Equal(t, len(tt.tokens), totalTokens)
		})
	}
}
