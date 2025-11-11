# Weather Data Collector - Monitoring Guide

## Overview

The Weather Data Collector provides comprehensive monitoring through Prometheus metrics, structured logging with Zap, and health checks. This guide covers setup, metrics interpretation, and troubleshooting.

## Quick Start

### 1. Environment Configuration

```bash
# Logging
LOG_LEVEL=info                          # debug, info, warn, error
LOG_OUTPUT=stdout                       # stdout, stderr, or file path

# Metrics Server
METRICS_PORT=9090                       # Port for /metrics and /health endpoints

# Scheduler
SCHEDULER_INTERVAL=1m                   # How often to check for alarms

# Database & Redis
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=joker
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# FCM
FCM_CREDENTIALS_PATH=/path/to/credentials.json
```

### 2. Start the Scheduler

```bash
cd services/weatherService
go run cmd/scheduler/main.go
```

### 3. Access Endpoints

- Health Check: `http://localhost:9090/health`
- Prometheus Metrics: `http://localhost:9090/metrics`

## Health Check Endpoint

### Request

```bash
curl http://localhost:9090/health
```

### Response (Healthy)

```json
{
  "status": "ok",
  "timestamp": "2025-11-11T10:30:00Z",
  "version": "1.0.0",
  "uptime": 3600000000000,
  "components": {
    "database": "ok",
    "redis": "ok",
    "scheduler": "running"
  },
  "error_count": 0
}
```

### Response (Unhealthy)

```json
{
  "status": "error",
  "timestamp": "2025-11-11T10:30:00Z",
  "version": "1.0.0",
  "uptime": 3600000000000,
  "components": {
    "database": "ok",
    "redis": "error: connection refused",
    "scheduler": "running"
  },
  "error_count": 15
}
```

### HTTP Status Codes

- `200 OK` - All components healthy
- `503 Service Unavailable` - One or more components unhealthy

## Prometheus Metrics

### Crawler Metrics

**`weather_crawl_requests_total`** (counter)
- Description: Total number of weather crawl requests
- Labels: `region`, `status` (success/failure)
- Example: `weather_crawl_requests_total{region="서울",status="success"} 1543`

**`weather_crawl_duration_seconds`** (histogram)
- Description: Duration of weather crawl requests
- Labels: `region`
- Example: `weather_crawl_duration_seconds_bucket{region="서울",le="1"} 1520`

**`weather_crawl_errors_total`** (counter)
- Description: Total number of weather crawl errors
- Labels: `region`, `error_type` (fetch_failed, parse_error, timeout)
- Example: `weather_crawl_errors_total{region="부산",error_type="timeout"} 5`

### Cache Metrics

**`weather_cache_hits_total`** (counter)
- Description: Total number of cache hits
- Example: `weather_cache_hits_total 2345`

**`weather_cache_misses_total`** (counter)
- Description: Total number of cache misses
- Example: `weather_cache_misses_total 456`

**`weather_cache_errors_total`** (counter)
- Description: Cache operation errors
- Labels: `operation` (get, set, delete)
- Example: `weather_cache_errors_total{operation="set"} 12`

### FCM Metrics

**`fcm_notifications_sent_total`** (counter)
- Description: Total number of FCM notifications sent
- Labels: `status` (success/failure)
- Example: `fcm_notifications_sent_total{status="success"} 5432`

**`fcm_send_duration_seconds`** (histogram)
- Description: Duration of FCM send operations
- Example: `fcm_send_duration_seconds_bucket{le="0.5"} 4321`

**`fcm_batch_size`** (histogram)
- Description: Size of FCM notification batches
- Example: `fcm_batch_size_bucket{le="100"} 234`

**`fcm_errors_total`** (counter)
- Description: Total number of FCM errors
- Labels: `error_type` (send_failed, invalid_token, network_error)
- Example: `fcm_errors_total{error_type="invalid_token"} 45`

### Scheduler Metrics

**`scheduler_ticks_total`** (counter)
- Description: Total number of scheduler ticks
- Example: `scheduler_ticks_total 8640`

**`scheduler_alarms_processed_total`** (counter)
- Description: Total number of alarms processed
- Labels: `status` (success/failed)
- Example: `scheduler_alarms_processed_total{status="success"} 12345`

**`scheduler_processing_duration_seconds`** (histogram)
- Description: Duration of scheduler processing
- Example: `scheduler_processing_duration_seconds_bucket{le="5"} 8500`

**`scheduler_consecutive_failures`** (gauge)
- Description: Number of consecutive scheduler failures
- Example: `scheduler_consecutive_failures 0`

**`scheduler_error_rate`** (gauge)
- Description: Scheduler error rate (errors per minute)
- Example: `scheduler_error_rate 2.5`

