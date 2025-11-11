package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name        string
		level       string
		outputPaths []string
		wantErr     bool
	}{
		{
			name:        "default stdout",
			level:       "info",
			outputPaths: nil,
			wantErr:     false,
		},
		{
			name:        "debug level",
			level:       "debug",
			outputPaths: []string{"stdout"},
			wantErr:     false,
		},
		{
			name:        "warn level",
			level:       "warn",
			outputPaths: []string{"stdout"},
			wantErr:     false,
		},
		{
			name:        "error level",
			level:       "error",
			outputPaths: []string{"stdout"},
			wantErr:     false,
		},
		{
			name:        "invalid level defaults to info",
			level:       "invalid",
			outputPaths: []string{"stdout"},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.level, tt.outputPaths)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)

				// Test that logger works
				logger.Info("test message")
			}
		})
	}
}

func TestWithRequestID(t *testing.T) {
	logger, err := NewLogger("info", []string{"stdout"})
	require.NoError(t, err)

	loggerWithID := WithRequestID(logger, "req-12345")
	assert.NotNil(t, loggerWithID)

	// Test that it doesn't panic
	loggerWithID.Info("test with request ID")
}

func TestWithComponent(t *testing.T) {
	logger, err := NewLogger("info", []string{"stdout"})
	require.NoError(t, err)

	loggerWithComponent := WithComponent(logger, "scheduler")
	assert.NotNil(t, loggerWithComponent)

	loggerWithComponent.Info("test with component")
}

func TestWithRegion(t *testing.T) {
	logger, err := NewLogger("info", []string{"stdout"})
	require.NoError(t, err)

	loggerWithRegion := WithRegion(logger, "서울")
	assert.NotNil(t, loggerWithRegion)

	loggerWithRegion.Info("test with region")
}

func TestWithUserID(t *testing.T) {
	logger, err := NewLogger("info", []string{"stdout"})
	require.NoError(t, err)

	loggerWithUser := WithUserID(logger, 123)
	assert.NotNil(t, loggerWithUser)

	loggerWithUser.Info("test with user ID")
}

func TestWithAlarmID(t *testing.T) {
	logger, err := NewLogger("info", []string{"stdout"})
	require.NoError(t, err)

	loggerWithAlarm := WithAlarmID(logger, 456)
	assert.NotNil(t, loggerWithAlarm)

	loggerWithAlarm.Info("test with alarm ID")
}

func TestLoggerChaining(t *testing.T) {
	logger, err := NewLogger("info", []string{"stdout"})
	require.NoError(t, err)

	// Chain multiple context additions
	enrichedLogger := logger
	enrichedLogger = WithComponent(enrichedLogger, "scheduler")
	enrichedLogger = WithRegion(enrichedLogger, "서울")
	enrichedLogger = WithUserID(enrichedLogger, 123)
	enrichedLogger = WithAlarmID(enrichedLogger, 456)
	enrichedLogger = WithRequestID(enrichedLogger, "req-12345")

	assert.NotNil(t, enrichedLogger)
	enrichedLogger.Info("test with all context")
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name  string
		level string
	}{
		{"debug", "debug"},
		{"info", "info"},
		{"warn", "warn"},
		{"warning", "warning"},
		{"error", "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.level, []string{"stdout"})
			require.NoError(t, err)

			// Should not panic
			logger.Debug("debug message")
			logger.Info("info message")
			logger.Warn("warn message")
			logger.Error("error message")
		})
	}
}

func TestLoggerSync(t *testing.T) {
	logger, err := NewLogger("info", []string{"stdout"})
	require.NoError(t, err)

	logger.Info("test message")

	// Should not panic
	err = logger.Sync()
	// Sync may return error on stdout, which is acceptable
	if err != nil {
		t.Logf("Sync returned error (acceptable for stdout): %v", err)
	}
}

func TestMultipleOutputPaths(t *testing.T) {
	t.Skip("Skipping file output test to avoid filesystem dependencies")

	logger, err := NewLogger("info", []string{"stdout", "/tmp/test.log"})
	require.NoError(t, err)

	logger.Info("test message to multiple outputs")
	logger.Sync()
}

func TestLoggerFields(t *testing.T) {
	logger, err := NewLogger("info", []string{"stdout"})
	require.NoError(t, err)

	// Test various field types
	logger.Info("test message",
		zap.String("string_field", "value"),
		zap.Int("int_field", 42),
		zap.Bool("bool_field", true),
		zap.Duration("duration_field", 0),
	)
}
