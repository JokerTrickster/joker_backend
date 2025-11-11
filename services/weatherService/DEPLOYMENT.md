# Weather Data Collector - Deployment Guide

This guide covers deployment and operation of the Weather Scheduler Service.

## Table of Contents

- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Configuration](#configuration)
- [Local Development](#local-development)
- [Testing](#testing)
- [Production Deployment](#production-deployment)
- [Monitoring](#monitoring)
- [Troubleshooting](#troubleshooting)

## Architecture

The Weather Scheduler Service is a standalone Go application that:

1. **Runs on a schedule** (default: every 1 minute)
2. **Fetches alarms** from MySQL database for upcoming notification times
3. **Retrieves weather data** from cache (Redis) or crawler (Naver)
4. **Sends notifications** via Firebase Cloud Messaging (FCM)
5. **Updates alarm status** to prevent duplicate notifications

### Component Flow

```
┌─────────────┐
│  Scheduler  │ (Ticker: 1 minute)
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Repository  │ ──► MySQL (get alarms, FCM tokens)
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Cache     │ ──► Redis (weather data, 30min TTL)
└──────┬──────┘
       │ (cache miss)
       ▼
┌─────────────┐
│  Crawler    │ ──► Naver Weather (scrape data)
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Notifier   │ ──► FCM (send push notifications)
└─────────────┘
```

## Prerequisites

### Required Services

- **MySQL 8.0+** - Database for alarms and FCM tokens
- **Redis 6.0+** - Cache for weather data (30-minute TTL)
- **Firebase Account** - FCM credentials for push notifications

### Required Software

- **Go 1.24+** - For building and running the service
- **Docker & Docker Compose** - For local testing
- **Make** - For build automation

## Configuration

### Environment Variables

The scheduler requires the following environment variables:

#### Database Configuration

```bash
DB_HOST=localhost          # MySQL host
DB_PORT=3306              # MySQL port
DB_USER=root              # MySQL user
DB_PASSWORD=password      # MySQL password
DB_NAME=joker             # Database name
```

#### Redis Configuration

```bash
REDIS_HOST=localhost      # Redis host
REDIS_PORT=6379           # Redis port
REDIS_PASSWORD=           # Redis password (empty if none)
```

#### FCM Configuration

```bash
FCM_CREDENTIALS_PATH=/path/to/fcm-credentials.json
```

The FCM credentials file is a JSON service account key from Firebase Console:
1. Go to Firebase Console → Project Settings → Service Accounts
2. Click "Generate New Private Key"
3. Save the JSON file securely

#### Scheduler Configuration

```bash
SCHEDULER_INTERVAL=1m     # Tick interval (default: 1m)
LOG_LEVEL=info            # Log level: debug, info, warn, error
ENV=production            # Environment: development, production
```

### Database Schema

The service requires the following tables:

**user_alarms**
```sql
CREATE TABLE user_alarms (
  id INT PRIMARY KEY AUTO_INCREMENT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  user_id INT NOT NULL,
  alarm_time TIME NOT NULL,
  region VARCHAR(255) NOT NULL,
  is_enabled BOOLEAN DEFAULT TRUE,
  last_sent TIMESTAMP NULL,
  INDEX idx_alarm_time (alarm_time),
  INDEX idx_user_id (user_id),
  INDEX idx_last_sent (last_sent)
);
```

**weather_service_tokens**
```sql
CREATE TABLE weather_service_tokens (
  id INT PRIMARY KEY AUTO_INCREMENT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  user_id INT NOT NULL,
  fcm_token VARCHAR(500) NOT NULL,
  device_id VARCHAR(255),
  INDEX idx_user_id (user_id)
);
```

## Local Development

### 1. Setup Test Environment

Start test containers (MySQL + Redis):

```bash
make docker-test-up
```

This starts:
- MySQL on port **3307** (joker_test database)
- Redis on port **6380**

### 2. Run Scheduler in Development Mode

```bash
make dev-scheduler
```

This runs the scheduler with:
- Debug logging enabled
- Test database connection
- 1-minute tick interval

### 3. Inspect Services

View test database:
```bash
make inspect-db
```

View test Redis:
```bash
make inspect-redis
```

### 4. Stop Test Environment

```bash
make docker-test-down
```

## Testing

### Unit Tests

Run unit tests (no external dependencies):

```bash
make test-unit
```

### Integration Tests

Run integration tests with real MySQL and Redis:

```bash
make test-integration
```

This will:
1. Start test containers
2. Run integration tests
3. Clean up containers

Integration tests cover:
- ✅ **Happy Path** - Complete flow from alarm to notification
- ✅ **Cache Hit** - Fast path using cached weather data
- ✅ **Cache Miss** - Crawler fallback when cache empty
- ✅ **Duplicate Prevention** - Last_sent logic prevents duplicate notifications
- ✅ **Multiple Alarms** - Concurrent alarm processing
- ✅ **No Tokens** - Graceful handling when user has no FCM tokens

### Load Testing

The integration tests include a load test scenario with 10 concurrent alarms:

```bash
go test -v -timeout 5m ./features/weather/integration/... -run TestEndToEnd_MultipleAlarms
```

### Coverage Report

Generate test coverage:

```bash
make test-coverage
```

Opens `coverage.html` with detailed coverage report.

## Production Deployment

### 1. Build Binary

```bash
make build-scheduler
```

Produces: `bin/scheduler`

### 2. Setup Environment

Create a production configuration file (`.env`):

```bash
# Database
DB_HOST=prod-mysql.example.com
DB_PORT=3306
DB_USER=weather_service
DB_PASSWORD=<secure_password>
DB_NAME=joker_production

# Redis
REDIS_HOST=prod-redis.example.com
REDIS_PORT=6379
REDIS_PASSWORD=<secure_password>

# FCM
FCM_CREDENTIALS_PATH=/etc/weather-scheduler/fcm-credentials.json

# Scheduler
SCHEDULER_INTERVAL=1m
LOG_LEVEL=info
ENV=production
```

### 3. Deploy Binary

**Using systemd (Linux)**

Create `/etc/systemd/system/weather-scheduler.service`:

```ini
[Unit]
Description=Weather Scheduler Service
After=network.target mysql.service redis.service

[Service]
Type=simple
User=weather-service
WorkingDirectory=/opt/weather-scheduler
EnvironmentFile=/etc/weather-scheduler/.env
ExecStart=/opt/weather-scheduler/bin/scheduler
Restart=always
RestartSec=10s
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable weather-scheduler
sudo systemctl start weather-scheduler
```

Check status:

```bash
sudo systemctl status weather-scheduler
sudo journalctl -u weather-scheduler -f
```

**Using Docker**

Create `Dockerfile`:

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o scheduler ./cmd/scheduler

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/scheduler .
COPY fcm-credentials.json .
CMD ["./scheduler"]
```

Build and run:

```bash
docker build -t weather-scheduler:latest .
docker run -d \
  --name weather-scheduler \
  --restart unless-stopped \
  -e DB_HOST=... \
  -e DB_PORT=... \
  # ... other env vars
  weather-scheduler:latest
```

### 4. Graceful Shutdown

The scheduler handles signals gracefully:

- **SIGTERM / SIGINT** - Initiates graceful shutdown
- Waits up to 30 seconds for in-flight operations to complete
- Logs all shutdown steps

Test graceful shutdown:

```bash
kill -TERM <pid>
# or
docker stop weather-scheduler
```

## Monitoring

### Health Checks

The scheduler logs key metrics:

```
INFO  Starting weather scheduler service  interval=1m
INFO  Database connected successfully  host=localhost database=joker_test
INFO  Successfully connected to Redis  address=localhost:6380
INFO  FCM notifier initialized successfully
INFO  Weather Scheduler Service started successfully  interval=1m
```

### Metrics to Monitor

1. **Alarm Processing**
   - Total alarms processed per tick
   - Success/failure counts
   - Processing latency

2. **Cache Performance**
   - Cache hit/miss ratio
   - Redis connection health

3. **Notification Delivery**
   - FCM success/failure counts
   - Token validity issues

4. **System Health**
   - Database connection pool utilization
   - Redis connection stability
   - Memory usage
   - CPU usage

### Logging

Logs are structured JSON in production:

```json
{
  "level": "info",
  "ts": 1699999999,
  "msg": "Processing alarms",
  "target_time": "2025-11-11T12:00:00Z",
  "count": 5
}
```

Log levels:
- **DEBUG** - Detailed operation traces (development only)
- **INFO** - Normal operations, alarm processing
- **WARN** - Recoverable errors, cache misses
- **ERROR** - Critical errors, database failures

### Alerts

Set up alerts for:

- ⚠️ **High failure rate** - >10% alarm processing failures
- ⚠️ **Database connectivity** - Connection failures
- ⚠️ **Redis unavailable** - Cache errors
- ⚠️ **FCM errors** - Notification delivery failures
- ⚠️ **High latency** - Processing time >5 seconds

## Troubleshooting

### Common Issues

#### 1. "FCM credentials file not found"

**Problem**: Cannot find FCM credentials file

**Solution**:
```bash
# Check file exists
ls -la $FCM_CREDENTIALS_PATH

# Verify file permissions
chmod 600 $FCM_CREDENTIALS_PATH

# Validate JSON format
jq . $FCM_CREDENTIALS_PATH
```

#### 2. "Failed to connect to database"

**Problem**: Cannot connect to MySQL

**Solution**:
```bash
# Test connection manually
mysql -h $DB_HOST -P $DB_PORT -u $DB_USER -p$DB_PASSWORD $DB_NAME

# Check network connectivity
telnet $DB_HOST $DB_PORT

# Verify credentials
echo "DB_USER=$DB_USER DB_PASSWORD=$DB_PASSWORD"
```

#### 3. "Failed to connect to Redis"

**Problem**: Cannot connect to Redis

**Solution**:
```bash
# Test connection manually
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD ping

# Check network connectivity
telnet $REDIS_HOST $REDIS_PORT
```

#### 4. "No alarms to process"

**Problem**: Scheduler runs but finds no alarms

**Solution**:
```sql
-- Check if alarms exist
SELECT * FROM user_alarms
WHERE is_enabled = true
AND deleted_at IS NULL;

-- Check alarm times
SELECT alarm_time, COUNT(*)
FROM user_alarms
WHERE is_enabled = true
GROUP BY alarm_time;

-- Check last_sent timestamps
SELECT id, alarm_time, last_sent
FROM user_alarms
WHERE is_enabled = true;
```

#### 5. "All notification sends failed"

**Problem**: FCM notifications fail

**Solution**:
```bash
# Verify FCM tokens exist
SELECT COUNT(*) FROM weather_service_tokens WHERE deleted_at IS NULL;

# Check token validity (FCM Console)
# Invalid tokens should be cleaned up

# Test FCM credentials
# Use Firebase Admin SDK to send test message
```

### Debug Mode

Enable debug logging:

```bash
LOG_LEVEL=debug ./bin/scheduler
```

Debug logs include:
- Detailed alarm processing steps
- Cache hit/miss events
- FCM token details
- SQL query execution

### Performance Tuning

#### Database Connection Pool

Adjust pool settings in `main.go`:

```go
sqlDB.SetMaxIdleConns(10)   // Idle connections
sqlDB.SetMaxOpenConns(100)  // Max open connections
sqlDB.SetConnMaxLifetime(time.Hour)
```

#### Redis Connection Pool

Adjust in `cache.go`:

```go
PoolSize:     10,  // Connection pool size
MinIdleConns: 5,   // Minimum idle connections
MaxRetries:   3,   // Retry failed commands
```

#### Crawler Timeout

Adjust in `main.go`:

```go
crawler.NewNaverWeatherCrawler(
    10*time.Second, // timeout
    3,              // max retries
)
```

## Maintenance

### Database Cleanup

Remove old alarms:

```sql
-- Delete alarms older than 30 days
DELETE FROM user_alarms
WHERE deleted_at IS NOT NULL
AND deleted_at < DATE_SUB(NOW(), INTERVAL 30 DAY);
```

### Cache Cleanup

Redis cache auto-expires after 30 minutes (TTL).

Manual cleanup if needed:

```bash
redis-cli -h $REDIS_HOST -p $REDIS_PORT
> KEYS weather:*
> DEL weather:서울시:강남구
```

### FCM Token Cleanup

Remove invalid/expired tokens:

```sql
-- Find users with no valid tokens
SELECT user_id, COUNT(*)
FROM weather_service_tokens
WHERE deleted_at IS NULL
GROUP BY user_id;

-- Soft delete invalid tokens
UPDATE weather_service_tokens
SET deleted_at = NOW()
WHERE fcm_token = '<invalid_token>';
```

## Support

For issues or questions:
- Check logs: `journalctl -u weather-scheduler -f`
- Review metrics and monitoring dashboards
- Inspect database and cache state
- Contact DevOps team

---

**Version**: 1.0.0
**Last Updated**: 2025-11-11
