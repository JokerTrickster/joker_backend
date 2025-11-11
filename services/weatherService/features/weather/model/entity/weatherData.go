package entity

import (
	"time"
)

// WeatherData represents weather information for a specific region
type WeatherData struct {
	Temperature   float64   `json:"temperature"`   // Temperature in Celsius
	Humidity      float64   `json:"humidity"`      // Humidity percentage
	Precipitation float64   `json:"precipitation"` // Precipitation in mm
	WindSpeed     float64   `json:"wind_speed"`    // Wind speed in m/s
	CachedAt      time.Time `json:"cached_at"`     // Time when data was cached
}
