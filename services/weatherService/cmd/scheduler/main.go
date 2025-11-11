package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/cache"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/crawler"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/notifier"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/repository"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/scheduler"
	"github.com/JokerTrickster/joker_backend/services/weatherService/pkg/health"
	"github.com/JokerTrickster/joker_backend/services/weatherService/pkg/metrics"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Build information (set via ldflags)
var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

func main() {
	// Initialize logger
	logger, err := initLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting Weather Scheduler Service",
		zap.String("version", version),
		zap.String("build_time", buildTime),
		zap.String("git_commit", gitCommit))

	// Initialize Prometheus metrics
	metrics.InitMetrics()
	logger.Info("Metrics initialized")

	// Load configuration from environment
	config := loadConfig()

	// Validate critical configuration
	if err := validateConfig(config); err != nil {
		logger.Fatal("Invalid configuration", zap.Error(err))
	}

	// Initialize database connection
	db, err := initDatabase(config, logger)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Initialize Redis client for health checks
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
		Password: config.RedisPassword,
		DB:       0,
	})
	defer redisClient.Close()

	// Test Redis connection
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	logger.Info("Redis connection established")

	// Initialize Redis cache
	weatherCache, err := cache.NewWeatherCache(
		fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
		config.RedisPassword,
		logger,
	)
	if err != nil {
		logger.Fatal("Failed to initialize Redis cache", zap.Error(err))
	}
	defer weatherCache.Close()

	// Initialize crawler
	weatherCrawler := crawler.NewNaverWeatherCrawler(
		10*time.Second, // timeout
		3,              // max retries
	)

	// Initialize FCM notifier
	fcmNotifier, err := initFCMNotifier(config.FCMCredentialsPath, logger)
	if err != nil {
		logger.Fatal("Failed to initialize FCM notifier", zap.Error(err))
	}

	// Initialize repository
	repo := repository.NewSchedulerWeatherRepository(db)

	// Create scheduler service
	schedulerService := scheduler.NewWeatherSchedulerService(
		repo,
		weatherCrawler,
		weatherCache,
		fcmNotifier,
		logger,
		config.SchedulerInterval,
	)

	// Create health checker
	healthChecker := health.NewHealthChecker(
		db,
		redisClient,
		logger,
		version,
		func() bool {
			return schedulerService != nil
		},
	)

	// Start metrics server
	metricsPort := getEnvInt("METRICS_PORT", 9090)
	metricsServer := NewMetricsServer(metricsPort, healthChecker, logger)
	go func() {
		if err := metricsServer.Start(); err != nil {
			logger.Error("Metrics server error", zap.Error(err))
		}
	}()

	// Start scheduler in goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := schedulerService.Start(ctx); err != nil {
			if err != context.Canceled {
				logger.Error("Scheduler error", zap.Error(err))
				healthChecker.RecordError()
			}
		}
	}()

	logger.Info("Weather Scheduler Service started successfully",
		zap.Duration("interval", config.SchedulerInterval),
		zap.Int("metrics_port", metricsPort))

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	sig := <-quit

	logger.Info("Shutdown signal received",
		zap.String("signal", sig.String()))

	// Cancel context to stop scheduler
	cancel()

	// Stop scheduler and metrics server with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop scheduler
	done := make(chan error, 1)
	go func() {
		done <- schedulerService.Stop()
	}()

	select {
	case <-shutdownCtx.Done():
		logger.Warn("Shutdown timeout, forcing exit")
	case err := <-done:
		if err != nil {
			logger.Error("Shutdown error", zap.Error(err))
		} else {
			logger.Info("Shutdown completed gracefully")
		}
	}

	// Stop metrics server
	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error stopping metrics server", zap.Error(err))
	}

	logger.Info("Weather Scheduler Service stopped")
}

// Config holds application configuration
type Config struct {
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	RedisHost          string
	RedisPort          string
	RedisPassword      string
	FCMCredentialsPath string
	SchedulerInterval  time.Duration
	LogLevel           string
}

// loadConfig loads configuration from environment variables
func loadConfig() *Config {
	intervalStr := getEnv("SCHEDULER_INTERVAL", "1m")
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		interval = 1 * time.Minute
	}

	return &Config{
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "3306"),
		DBUser:             getEnv("DB_USER", "root"),
		DBPassword:         getEnv("DB_PASSWORD", ""),
		DBName:             getEnv("DB_NAME", "joker"),
		RedisHost:          getEnv("REDIS_HOST", "localhost"),
		RedisPort:          getEnv("REDIS_PORT", "6379"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		FCMCredentialsPath: getEnv("FCM_CREDENTIALS_PATH", ""),
		SchedulerInterval:  interval,
		LogLevel:           getEnv("LOG_LEVEL", "info"),
	}
}

// validateConfig validates critical configuration
func validateConfig(config *Config) error {
	if config.DBHost == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if config.DBUser == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if config.DBName == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if config.RedisHost == "" {
		return fmt.Errorf("REDIS_HOST is required")
	}
	if config.FCMCredentialsPath == "" {
		return fmt.Errorf("FCM_CREDENTIALS_PATH is required")
	}
	return nil
}

// initDatabase initializes database connection
func initDatabase(config *Config, logger *zap.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.DBUser,
		config.DBPassword,
		config.DBHost,
		config.DBPort,
		config.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: nil, // Use custom logger if needed
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connected successfully",
		zap.String("host", config.DBHost),
		zap.String("database", config.DBName))

	return db, nil
}

// initFCMNotifier initializes FCM notifier
func initFCMNotifier(credentialsPath string, logger *zap.Logger) (*notifier.FCMNotifier, error) {
	// Check if credentials file exists
	if _, err := os.Stat(credentialsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("FCM credentials file not found at %s", credentialsPath)
	}

	fcmNotifier, err := notifier.NewFCMNotifier(credentialsPath, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create FCM notifier: %w", err)
	}

	logger.Info("FCM notifier initialized successfully")
	return fcmNotifier, nil
}

// initLogger initializes zap logger
func initLogger() (*zap.Logger, error) {
	logLevel := getEnv("LOG_LEVEL", "info")

	var config zap.Config
	if os.Getenv("ENV") == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	// Set log level
	switch logLevel {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	return config.Build()
}

// getEnv gets environment variable with fallback
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets environment variable as int with fallback
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}
