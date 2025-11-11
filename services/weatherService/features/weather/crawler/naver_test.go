package crawler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockNaverWeatherHTML returns a realistic mock of Naver's weather page HTML
func mockNaverWeatherHTML() string {
	return `
<!DOCTYPE html>
<html>
<head><title>날씨 서울</title></head>
<body>
	<div class="weather_area">
		<div class="temperature_text">
			<strong>15°</strong>
		</div>
		<div class="info_list">
			<div class="sort">
				<span class="term">습도</span>
				<span class="desc">60%</span>
			</div>
			<div class="sort">
				<span class="term">강수량</span>
				<span class="desc">0mm</span>
			</div>
			<div class="sort">
				<span class="term">바람</span>
				<span class="desc">2.5m/s</span>
			</div>
		</div>
	</div>
</body>
</html>
`
}

// mockIncompleteWeatherHTML returns HTML with missing weather data
func mockIncompleteWeatherHTML() string {
	return `
<!DOCTYPE html>
<html>
<head><title>날씨</title></head>
<body>
	<div class="weather_area">
		<div class="info_list">
			<div class="sort">
				<span class="term">습도</span>
				<span class="desc">60%</span>
			</div>
		</div>
	</div>
</body>
</html>
`
}

func TestCrawlWeather_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request parameters
		query := r.URL.Query().Get("query")
		assert.Contains(t, query, "날씨")
		assert.Contains(t, query, "서울")

		// Return mock HTML
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockNaverWeatherHTML()))
	}))
	defer server.Close()

	// Create crawler with test server
	logger := zap.NewNop()
	crawler := NewNaverWeatherCrawlerWithClient(server.Client(), server.URL, logger)

	// Test with context
	ctx := context.Background()
	result, err := crawler.CrawlWeather(ctx, "서울")

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 15.0, result.Temperature)
	assert.Equal(t, 60.0, result.Humidity)
	assert.Equal(t, 0.0, result.Precipitation)
	assert.Equal(t, 2.5, result.WindSpeed)
	assert.False(t, result.CachedAt.IsZero())
}

func TestCrawlWeather_EmptyRegion(t *testing.T) {
	ctx := context.Background()
	result, err := CrawlWeather(ctx, "")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "region cannot be empty")
}

func TestCrawlWeather_ContextCancellation(t *testing.T) {
	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := CrawlWeather(ctx, "서울")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, context.Canceled, err)
}

func TestCrawlWeather_ContextTimeout(t *testing.T) {
	// Create slow mock server that will timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := CrawlWeather(ctx, "서울")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAttemptCrawl_HTTPError(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	logger := zap.NewNop()
	crawler := NewNaverWeatherCrawlerWithClient(server.Client(), server.URL, logger)
	ctx := context.Background()
	result, err := crawler.attemptCrawl(ctx, "서울")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unexpected status code: 500")
}

func TestParseWeatherHTML_Success(t *testing.T) {
	htmlContent := mockNaverWeatherHTML()

	result, err := parseWeatherHTML(htmlContent)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 15.0, result.Temperature)
	assert.Equal(t, 60.0, result.Humidity)
	assert.Equal(t, 0.0, result.Precipitation)
	assert.Equal(t, 2.5, result.WindSpeed)
}

func TestParseWeatherHTML_MissingTemperature(t *testing.T) {
	htmlContent := mockIncompleteWeatherHTML()

	result, err := parseWeatherHTML(htmlContent)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to extract temperature")
}

func TestParseWeatherHTML_InvalidHTML(t *testing.T) {
	// Test with completely invalid HTML
	htmlContent := "<<<invalid html>>>"

	result, err := parseWeatherHTML(htmlContent)

	// goquery is lenient and will parse even invalid HTML
	// but we should still fail if temperature is not found
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestParseWeatherHTML_EmptyHTML(t *testing.T) {
	htmlContent := ""

	result, err := parseWeatherHTML(htmlContent)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to extract temperature")
}

func TestParseWeatherHTML_AlternativeSelectors(t *testing.T) {
	// Test with different HTML structure using alternative selectors
	htmlContent := `
<!DOCTYPE html>
<html>
<body>
	<div class="temperature">
		<span class="num">22°</span>
	</div>
	<div class="info_list">
		<div class="sort">
			<span class="term">습도</span>
			<span class="desc">75%</span>
		</div>
	</div>
</body>
</html>
`

	result, err := parseWeatherHTML(htmlContent)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 22.0, result.Temperature)
	assert.Equal(t, 75.0, result.Humidity)
}