## Prometheus Queries

### Performance Analysis

#### Average Crawl Duration by Region
```promql
rate(weather_crawl_duration_seconds_sum[5m]) / rate(weather_crawl_duration_seconds_count[5m])
```

#### Cache Hit Rate
```promql
rate(weather_cache_hits_total[5m]) / (rate(weather_cache_hits_total[5m]) + rate(weather_cache_misses_total[5m]))
```

#### FCM Error Rate
```promql
rate(fcm_errors_total[5m])
```

#### Alarms Processed Per Minute
```promql
rate(scheduler_alarms_processed_total[1m]) * 60
```

### Alerting Queries

#### High Crawl Error Rate (>10% errors)
```promql
rate(weather_crawl_errors_total[5m]) / rate(weather_crawl_requests_total[5m]) > 0.1
```

#### Low Cache Hit Rate (<70%)
```promql
rate(weather_cache_hits_total[5m]) / (rate(weather_cache_hits_total[5m]) + rate(weather_cache_misses_total[5m])) < 0.7
```

#### High FCM Failure Rate (>5%)
```promql
rate(fcm_notifications_sent_total{status="failure"}[5m]) / rate(fcm_notifications_sent_total[5m]) > 0.05
```

#### Scheduler Not Processing Alarms
```promql
rate(scheduler_ticks_total[5m]) == 0
```

#### High Error Rate (>10 errors/minute)
```promql
scheduler_error_rate > 10
```

## Grafana Dashboard

### Import Dashboard

1. Open Grafana
2. Go to Dashboards → Import
3. Upload `grafana-dashboard.json`
4. Select your Prometheus data source
5. Click Import

### Dashboard Panels

1. **Crawler Performance** - Average duration and success rate by region
2. **Cache Hit Rate** - Percentage of requests served from cache
3. **FCM Delivery Rate** - Success vs failure rate for notifications
4. **Scheduler Throughput** - Alarms processed per minute
5. **Error Rate** - Errors by component and type
6. **System Metrics** - Goroutines, memory, GC duration
7. **FCM Batch Size Distribution** - p50, p95, p99 batch sizes
8. **Scheduler Processing Time** - p50, p95, p99 processing duration
9. **Consecutive Failures** - Current consecutive failure count
10. **Error Rate (per minute)** - Real-time error rate
11. **Scheduler Ticks** - Ticks per minute
12. **Total Alarms Processed** - Cumulative total

## Structured Logging

### Log Levels

- **debug** - Detailed diagnostic information (cache hits/misses, individual operations)
- **info** - General informational messages (service start, successful operations)
- **warn** - Warning messages (recoverable errors, fallback to crawler)
- **error** - Error messages (operation failures, critical issues)

### Log Format

All logs are JSON-structured with the following fields:

```json
{
  "level": "info",
  "ts": 1699707600.123456,
  "caller": "scheduler/scheduler.go:150",
  "msg": "Processing alarms",
  "target_time": "2025-11-11T10:30:00Z",
  "component": "scheduler",
  "region": "서울",
  "user_id": 123,
  "alarm_id": 456
}
```

### Correlation IDs

Use the logger helper functions to add correlation fields:

```go
import customlogger "github.com/JokerTrickster/joker_backend/services/weatherService/pkg/logger"

// Add request ID
logger = customlogger.WithRequestID(logger, "req-12345")

// Add component context
logger = customlogger.WithComponent(logger, "crawler")

// Add region context
logger = customlogger.WithRegion(logger, "서울")

// Add user context
logger = customlogger.WithUserID(logger, 123)
logger = customlogger.WithAlarmID(logger, 456)
```

## Alerting Rules

### Prometheus Alert Manager

Create `/etc/prometheus/alert.rules.yml`:

```yaml
groups:
  - name: weather_collector_alerts
    interval: 30s
    rules:
      - alert: HighCrawlErrorRate
        expr: rate(weather_crawl_errors_total[5m]) / rate(weather_crawl_requests_total[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High crawl error rate"
          description: "Crawl error rate is {{ $value | humanizePercentage }} (>10%)"

      - alert: LowCacheHitRate
        expr: rate(weather_cache_hits_total[5m]) / (rate(weather_cache_hits_total[5m]) + rate(weather_cache_misses_total[5m])) < 0.7
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Low cache hit rate"
          description: "Cache hit rate is {{ $value | humanizePercentage }} (<70%)"

      - alert: HighFCMFailureRate
        expr: rate(fcm_notifications_sent_total{status="failure"}[5m]) / rate(fcm_notifications_sent_total[5m]) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High FCM failure rate"
          description: "FCM failure rate is {{ $value | humanizePercentage }} (>5%)"

      - alert: SchedulerNotProcessing
        expr: rate(scheduler_ticks_total[5m]) == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Scheduler not processing"
          description: "Scheduler has not ticked in the last 2 minutes"

      - alert: HighSchedulerErrorRate
        expr: scheduler_error_rate > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High scheduler error rate"
          description: "Scheduler error rate is {{ $value }} errors/minute (>10)"

      - alert: HighConsecutiveFailures
        expr: scheduler_consecutive_failures > 5
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "High consecutive failures"
          description: "{{ $value }} consecutive scheduler failures detected"

      - alert: ComponentUnhealthy
        expr: up{job="weather_collector"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Weather collector down"
          description: "Weather collector service is not responding"
```

