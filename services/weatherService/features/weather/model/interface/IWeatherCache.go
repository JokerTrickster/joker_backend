package _interface

import (
	"context"
	"time"

	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
)

// IWeatherCache defines the interface for weather data caching operations
type IWeatherCache interface {
	// Get retrieves weather data from cache for a given region
	// Returns nil if cache miss or error
	Get(ctx context.Context, region string) (*entity.WeatherData, error)

	// Set stores weather data in cache with TTL
	Set(ctx context.Context, region string, data *entity.WeatherData) error

	// Delete removes weather data from cache
	Delete(ctx context.Context, region string) error

	// Close closes the cache connection
	Close() error

	// Ping checks if the cache connection is alive
	Ping(ctx context.Context) error

	// GetTTL returns the remaining TTL for a cached entry
	GetTTL(ctx context.Context, region string) (time.Duration, error)
}
