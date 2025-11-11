package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
)

const (
	naverWeatherBaseURL = "https://search.naver.com/search.naver"
	defaultTimeout      = 10 * time.Second
	maxRetries          = 3
	initialBackoff      = 1 * time.Second
)

// HTTPClient interface for testability
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NaverWeatherCrawler handles weather data crawling with configurable dependencies
type NaverWeatherCrawler struct {
	client  HTTPClient
	baseURL string
	logger  *zap.Logger
	timeout time.Duration
	retries int
}

// NewNaverWeatherCrawler creates a new crawler with specified timeout and retry count
func NewNaverWeatherCrawler(timeout time.Duration, retries int) *NaverWeatherCrawler {
	logger, _ := zap.NewProduction()
	return &NaverWeatherCrawler{
		client: &http.Client{
			Timeout: timeout,
		},
		baseURL: naverWeatherBaseURL,
		logger:  logger,
		timeout: timeout,
		retries: retries,
	}
}

// NewNaverWeatherCrawlerWithClient creates a crawler with custom client (for testing)
func NewNaverWeatherCrawlerWithClient(client HTTPClient, baseURL string, logger *zap.Logger) *NaverWeatherCrawler {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}
	return &NaverWeatherCrawler{
		client:  client,
		baseURL: baseURL,
		logger:  logger,
		timeout: defaultTimeout,
		retries: maxRetries,
	}
}

// NewCrawler creates a new Crawler with default settings (backward compatibility)
func NewCrawler() *NaverWeatherCrawler {
	return NewNaverWeatherCrawler(defaultTimeout, maxRetries)
}

// NewCrawlerWithClient creates a Crawler with custom client and URL (backward compatibility)
func NewCrawlerWithClient(client HTTPClient, baseURL string) *NaverWeatherCrawler {
	return NewNaverWeatherCrawlerWithClient(client, baseURL, nil)
}

// CrawlWeather crawls weather data from Naver for the specified region
// It implements retry logic with exponential backoff (backward compatibility)
func CrawlWeather(ctx context.Context, region string) (*entity.WeatherData, error) {
	crawler := NewCrawler()
	return crawler.Fetch(ctx, region)
}

// Fetch crawls weather data from Naver for the specified region (new API)
func (c *NaverWeatherCrawler) Fetch(ctx context.Context, region string) (*entity.WeatherData, error) {
	if region == "" {
		return nil, fmt.Errorf("region cannot be empty")
	}

	startTime := time.Now()
	var lastErr error
	backoff := initialBackoff

	for attempt := 1; attempt <= c.retries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		c.logger.Info("attempting to fetch weather data",
			zap.String("region", region),
			zap.Int("attempt", attempt),
			zap.Int("max_retries", c.retries))

		weatherData, err := c.fetchWithRetry(ctx, region, attempt)
		if err == nil {
			duration := time.Since(startTime)
			c.logger.Info("successfully fetched weather data",
				zap.String("region", region),
				zap.Int("attempt", attempt),
				zap.Duration("duration", duration))
			return weatherData, nil
		}

		lastErr = err
		c.logger.Warn("fetch attempt failed",
			zap.String("region", region),
			zap.Int("attempt", attempt),
			zap.Error(err))

		// Don't sleep on the last attempt
		if attempt < c.retries {
			delay := backoff
			c.logger.Info("backing off before retry",
				zap.Duration("delay", delay))

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				// Exponential backoff: 1s, 2s, 4s
				backoff *= 2
			}
		}
	}

	c.logger.Error("all fetch attempts failed",
		zap.String("region", region),
		zap.Int("attempts", c.retries),
		zap.Error(lastErr))

	return nil, fmt.Errorf("failed after %d attempts: %w", c.retries, lastErr)
}

// CrawlWeather is the instance method for crawling (backward compatibility)
func (c *NaverWeatherCrawler) CrawlWeather(ctx context.Context, region string) (*entity.WeatherData, error) {
	return c.Fetch(ctx, region)
}

