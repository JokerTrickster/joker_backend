package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	once sync.Once

	// Crawler metrics
	crawlRequestsTotal *prometheus.CounterVec
	crawlDuration      *prometheus.HistogramVec
	crawlErrorsTotal   *prometheus.CounterVec

	// Cache metrics
	cacheHitsTotal   prometheus.Counter
	cacheMissesTotal prometheus.Counter
	cacheErrorsTotal *prometheus.CounterVec

	// FCM metrics
	fcmNotificationsSentTotal *prometheus.CounterVec
	fcmSendDuration           prometheus.Histogram
	fcmBatchSize              prometheus.Histogram
	fcmErrorsTotal            *prometheus.CounterVec

	// Scheduler metrics
	schedulerTicksTotal             prometheus.Counter
	schedulerAlarmsProcessedTotal   *prometheus.CounterVec
	schedulerProcessingDuration     prometheus.Histogram
	schedulerConsecutiveFailures    prometheus.Gauge
	schedulerErrorRate              prometheus.Gauge
)

// InitMetrics initializes all Prometheus metrics
// Should be called once at application startup
func InitMetrics() {
	once.Do(func() {
		// Crawler metrics
		crawlRequestsTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "weather_crawl_requests_total",
				Help: "Total number of weather crawl requests",
			},
			[]string{"region", "status"},
		)

		crawlDuration = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "weather_crawl_duration_seconds",
				Help:    "Duration of weather crawl requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"region"},
		)

		crawlErrorsTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "weather_crawl_errors_total",
				Help: "Total number of weather crawl errors",
			},
			[]string{"region", "error_type"},
		)

		// Cache metrics
		cacheHitsTotal = promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "weather_cache_hits_total",
				Help: "Total number of cache hits",
			},
		)

		cacheMissesTotal = promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "weather_cache_misses_total",
				Help: "Total number of cache misses",
			},
		)

		cacheErrorsTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "weather_cache_errors_total",
				Help: "Total number of cache operation errors",
			},
			[]string{"operation"},
		)

		// FCM metrics
		fcmNotificationsSentTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "fcm_notifications_sent_total",
				Help: "Total number of FCM notifications sent",
			},
			[]string{"status"},
		)

		fcmSendDuration = promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "fcm_send_duration_seconds",
				Help:    "Duration of FCM send operations in seconds",
				Buckets: prometheus.DefBuckets,
			},
		)

		fcmBatchSize = promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "fcm_batch_size",
				Help:    "Size of FCM notification batches",
				Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500},
			},
		)

		fcmErrorsTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "fcm_errors_total",
				Help: "Total number of FCM errors",
			},
			[]string{"error_type"},
		)

		// Scheduler metrics
		schedulerTicksTotal = promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "scheduler_ticks_total",
				Help: "Total number of scheduler ticks",
			},
		)

		schedulerAlarmsProcessedTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "scheduler_alarms_processed_total",
				Help: "Total number of alarms processed",
			},
			[]string{"status"},
		)

		schedulerProcessingDuration = promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "scheduler_processing_duration_seconds",
				Help:    "Duration of scheduler processing in seconds",
				Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
			},
		)

		schedulerConsecutiveFailures = promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "scheduler_consecutive_failures",
				Help: "Number of consecutive scheduler failures",
			},
		)

		schedulerErrorRate = promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "scheduler_error_rate",
				Help: "Scheduler error rate (errors per minute)",
			},
		)
	})
}

// RecordCrawlRequest records a weather crawl request with its status and duration
func RecordCrawlRequest(region string, status string, duration time.Duration) {
	crawlRequestsTotal.WithLabelValues(region, status).Inc()
	crawlDuration.WithLabelValues(region).Observe(duration.Seconds())
}

// RecordCrawlError records a weather crawl error
func RecordCrawlError(region string, errorType string) {
	crawlErrorsTotal.WithLabelValues(region, errorType).Inc()
}

// RecordCacheHit records a cache hit
func RecordCacheHit() {
	cacheHitsTotal.Inc()
}

// RecordCacheMiss records a cache miss
func RecordCacheMiss() {
	cacheMissesTotal.Inc()
}

// RecordCacheError records a cache operation error
func RecordCacheError(operation string) {
	cacheErrorsTotal.WithLabelValues(operation).Inc()
}

// RecordFCMSent records an FCM notification send operation
func RecordFCMSent(status string, duration time.Duration, batchSize int) {
	fcmNotificationsSentTotal.WithLabelValues(status).Inc()
	fcmSendDuration.Observe(duration.Seconds())
	fcmBatchSize.Observe(float64(batchSize))
}

// RecordFCMError records an FCM error
func RecordFCMError(errorType string) {
	fcmErrorsTotal.WithLabelValues(errorType).Inc()
}

// RecordSchedulerTick records a scheduler tick with processed alarms count and duration
func RecordSchedulerTick(alarmsProcessed int, duration time.Duration) {
	schedulerTicksTotal.Inc()
	schedulerProcessingDuration.Observe(duration.Seconds())

	if alarmsProcessed > 0 {
		schedulerAlarmsProcessedTotal.WithLabelValues("success").Add(float64(alarmsProcessed))
	}
}

// RecordSchedulerAlarmStatus records the status of an individual alarm processing
func RecordSchedulerAlarmStatus(status string) {
	schedulerAlarmsProcessedTotal.WithLabelValues(status).Inc()
}

// SetSchedulerConsecutiveFailures sets the current consecutive failure count
func SetSchedulerConsecutiveFailures(count int) {
	schedulerConsecutiveFailures.Set(float64(count))
}

// SetSchedulerErrorRate sets the current error rate (errors per minute)
func SetSchedulerErrorRate(rate float64) {
	schedulerErrorRate.Set(rate)
}
