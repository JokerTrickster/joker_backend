package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// NewLogger creates a new zap logger with specified level and output paths
func NewLogger(level string, outputPaths []string) (*zap.Logger, error) {
	if len(outputPaths) == 0 {
		outputPaths = []string{"stdout"}
	}

	// Parse log level
	var zapLevel zap.AtomicLevel
	switch level {
	case "debug":
		zapLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn", "warning":
		zapLevel = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapLevel = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		zapLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	// Create logger configuration
	config := zap.Config{
		Level:            zapLevel,
		Development:      false,
		Encoding:         "json",
		OutputPaths:      outputPaths,
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig:    zap.NewProductionEncoderConfig(),
	}

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return logger, nil
}

// WithRequestID adds a request ID field to the logger
func WithRequestID(logger *zap.Logger, requestID string) *zap.Logger {
	return logger.With(zap.String("request_id", requestID))
}

// WithComponent adds a component name field to the logger
func WithComponent(logger *zap.Logger, component string) *zap.Logger {
	return logger.With(zap.String("component", component))
}

// WithRegion adds a region field to the logger
func WithRegion(logger *zap.Logger, region string) *zap.Logger {
	return logger.With(zap.String("region", region))
}

// WithUserID adds a user ID field to the logger
func WithUserID(logger *zap.Logger, userID int) *zap.Logger {
	return logger.With(zap.Int("user_id", userID))
}

// WithAlarmID adds an alarm ID field to the logger
func WithAlarmID(logger *zap.Logger, alarmID int) *zap.Logger {
	return logger.With(zap.Int("alarm_id", alarmID))
}
