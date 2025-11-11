package scheduler

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/cache"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/crawler"
	_interface "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/interface"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ExampleSchedulerIntegration demonstrates how to integrate the scheduler service
// This is a reference implementation - adapt to your application structure
func ExampleSchedulerIntegration(db *gorm.DB, redisAddr string) error {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}
	defer logger.Sync()

	// Initialize components
	repo := repository.NewSchedulerWeatherRepository(db)
	weatherCrawler := crawler.NewNaverWeatherCrawler(10*time.Second, 3)
	weatherCache, err := cache.NewWeatherCache(redisAddr, "", logger)
	if err != nil {
		logger.Error("Failed to initialize cache", zap.Error(err))
		return err
	}
	defer weatherCache.Close()

	// Initialize FCM notifier (placeholder - implement in Task #7)
	var fcmNotifier _interface.IFCMNotifier
	// fcmNotifier = fcm.NewNotifier(...)  // TODO: Implement FCM notifier

	// Create scheduler with 1-minute interval
	scheduler := NewWeatherSchedulerService(
		repo,
		weatherCrawler,
		weatherCache,
		fcmNotifier,
		logger,
		1*time.Minute,
	)

	// Start scheduler in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	schedulerErr := make(chan error, 1)
	go func() {
		schedulerErr <- scheduler.Start(ctx)
	}()

	logger.Info("Weather scheduler started successfully")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
		cancel() // Cancel context

	case err := <-schedulerErr:
		logger.Error("Scheduler stopped unexpectedly", zap.Error(err))
		return err
	}

	// Graceful shutdown
	logger.Info("Initiating graceful shutdown")
	if err := scheduler.Stop(); err != nil {
		logger.Error("Error during scheduler shutdown", zap.Error(err))
		return err
	}

	logger.Info("Scheduler stopped gracefully")
	return nil
}

// MinimalExample shows minimal setup for testing
func MinimalExample() {
	// Create logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Mock dependencies for testing
	// repo := &MockRepository{}
	// crawler := &MockCrawler{}
	// cache := &MockCache{}
	// notifier := &MockNotifier{}

	// Create scheduler
	// scheduler := NewWeatherSchedulerService(
	// 	repo,
	// 	crawler,
	// 	cache,
	// 	notifier,
	// 	logger,
	// 	1*time.Minute,
	// )

	// Start in background
	// ctx := context.Background()
	// go scheduler.Start(ctx)

	// Run for some time
	// time.Sleep(5 * time.Minute)

	// Stop gracefully
	// scheduler.Stop()
}

// DockerComposeExample shows Docker Compose configuration
/*
version: '3.8'

services:
  weather-scheduler:
    build:
      context: .
      dockerfile: Dockerfile
    environment:
      - DB_HOST=mysql
      - DB_PORT=3306
      - DB_USER=root
      - DB_PASSWORD=rootpassword
      - DB_NAME=joker_db
      - REDIS_ADDR=redis:6379
      - SCHEDULER_INTERVAL=1m
      - LOG_LEVEL=info
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: unless-stopped
    deploy:
      replicas: 1  # IMPORTANT: Only run one scheduler instance!
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 256M

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: joker_db
    ports:
      - "3306:3306"
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5
*/

// KubernetesExample shows Kubernetes deployment configuration
/*
apiVersion: apps/v1
kind: Deployment
metadata:
  name: weather-scheduler
  namespace: weather-service
spec:
  # IMPORTANT: Only 1 replica to prevent duplicate scheduling
  replicas: 1
  strategy:
    type: Recreate  # Ensure old pod stops before new one starts
  selector:
    matchLabels:
      app: weather-scheduler
  template:
    metadata:
      labels:
        app: weather-scheduler
    spec:
      containers:
      - name: scheduler
        image: your-registry/weather-service:latest
        env:
        - name: DB_HOST
          valueFrom:
            configMapKeyRef:
              name: weather-config
              key: db-host
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: weather-secrets
              key: db-password
        - name: REDIS_ADDR
          valueFrom:
            configMapKeyRef:
              name: weather-config
              key: redis-addr
        - name: SCHEDULER_INTERVAL
          value: "1m"
        resources:
          requests:
            memory: "256Mi"
            cpu: "200m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: weather-config
  namespace: weather-service
data:
  db-host: "mysql-service"
  redis-addr: "redis-service:6379"
---
apiVersion: v1
kind: Secret
metadata:
  name: weather-secrets
  namespace: weather-service
type: Opaque
data:
  db-password: <base64-encoded-password>
*/

// SystemdServiceExample shows systemd unit file for Linux deployment
/*
[Unit]
Description=Weather Alarm Scheduler Service
After=network.target mysql.service redis.service
Requires=mysql.service redis.service

[Service]
Type=simple
User=weather-service
Group=weather-service
WorkingDirectory=/opt/weather-service
Environment="DB_HOST=localhost"
Environment="DB_PORT=3306"
Environment="REDIS_ADDR=localhost:6379"
Environment="SCHEDULER_INTERVAL=1m"
ExecStart=/opt/weather-service/bin/scheduler
Restart=on-failure
RestartSec=10s
StandardOutput=journal
StandardError=journal

# Security hardening
PrivateTmp=true
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true

[Install]
WantedBy=multi-user.target
*/

// MonitoringExample shows Prometheus metrics integration
/*
package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	alarmsProcessedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "weather_scheduler_alarms_processed_total",
			Help: "Total number of alarms processed",
		},
		[]string{"status"}, // success, failure
	)

	processingDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "weather_scheduler_processing_duration_seconds",
			Help:    "Time spent processing alarms",
			Buckets: prometheus.DefBuckets,
		},
	)

	cacheHitRatio = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "weather_scheduler_cache_hit_ratio",
			Help: "Ratio of cache hits to total cache queries",
		},
	)

	activeAlarmsGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "weather_scheduler_active_alarms",
			Help: "Number of active alarms in the system",
		},
	)
)

// Usage in scheduler:
// alarmsProcessedTotal.WithLabelValues("success").Inc()
// alarmsProcessedTotal.WithLabelValues("failure").Inc()
// processingDuration.Observe(duration.Seconds())
// cacheHitRatio.Set(hits / total)
*/
