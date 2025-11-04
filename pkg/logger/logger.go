package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

// Init initializes the structured logger with the specified level
func Init(level string) {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapLevel),
		Development:      os.Getenv("ENV") != "production",
		Encoding:         "json",
		EncoderConfig:    getEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	var err error
	logger, err = config.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
}

func getEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// GetLogger returns the underlying zap logger for advanced usage
func GetLogger() *zap.Logger {
	if logger == nil {
		Init("info")
	}
	return logger
}

// Sync flushes any buffered log entries
func Sync() {
	if logger != nil {
		_ = logger.Sync()
	}
}

// Debug logs a debug message with optional fields
func Debug(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Debug(msg, fields...)
	}
}

// Info logs an info message with optional fields
func Info(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Info(msg, fields...)
	}
}

// Warn logs a warning message with optional fields
func Warn(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Warn(msg, fields...)
	}
}

// Error logs an error message with optional fields
func Error(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Error(msg, fields...)
	}
}

// Fatal logs a fatal message with optional fields and exits
func Fatal(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Fatal(msg, fields...)
	}
}

// With creates a child logger with additional fields
func With(fields ...zap.Field) *zap.Logger {
	if logger != nil {
		return logger.With(fields...)
	}
	return nil
}
