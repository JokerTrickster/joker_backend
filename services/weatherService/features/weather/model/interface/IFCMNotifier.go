package _interface

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
)

// IFCMNotifier defines the interface for FCM push notification operations
type IFCMNotifier interface {
	// SendWeatherNotification sends weather notification to specified FCM tokens
	// Returns error if notification fails to send to any token
	SendWeatherNotification(ctx context.Context, tokens []string, data *entity.WeatherData, region string) error
}