func TestParseWeatherHTML_WithWhitespace(t *testing.T) {
	// Test HTML with extra whitespace
	htmlContent := `
<!DOCTYPE html>
<html>
<body>
	<div class="temperature_text">
		<strong>  18°  </strong>
	</div>
	<div class="info_list">
		<div class="sort">
			<span class="term">  습도  </span>
			<span class="desc">  55%  </span>
		</div>
	</div>
</body>
</html>
`

	result, err := parseWeatherHTML(htmlContent)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 18.0, result.Temperature)
	assert.Equal(t, 55.0, result.Humidity)
}

func TestCrawlWeather_RetryLogic(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		// Fail first 2 attempts, succeed on 3rd
		if attemptCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockNaverWeatherHTML()))
	}))
	defer server.Close()

	logger := zap.NewNop()
	crawler := NewNaverWeatherCrawlerWithClient(server.Client(), server.URL, logger)
	ctx := context.Background()

	result, err := crawler.CrawlWeather(ctx, "서울")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 3, attemptCount, "Should retry until success")
	assert.Equal(t, 15.0, result.Temperature)
}

func TestCrawlWeather_MaxRetriesExceeded(t *testing.T) {
	// Create mock server that always fails
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	logger := zap.NewNop()
	crawler := NewNaverWeatherCrawlerWithClient(server.Client(), server.URL, logger)
	ctx := context.Background()

	start := time.Now()
	result, err := crawler.CrawlWeather(ctx, "서울")
	duration := time.Since(start)

	// Should fail after 3 attempts
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed after 3 attempts")
	assert.Equal(t, 3, attemptCount)

	// Should take at least 3 seconds (1s + 2s backoff)
	assert.GreaterOrEqual(t, duration.Seconds(), 3.0)
}

func TestAttemptCrawl_RequestHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify User-Agent header is set
		userAgent := r.Header.Get("User-Agent")
		assert.NotEmpty(t, userAgent)
		assert.Contains(t, userAgent, "Mozilla")

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockNaverWeatherHTML()))
	}))
	defer server.Close()

	logger := zap.NewNop()
	crawler := NewNaverWeatherCrawlerWithClient(server.Client(), server.URL, logger)
	ctx := context.Background()
	result, err := crawler.attemptCrawl(ctx, "서울")

	require.NoError(t, err)
	require.NotNil(t, result)
}

func BenchmarkParseWeatherHTML(b *testing.B) {
	htmlContent := mockNaverWeatherHTML()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseWeatherHTML(htmlContent)
	}
}

func BenchmarkCrawlWeather(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockNaverWeatherHTML()))
	}))
	defer server.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = CrawlWeather(ctx, "서울")
	}
}

func TestParseWeatherHTML_PartialData(t *testing.T) {
	// Test when only temperature and humidity are available
	htmlContent := `
<!DOCTYPE html>
<html>
<body>
	<div class="temperature_text">
		<strong>10°</strong>
	</div>
	<div class="info_list">
		<div class="sort">
			<span class="term">습도</span>
			<span class="desc">80%</span>
		</div>
	</div>
</body>
</html>
`

	result, err := parseWeatherHTML(htmlContent)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 10.0, result.Temperature)
	assert.Equal(t, 80.0, result.Humidity)
	assert.Equal(t, 0.0, result.Precipitation)
	assert.Equal(t, 0.0, result.WindSpeed)
}