// fetchWithRetry performs a single fetch attempt
func (c *NaverWeatherCrawler) fetchWithRetry(ctx context.Context, region string, _ int) (*entity.WeatherData, error) {
	// Build URL with query parameter
	params := url.Values{}
	params.Add("query", fmt.Sprintf("날씨 %s", region))

	crawlURL := fmt.Sprintf("%s?%s", c.baseURL, params.Encode())

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, crawlURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse HTML and extract weather data
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	weatherData, err := c.parseWeatherData(doc)
	if err != nil {
		// Log snippet of HTML for debugging
		snippet := string(body)
		if len(snippet) > 500 {
			snippet = snippet[:500]
		}
		c.logger.Error("failed to parse weather data",
			zap.String("region", region),
			zap.String("html_snippet", snippet),
			zap.Error(err))
		return nil, fmt.Errorf("failed to parse weather data: %w", err)
	}

	weatherData.CachedAt = time.Now()
	return weatherData, nil
}

// attemptCrawl performs a single crawl attempt (backward compatibility)
func (c *NaverWeatherCrawler) attemptCrawl(ctx context.Context, region string) (*entity.WeatherData, error) {
	return c.fetchWithRetry(ctx, region, 1)
}

// parseWeatherData parses the HTML document and extracts weather information
func (c *NaverWeatherCrawler) parseWeatherData(doc *goquery.Document) (*entity.WeatherData, error) {
	weatherData := &entity.WeatherData{}
	var tempStr, humidStr, precipStr, windStr string

	// Extract temperature - looking for current temperature in weather module
	// Naver uses different selectors, we'll try multiple patterns
	temperatureSelectors := []string{
		".temperature_text strong",
		".weather_graphic .temperature_text",
		".temperature_text",
		".temperature .num",
	}

	for _, selector := range temperatureSelectors {
		temp := doc.Find(selector).First().Text()
		if temp != "" {
			tempStr = strings.TrimSpace(temp)
			break
		}
	}

	// Extract humidity, precipitation, and wind speed from info list
	// Naver displays these in a structured format
	doc.Find(".info_list .sort").Each(func(i int, s *goquery.Selection) {
		label := strings.TrimSpace(s.Find(".term").Text())
		value := strings.TrimSpace(s.Find(".desc").Text())

		switch {
		case strings.Contains(label, "습도"):
			humidStr = value
		case strings.Contains(label, "강수"):
			precipStr = value
		case strings.Contains(label, "바람"):
			windStr = value
		}
	})

	// Validate that we got at least temperature
	if tempStr == "" {
		return nil, fmt.Errorf("failed to extract temperature from HTML")
	}

	// Parse string values to float64
	var err error
	weatherData.Temperature, err = parseNumericValue(tempStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse temperature '%s': %w", tempStr, err)
	}

	// Parse optional fields (they may be empty)
	if humidStr != "" {
		weatherData.Humidity, _ = parseNumericValue(humidStr)
	}
	if precipStr != "" {
		weatherData.Precipitation, _ = parseNumericValue(precipStr)
	}
	if windStr != "" {
		weatherData.WindSpeed, _ = parseNumericValue(windStr)
	}

	return weatherData, nil
}

// parseNumericValue extracts numeric value from string like "15°", "60%", "2.5m/s", "0mm"
func parseNumericValue(value string) (float64, error) {
	// Use regex to extract numeric value (including decimals)
	re := regexp.MustCompile(`-?\d+\.?\d*`)
	match := re.FindString(value)
	if match == "" {
		return 0, fmt.Errorf("no numeric value found in '%s'", value)
	}
	return strconv.ParseFloat(match, 64)
}

// parseWeatherHTML parses the HTML content and extracts weather information (backward compatibility)
func parseWeatherHTML(htmlContent string) (*entity.WeatherData, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	crawler := NewCrawler()
	return crawler.parseWeatherData(doc)
}
