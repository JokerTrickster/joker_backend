package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// HealthStatus represents the health status of the service
type HealthStatus struct {
	Status     string            `json:"status"`      // "ok" or "error"
	Timestamp  time.Time         `json:"timestamp"`   // Current time
	Version    string            `json:"version"`     // Application version
	Uptime     time.Duration     `json:"uptime"`      // Time since service started
	Components map[string]string `json:"components"`  // Component statuses
	ErrorCount int               `json:"error_count"` // Recent error count
}

// HealthChecker performs health checks on service components
type HealthChecker struct {
	db            *gorm.DB
	redis         *redis.Client
	logger        *zap.Logger
	version       string
	startTime     time.Time
	schedulerFunc func() bool // Function to check if scheduler is running
	errorCounter  *ErrorCounter
}

// ErrorCounter tracks recent errors with a sliding window
type ErrorCounter struct {
	mu      sync.RWMutex
	errors  []time.Time
	window  time.Duration
	maxRate int // Maximum errors per minute before alerting
}

// NewErrorCounter creates a new error counter
func NewErrorCounter(window time.Duration, maxRate int) *ErrorCounter {
	return &ErrorCounter{
		errors:  make([]time.Time, 0),
		window:  window,
		maxRate: maxRate,
	}
}

// Add records a new error
func (ec *ErrorCounter) Add() {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	now := time.Now()
	ec.errors = append(ec.errors, now)

	// Clean old errors outside the window
	ec.cleanup(now)
}

// Count returns the number of errors in the current window
func (ec *ErrorCounter) Count() int {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	ec.cleanup(time.Now())
	return len(ec.errors)
}

// cleanup removes errors outside the sliding window (must be called with lock held or read lock)
func (ec *ErrorCounter) cleanup(now time.Time) {
	cutoff := now.Add(-ec.window)
	validErrors := make([]time.Time, 0, len(ec.errors))

	for _, errTime := range ec.errors {
		if errTime.After(cutoff) {
			validErrors = append(validErrors, errTime)
		}
	}

	ec.errors = validErrors
}

// ShouldAlert returns true if error rate exceeds threshold
func (ec *ErrorCounter) ShouldAlert() bool {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	if len(ec.errors) == 0 {
		return false
	}

	// Calculate errors per minute
	errorsInWindow := len(ec.errors)
	minutesInWindow := ec.window.Minutes()
	errorsPerMinute := float64(errorsInWindow) / minutesInWindow

	return int(errorsPerMinute) > ec.maxRate
}

// NewHealthChecker creates a new HealthChecker instance
func NewHealthChecker(
	db *gorm.DB,
	redis *redis.Client,
	logger *zap.Logger,
	version string,
	schedulerFunc func() bool,
) *HealthChecker {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}

	return &HealthChecker{
		db:            db,
		redis:         redis,
		logger:        logger,
		version:       version,
		startTime:     time.Now(),
		schedulerFunc: schedulerFunc,
		errorCounter:  NewErrorCounter(5*time.Minute, 10), // 10 errors per minute threshold
	}
}

// RecordError records an error for tracking
func (h *HealthChecker) RecordError() {
	h.errorCounter.Add()

	if h.errorCounter.ShouldAlert() {
		h.logger.Warn("High error rate detected",
			zap.Int("error_count", h.errorCounter.Count()),
			zap.Duration("window", 5*time.Minute))
	}
}

// Check performs health checks on all components
func (h *HealthChecker) Check(ctx context.Context) *HealthStatus {
	components := make(map[string]string)
	overallHealthy := true

	// Check database
	if h.db != nil {
		sqlDB, err := h.db.DB()
		if err != nil {
			components["database"] = "error: " + err.Error()
			overallHealthy = false
		} else if err := sqlDB.PingContext(ctx); err != nil {
			components["database"] = "error: " + err.Error()
			overallHealthy = false
		} else {
			components["database"] = "ok"
		}
	} else {
		components["database"] = "not_configured"
	}

	// Check Redis
	if h.redis != nil {
		if err := h.redis.Ping(ctx).Err(); err != nil {
			components["redis"] = "error: " + err.Error()
			overallHealthy = false
		} else {
			components["redis"] = "ok"
		}
	} else {
		components["redis"] = "not_configured"
	}

	// Check scheduler
	if h.schedulerFunc != nil {
		if h.schedulerFunc() {
			components["scheduler"] = "running"
		} else {
			components["scheduler"] = "stopped"
			overallHealthy = false
		}
	} else {
		components["scheduler"] = "not_configured"
	}

	status := "ok"
	if !overallHealthy {
		status = "error"
	}

	return &HealthStatus{
		Status:     status,
		Timestamp:  time.Now(),
		Version:    h.version,
		Uptime:     time.Since(h.startTime),
		Components: components,
		ErrorCount: h.errorCounter.Count(),
	}
}

// Handler returns an HTTP handler for health checks
func (h *HealthChecker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		status := h.Check(ctx)

		w.Header().Set("Content-Type", "application/json")

		// Set HTTP status code based on health
		if status.Status == "ok" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		if err := json.NewEncoder(w).Encode(status); err != nil {
			h.logger.Error("Failed to encode health status",
				zap.Error(err))
		}
	}
}