func TestCrawlWeather_DifferentRegions(t *testing.T) {
	testCases := []struct {
		name   string
		region string
	}{
		{"Seoul", "서울"},
		{"Busan", "부산"},
		{"Incheon", "인천"},
		{"Daegu", "대구"},
		{"MultiWord", "서울 강남구"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				query := r.URL.Query().Get("query")
				assert.Contains(t, query, tc.region)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(mockNaverWeatherHTML()))
			}))
			defer server.Close()

			logger := zap.NewNop()
			crawler := NewNaverWeatherCrawlerWithClient(server.Client(), server.URL, logger)
			ctx := context.Background()
			result, err := crawler.CrawlWeather(ctx, tc.region)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, 15.0, result.Temperature)
		})
	}
}

func TestAttemptCrawl_URLEncoding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify proper URL encoding
		query := r.URL.Query().Get("query")
		// Should contain encoded Korean characters
		assert.NotEmpty(t, query)
		assert.Contains(t, query, "서울 강남구")

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockNaverWeatherHTML()))
	}))
	defer server.Close()

	logger := zap.NewNop()
	crawler := NewNaverWeatherCrawlerWithClient(server.Client(), server.URL, logger)
	ctx := context.Background()
	region := "서울 강남구"
	result, err := crawler.attemptCrawl(ctx, region)

	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestParseWeatherHTML_RobustLabelMatching(t *testing.T) {
	// Test with variations in label text
	htmlContent := `
<!DOCTYPE html>
<html>
<body>
	<div class="temperature_text">
		<strong>16°</strong>
	</div>
	<div class="info_list">
		<div class="sort">
			<span class="term">현재습도</span>
			<span class="desc">70%</span>
		</div>
		<div class="sort">
			<span class="term">강수확률</span>
			<span class="desc">30%</span>
		</div>
		<div class="sort">
			<span class="term">바람속도</span>
			<span class="desc">1.8m/s</span>
		</div>
	</div>
</body>
</html>
`

	result, err := parseWeatherHTML(htmlContent)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 16.0, result.Temperature)
	// Should match partial label text using Contains
	assert.Equal(t, 70.0, result.Humidity)
	assert.Equal(t, 30.0, result.Precipitation)
	assert.Equal(t, 1.8, result.WindSpeed)
}

func TestConstants(t *testing.T) {
	// Test that constants are properly defined
	assert.Equal(t, "https://search.naver.com/search.naver", naverWeatherBaseURL)
	assert.Equal(t, 10*time.Second, defaultTimeout)
	assert.Equal(t, 3, maxRetries)
	assert.Equal(t, 1*time.Second, initialBackoff)
}

func TestCrawlWeather_ThreadSafety(t *testing.T) {
	// Test concurrent crawling
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockNaverWeatherHTML()))
	}))
	defer server.Close()

	logger := zap.NewNop()
	crawler := NewNaverWeatherCrawlerWithClient(server.Client(), server.URL, logger)
	ctx := context.Background()
	concurrentCalls := 10

	done := make(chan error, concurrentCalls)

	for i := 0; i < concurrentCalls; i++ {
		go func() {
			_, err := crawler.CrawlWeather(ctx, "서울")
			done <- err
		}()
	}

	// Wait for all goroutines to complete and verify no errors
	for i := 0; i < concurrentCalls; i++ {
		err := <-done
		assert.NoError(t, err)
	}
}

func TestParseWeatherHTML_SpecialCharacters(t *testing.T) {
	htmlContent := `
<!DOCTYPE html>
<html>
<body>
	<div class="temperature_text">
		<strong>15.5°</strong>
	</div>
	<div class="info_list">
		<div class="sort">
			<span class="term">습도</span>
			<span class="desc">60%</span>
		</div>
	</div>
</body>
</html>
`

	result, err := parseWeatherHTML(htmlContent)

	require.NoError(t, err)
	assert.Equal(t, 15.5, result.Temperature)
}

// Test parseNumericValue function
func TestParseNumericValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
		wantErr  bool
	}{
		{"Temperature", "15°", 15.0, false},
		{"Decimal Temperature", "15.5°", 15.5, false},
		{"Humidity", "60%", 60.0, false},
		{"Precipitation", "0mm", 0.0, false},
		{"Wind Speed", "2.5m/s", 2.5, false},
		{"Negative", "-5°", -5.0, false},
		{"No Number", "N/A", 0.0, true},
		{"Only Text", "sunny", 0.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseNumericValue(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
