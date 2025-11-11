package _interface

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
)

// IWeatherCrawler defines the interface for weather data crawling
type IWeatherCrawler interface {
	// Fetch retrieves weather data for the specified region
	Fetch(ctx context.Context, region string) (*entity.WeatherData, error)
}
