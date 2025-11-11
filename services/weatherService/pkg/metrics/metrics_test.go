package metrics

import (
	"testing"
	"time"
)

func TestInitMetrics(t *testing.T) {
	// Should not panic when called
	InitMetrics()

	// Should be idempotent
	InitMetrics()
}

func TestRecordCrawlRequest(t *testing.T) {
	InitMetrics()

	tests := []struct {
		name     string
		region   string
		status   string
		duration time.Duration
	}{
		{"success request", "서울", "success", 1 * time.Second},
		{"failed request", "부산", "failure", 500 * time.Millisecond},
		{"timeout request", "대구", "timeout", 10 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			RecordCrawlRequest(tt.region, tt.status, tt.duration)
		})
	}
}

func TestRecordCrawlError(t *testing.T) {
	InitMetrics()

	tests := []struct {
		name      string
		region    string
		errorType string
	}{
		{"network error", "서울", "network_error"},
		{"parse error", "부산", "parse_error"},
		{"timeout error", "대구", "timeout"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RecordCrawlError(tt.region, tt.errorType)
		})
	}
}

func TestRecordCacheOperations(t *testing.T) {
	InitMetrics()

	t.Run("cache hit", func(t *testing.T) {
		RecordCacheHit()
	})

	t.Run("cache miss", func(t *testing.T) {
		RecordCacheMiss()
	})

	t.Run("cache error", func(t *testing.T) {
		RecordCacheError("get")
		RecordCacheError("set")
		RecordCacheError("delete")
	})
}

func TestRecordFCMOperations(t *testing.T) {
	InitMetrics()

	tests := []struct {
		name      string
		status    string
		duration  time.Duration
		batchSize int
	}{
		{"successful send", "success", 500 * time.Millisecond, 10},
		{"failed send", "failure", 1 * time.Second, 5},
		{"large batch", "success", 2 * time.Second, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RecordFCMSent(tt.status, tt.duration, tt.batchSize)
		})
	}

	t.Run("fcm error", func(t *testing.T) {
		RecordFCMError("invalid_token")
		RecordFCMError("network_error")
	})
}

func TestRecordSchedulerOperations(t *testing.T) {
	InitMetrics()

	t.Run("scheduler tick", func(t *testing.T) {
		RecordSchedulerTick(10, 5*time.Second)
		RecordSchedulerTick(0, 100*time.Millisecond)
	})

	t.Run("alarm status", func(t *testing.T) {
		RecordSchedulerAlarmStatus("success")
		RecordSchedulerAlarmStatus("failed")
	})

	t.Run("consecutive failures", func(t *testing.T) {
		SetSchedulerConsecutiveFailures(0)
		SetSchedulerConsecutiveFailures(3)
		SetSchedulerConsecutiveFailures(5)
	})

	t.Run("error rate", func(t *testing.T) {
		SetSchedulerErrorRate(0.5)
		SetSchedulerErrorRate(5.0)
		SetSchedulerErrorRate(15.0)
	})
}
