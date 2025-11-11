package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsEndpointIntegration(t *testing.T) {
	// Initialize metrics
	InitMetrics()

	// Record some sample metrics
	RecordCrawlRequest("서울", "success", 1*time.Second)
	RecordCrawlRequest("부산", "failure", 2*time.Second)
	RecordCacheHit()
	RecordCacheMiss()
	RecordFCMSent("success", 500*time.Millisecond, 10)
	RecordSchedulerTick(5, 3*time.Second)

	// Create test server
	handler := promhttp.Handler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/plain")

	bodyStr := string(body)

	// Verify crawler metrics
	assert.Contains(t, bodyStr, "weather_crawl_requests_total")
	assert.Contains(t, bodyStr, "weather_crawl_duration_seconds")
	assert.Contains(t, bodyStr, `region="서울"`)
	assert.Contains(t, bodyStr, `region="부산"`)
	assert.Contains(t, bodyStr, `status="success"`)
	assert.Contains(t, bodyStr, `status="failure"`)

	// Verify cache metrics
	assert.Contains(t, bodyStr, "weather_cache_hits_total")
	assert.Contains(t, bodyStr, "weather_cache_misses_total")

	// Verify FCM metrics
	assert.Contains(t, bodyStr, "fcm_notifications_sent_total")
	assert.Contains(t, bodyStr, "fcm_send_duration_seconds")
	assert.Contains(t, bodyStr, "fcm_batch_size")

	// Verify scheduler metrics
	assert.Contains(t, bodyStr, "scheduler_ticks_total")
	assert.Contains(t, bodyStr, "scheduler_processing_duration_seconds")

	// Verify Go runtime metrics (automatically included)
	assert.Contains(t, bodyStr, "go_goroutines")
	assert.Contains(t, bodyStr, "go_memstats_alloc_bytes")
}

func TestMetricsFormat(t *testing.T) {
	InitMetrics()

	// Record metrics
	RecordCrawlRequest("test_region", "success", 100*time.Millisecond)

	// Get metrics
	handler := promhttp.Handler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	body, _ := io.ReadAll(w.Result().Body)
	bodyStr := string(body)

	// Verify Prometheus format
	lines := strings.Split(bodyStr, "\n")
	hasHelpLine := false
	hasTypeLine := false
	hasMetricLine := false

	for _, line := range lines {
		if strings.HasPrefix(line, "# HELP weather_crawl_requests_total") {
			hasHelpLine = true
		}
		if strings.HasPrefix(line, "# TYPE weather_crawl_requests_total counter") {
			hasTypeLine = true
		}
		if strings.Contains(line, "weather_crawl_requests_total{") {
			hasMetricLine = true
		}
	}

	assert.True(t, hasHelpLine, "Should have HELP line")
	assert.True(t, hasTypeLine, "Should have TYPE line")
	assert.True(t, hasMetricLine, "Should have metric line")
}

func TestMetricsConcurrency(t *testing.T) {
	InitMetrics()

	// Test concurrent metric recording
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				RecordCacheHit()
				RecordCacheMiss()
				RecordSchedulerTick(1, 100*time.Millisecond)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify metrics endpoint still works
	handler := promhttp.Handler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
}

func TestAllMetricsRecorded(t *testing.T) {
	InitMetrics()

	// Record at least one of each metric type
	RecordCrawlRequest("region1", "success", 1*time.Second)
	RecordCrawlError("region1", "timeout")
	RecordCacheHit()
	RecordCacheMiss()
	RecordCacheError("get")
	RecordFCMSent("success", 500*time.Millisecond, 10)
	RecordFCMError("invalid_token")
	RecordSchedulerTick(5, 2*time.Second)
	RecordSchedulerAlarmStatus("success")
	SetSchedulerConsecutiveFailures(3)
	SetSchedulerErrorRate(5.5)

	// Get metrics
	handler := promhttp.Handler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	body, _ := io.ReadAll(w.Result().Body)
	bodyStr := string(body)

	// Verify all metric families are present
	expectedMetrics := []string{
		"weather_crawl_requests_total",
		"weather_crawl_duration_seconds",
		"weather_crawl_errors_total",
		"weather_cache_hits_total",
		"weather_cache_misses_total",
		"weather_cache_errors_total",
		"fcm_notifications_sent_total",
		"fcm_send_duration_seconds",
		"fcm_batch_size",
		"fcm_errors_total",
		"scheduler_ticks_total",
		"scheduler_alarms_processed_total",
		"scheduler_processing_duration_seconds",
		"scheduler_consecutive_failures",
		"scheduler_error_rate",
	}

	for _, metric := range expectedMetrics {
		assert.Contains(t, bodyStr, metric, "Should contain metric: %s", metric)
	}
}