## Troubleshooting

### High Crawl Error Rate

**Symptoms:** `weather_crawl_errors_total` increasing rapidly

**Possible Causes:**
1. Naver blocking requests (rate limiting)
2. Network connectivity issues
3. HTML structure changed (parse errors)

**Resolution:**
```bash
# Check error types
curl http://localhost:9090/metrics | grep weather_crawl_errors_total

# Review logs
tail -f /var/log/weather-collector.log | grep "error_type"

# Adjust retry backoff or add delays between requests
```

### Low Cache Hit Rate

**Symptoms:** Cache hit rate <70%

**Possible Causes:**
1. Cache TTL too short (30 minutes default)
2. High diversity of regions
3. Redis eviction policy

**Resolution:**
```bash
# Check Redis memory
redis-cli INFO memory

# Check cache metrics
curl http://localhost:9090/metrics | grep weather_cache

# Increase cache TTL in cache/weather.go
```

### High FCM Failure Rate

**Symptoms:** `fcm_notifications_sent_total{status="failure"}` increasing

**Possible Causes:**
1. Invalid FCM tokens
2. Firebase service issues
3. Network timeout

**Resolution:**
```bash
# Check error types
curl http://localhost:9090/metrics | grep fcm_errors_total

# Review FCM logs
tail -f /var/log/weather-collector.log | grep "FCM"

# Clean up invalid tokens from database
```

### Scheduler Not Processing

**Symptoms:** `scheduler_ticks_total` not increasing

**Possible Causes:**
1. Scheduler stopped/crashed
2. Database connection lost
3. Context cancelled

**Resolution:**
```bash
# Check health endpoint
curl http://localhost:9090/health

# Check scheduler logs
tail -f /var/log/weather-collector.log | grep "scheduler"

# Restart service
systemctl restart weather-collector
```

## Performance Tuning

### Crawler Optimization

```go
// Adjust timeout and retries
crawler := crawler.NewNaverWeatherCrawler(
    15*time.Second,  // Increase timeout
    5,               // Increase max retries
)
```

### Cache Optimization

```go
// Increase TTL in cache/weather.go
const CacheTTL = 60 * time.Minute  // Increase from 30 to 60 minutes

// Increase Redis pool size
opt := &redis.Options{
    PoolSize:     20,  // Increase from 10
    MinIdleConns: 10,  // Increase from 5
}
```

### Scheduler Optimization

```go
// Adjust processing interval
SCHEDULER_INTERVAL=30s  // Decrease from 1m for more frequent checks
```

### Database Connection Pool

```go
// In cmd/scheduler/main.go
sqlDB.SetMaxIdleConns(20)    // Increase from 10
sqlDB.SetMaxOpenConns(200)   // Increase from 100
sqlDB.SetConnMaxLifetime(2 * time.Hour)  // Increase from 1 hour
```

## Maintenance

### Log Rotation

Configure logrotate for file-based logging:

```bash
# /etc/logrotate.d/weather-collector
/var/log/weather-collector.log {
    daily
    rotate 7
    compress
    delaycompress
    notifempty
    create 0644 weather weather
    sharedscripts
    postrotate
        systemctl reload weather-collector
    endscript
}
```

### Metrics Retention

Configure Prometheus retention:

```yaml
# /etc/prometheus/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

storage:
  tsdb:
    retention.time: 30d    # Keep 30 days of data
    retention.size: 50GB   # Or 50GB limit
```

### Backup Grafana Dashboard

```bash
# Export dashboard
curl -H "Authorization: Bearer YOUR_API_KEY" \
  http://localhost:3000/api/dashboards/uid/DASHBOARD_UID \
  > dashboard-backup.json
```

## Support

For issues or questions:
1. Check logs: `/var/log/weather-collector.log`
2. Check health endpoint: `http://localhost:9090/health`
3. Review metrics: `http://localhost:9090/metrics`
4. Check Grafana dashboard for visual analysis
