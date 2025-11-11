package crawler_test

import (
	"context"
	"fmt"
	"time"

	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/crawler"
)

// ExampleCrawlWeather demonstrates basic usage of the weather crawler
func ExampleCrawlWeather() {
	ctx := context.Background()

	// Crawl weather data for Seoul
	data, err := crawler.CrawlWeather(ctx, "서울")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Temperature: %.1f°C\n", data.Temperature)
	fmt.Printf("Humidity: %.0f%%\n", data.Humidity)
	fmt.Printf("Precipitation: %.1fmm\n", data.Precipitation)
	fmt.Printf("Wind Speed: %.1fm/s\n", data.WindSpeed)
	fmt.Printf("Cached At: %v\n", data.CachedAt.Format(time.RFC3339))
}

// ExampleNaverWeatherCrawler_Fetch demonstrates usage with custom configuration
func ExampleNaverWeatherCrawler_Fetch() {
	// Create a custom crawler instance
	customCrawler := crawler.NewNaverWeatherCrawler(10*time.Second, 3)

	// Use context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Crawl weather data for Busan
	data, err := customCrawler.Fetch(ctx, "부산")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Weather data retrieved successfully\n")
	fmt.Printf("Temperature: %.1f°C\n", data.Temperature)
}

// ExampleCrawlWeather_withRetry demonstrates retry behavior
func ExampleCrawlWeather_withRetry() {
	ctx := context.Background()

	// The crawler automatically retries up to 3 times with exponential backoff
	data, err := crawler.CrawlWeather(ctx, "인천")
	if err != nil {
		fmt.Printf("Failed after all retries: %v\n", err)
		return
	}

	fmt.Printf("Successfully retrieved weather for Incheon: %.1f°C\n", data.Temperature)
}
