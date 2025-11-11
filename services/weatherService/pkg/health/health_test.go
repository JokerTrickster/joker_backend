package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNewHealthChecker(t *testing.T) {
	logger, _ := zap.NewProduction()

	// Create in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create miniredis
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	schedulerRunning := true
	checker := NewHealthChecker(
		db,
		redisClient,
		logger,
		"1.0.0",
		func() bool { return schedulerRunning },
	)

	assert.NotNil(t, checker)
	assert.Equal(t, "1.0.0", checker.version)
	assert.NotNil(t, checker.errorCounter)
}

func TestHealthChecker_Check(t *testing.T) {
	logger, _ := zap.NewProduction()

	// Create in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create miniredis
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	schedulerRunning := true
	checker := NewHealthChecker(
		db,
		redisClient,
		logger,
		"1.0.0",
		func() bool { return schedulerRunning },
	)

	t.Run("healthy system", func(t *testing.T) {
		status := checker.Check(context.Background())

		assert.Equal(t, "ok", status.Status)
		assert.Equal(t, "1.0.0", status.Version)
		assert.Equal(t, "ok", status.Components["database"])
		assert.Equal(t, "ok", status.Components["redis"])
		assert.Equal(t, "running", status.Components["scheduler"])
		assert.True(t, status.Uptime > 0)
	})

	t.Run("scheduler stopped", func(t *testing.T) {
		schedulerRunning = false
		status := checker.Check(context.Background())

		assert.Equal(t, "error", status.Status)
		assert.Equal(t, "stopped", status.Components["scheduler"])
	})

	t.Run("redis down", func(t *testing.T) {
		schedulerRunning = true
		mr.Close()

		status := checker.Check(context.Background())

		assert.Equal(t, "error", status.Status)
		assert.Contains(t, status.Components["redis"], "error")
	})
}

func TestHealthChecker_Handler(t *testing.T) {
	logger, _ := zap.NewProduction()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	checker := NewHealthChecker(
		db,
		redisClient,
		logger,
		"1.0.0",
		func() bool { return true },
	)

	handler := checker.Handler()

	t.Run("healthy response", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	})

	t.Run("unhealthy response", func(t *testing.T) {
		mr.Close()

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

func TestErrorCounter(t *testing.T) {
	t.Run("add and count errors", func(t *testing.T) {
		counter := NewErrorCounter(5*time.Minute, 10)

		assert.Equal(t, 0, counter.Count())

		counter.Add()
		counter.Add()
		counter.Add()

		assert.Equal(t, 3, counter.Count())
	})

	t.Run("cleanup old errors", func(t *testing.T) {
		counter := NewErrorCounter(100*time.Millisecond, 10)

		counter.Add()
		counter.Add()
		assert.Equal(t, 2, counter.Count())

		time.Sleep(150 * time.Millisecond)

		assert.Equal(t, 0, counter.Count())
	})

	t.Run("should alert on high error rate", func(t *testing.T) {
		counter := NewErrorCounter(1*time.Minute, 5)

		assert.False(t, counter.ShouldAlert())

		// Add 6 errors in 1 minute = 6 errors/minute > 5 threshold
		for i := 0; i < 6; i++ {
			counter.Add()
		}

		assert.True(t, counter.ShouldAlert())
	})

	t.Run("should not alert on low error rate", func(t *testing.T) {
		counter := NewErrorCounter(1*time.Minute, 10)

		for i := 0; i < 5; i++ {
			counter.Add()
		}

		assert.False(t, counter.ShouldAlert())
	})
}

func TestHealthChecker_RecordError(t *testing.T) {
	logger, _ := zap.NewProduction()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	checker := NewHealthChecker(
		db,
		redisClient,
		logger,
		"1.0.0",
		func() bool { return true },
	)

	status := checker.Check(context.Background())
	assert.Equal(t, 0, status.ErrorCount)

	checker.RecordError()
	checker.RecordError()

	status = checker.Check(context.Background())
	assert.Equal(t, 2, status.ErrorCount)
}
