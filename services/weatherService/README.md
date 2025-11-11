# Weather Data Collector Service

A robust, production-ready weather notification system that delivers timely weather alerts to users via Firebase Cloud Messaging (FCM).

**Version:** 1.0.0
**Last Updated:** 2025-11-11

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [API Documentation](#api-documentation)
- [Development](#development)
- [Testing](#testing)
- [Deployment](#deployment)
- [Monitoring](#monitoring)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)

## Overview

The Weather Data Collector Service is a scheduled notification system that:

1. Monitors user-configured weather alarms in the database
2. Fetches real-time weather data from Naver Weather
3. Caches weather data in Redis for performance
4. Delivers personalized notifications via Firebase Cloud Messaging
5. Tracks notification delivery and manages alarm state

### Purpose

Provide users with timely, location-specific weather alerts at their preferred times, with high reliability and low latency.

### Key Components

- **Scheduler**: Time-based alarm processor with 1-minute granularity
- **Crawler**: Naver Weather data scraper with retry logic
- **Cache**: Redis-based weather data cache (10-minute TTL)
- **Notifier**: FCM batch notification sender (up to 500 tokens/batch)
- **Repository**: Database access layer for alarms and tokens

## Features

- **Scheduled Notifications**: Process alarms every minute with configurable intervals
- **Intelligent Caching**: 10-minute Redis cache reduces API calls by ~70%
- **Batch Processing**: Send to 500+ users simultaneously via FCM multicast
- **Retry Logic**: Exponential backoff for crawler (3 retries) and FCM (1 retry)
- **Graceful Shutdown**: 30-second timeout for in-flight operations
- **Health Monitoring**: Prometheus metrics + health check endpoint
- **Production Ready**: Structured logging, error tracking, connection pooling

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│  Scheduler  │────▶│  Repository  │────▶│   MySQL     │
│  (1-min)    │     │              │     │  (Alarms)   │
└─────────────┘     └──────────────┘     └─────────────┘
       │
       ├──────────▶┌──────────────┐     ┌─────────────┐
       │           │    Cache     │────▶│   Redis     │
       │           │  (10-min)    │     │  (Weather)  │
       │           └──────────────┘     └─────────────┘
       │
       ├──────────▶┌──────────────┐     ┌─────────────┐
       │           │   Crawler    │────▶│   Naver     │
       │           │  (Fallback)  │     │  Weather    │
       │           └──────────────┘     └─────────────┘
       │
       └──────────▶┌──────────────┐     ┌─────────────┐
                   │   Notifier   │────▶│     FCM     │
                   │  (Batch)     │     │  (Push)     │
                   └──────────────┘     └─────────────┘
```

For detailed architecture diagrams, see [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## Quick Start

### Prerequisites

- Go 1.21+
- MySQL 8.0+
- Redis 7.0+
- Firebase project with FCM enabled

### Installation

```bash
# Clone repository
git clone https://github.com/JokerTrickster/joker_backend.git
cd joker_backend/services/weatherService

# Install dependencies
go mod download

# Copy environment template
cp .env.example .env

# Edit configuration
vim .env
```

### Configuration

Edit `.env` with your settings:

```bash
# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=joker

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Firebase
FCM_CREDENTIALS_PATH=/path/to/firebase-credentials.json

# Scheduler
SCHEDULER_INTERVAL=1m
METRICS_PORT=9090

# Logging
LOG_LEVEL=info
ENV=production
```

### Run Scheduler

```bash
# Build
make build-scheduler

# Run
./scheduler
```

### Verify

```bash
# Check health
curl http://localhost:9090/health

# Check metrics
curl http://localhost:9090/metrics
```

## Configuration

See [docs/CONFIGURATION.md](docs/CONFIGURATION.md) for complete reference.

### Required Variables

| Variable | Type | Description | Example |
|----------|------|-------------|---------|
| `DB_HOST` | String | MySQL hostname | `localhost` |
| `DB_USER` | String | MySQL username | `root` |
| `DB_PASSWORD` | String | MySQL password | `secret123` |
| `DB_NAME` | String | Database name | `joker` |
| `REDIS_HOST` | String | Redis hostname | `localhost` |
| `FCM_CREDENTIALS_PATH` | String | Firebase credentials | `/etc/firebase.json` |

### Optional Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `SCHEDULER_INTERVAL` | Duration | `1m` | Alarm check interval |
| `METRICS_PORT` | Integer | `9090` | Metrics server port |
| `LOG_LEVEL` | String | `info` | Log verbosity |
| `ENV` | String | `development` | Environment |

## API Documentation

The service exposes internal interfaces for modularity and testability.

### Repository Interface

```go
type ISchedulerWeatherRepository interface {
    GetAlarmsToNotify(ctx context.Context, targetTime time.Time) ([]entity.UserAlarm, error)
    UpdateLastSent(ctx context.Context, alarmID int, sentTime time.Time) error
    GetFCMTokens(ctx context.Context, userID int) ([]entity.WeatherServiceToken, error)
}
```

### Cache Interface

```go
type IWeatherCache interface {
    Get(ctx context.Context, region string) (*entity.WeatherData, error)
    Set(ctx context.Context, region string, data *entity.WeatherData) error
    Delete(ctx context.Context, region string) error
    Close() error
    Ping(ctx context.Context) error
    GetTTL(ctx context.Context, region string) (time.Duration, error)
}
```

See [docs/API.md](docs/API.md) for complete interface documentation.

## Development

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for development guide.

### Setup Development Environment

```bash
# Install dependencies
go mod download

# Run tests
make test

# Run integration tests
make integration-test

# Check coverage
make coverage

# Run linter
make lint
```

### Code Organization

```
services/weatherService/
├── cmd/
│   ├── main.go              # API server
│   └── scheduler/
│       └── main.go          # Scheduler service
├── features/weather/
│   ├── handler/             # HTTP handlers
│   ├── repository/          # Data access
│   ├── scheduler/           # Alarm processor
│   ├── crawler/             # Weather fetcher
│   ├── cache/               # Redis cache
│   ├── notifier/            # FCM sender
│   └── model/               # Data models
├── pkg/
│   ├── health/              # Health checks
│   ├── metrics/             # Prometheus
│   └── logger/              # Structured logging
└── tests/                   # Integration tests
```

## Testing

See [docs/DEVELOPMENT.md#testing](docs/DEVELOPMENT.md#testing) for testing guide.

### Unit Tests

```bash
# Run all unit tests
make test

# Run specific package
go test ./features/weather/scheduler/...

# With coverage
make coverage
```

### Integration Tests

```bash
# Start test dependencies
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
make integration-test

# Clean up
docker-compose -f docker-compose.test.yml down
```

### Test Coverage

Current coverage: ~85%

- Scheduler: 95%
- Repository: 90%
- Cache: 88%
- Crawler: 82%
- Notifier: 85%

## Deployment

See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for deployment guides.

### Docker

```bash
# Build image
docker build -t weather-scheduler:1.0.0 .

# Run container
docker run -d \
  --name weather-scheduler \
  --env-file .env \
  -p 9090:9090 \
  weather-scheduler:1.0.0
```

### Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f scheduler

# Stop services
docker-compose down
```

### Systemd

```bash
# Install service
sudo cp scripts/weather-scheduler.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable weather-scheduler
sudo systemctl start weather-scheduler

# Check status
sudo systemctl status weather-scheduler
```

### Kubernetes

```bash
# Apply manifests
kubectl apply -f k8s/

# Check status
kubectl get pods -l app=weather-scheduler

# View logs
kubectl logs -l app=weather-scheduler -f
```

## Monitoring

See [docs/MONITORING.md](docs/MONITORING.md) for monitoring guide.

### Health Check

```bash
curl http://localhost:9090/health
```

Response:
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "timestamp": "2025-11-11T10:30:00Z",
  "dependencies": {
    "database": "healthy",
    "redis": "healthy",
    "scheduler": "healthy"
  }
}
```

### Prometheus Metrics

Key metrics exposed at `http://localhost:9090/metrics`:

- `weather_alarms_processed_total`: Total alarms processed
- `weather_alarms_failed_total`: Total alarm failures
- `weather_cache_hits_total`: Cache hit count
- `weather_cache_misses_total`: Cache miss count
- `weather_fcm_sent_total`: FCM notifications sent
- `weather_fcm_failed_total`: FCM notification failures
- `weather_processing_duration_seconds`: Processing latency

### Grafana Dashboard

Import `grafana-dashboard.json` for pre-configured visualization.

### Alerts

Recommended alerts:

- Cache hit rate < 70%
- FCM success rate < 95%
- Processing latency > 500ms
- Error rate > 10 errors/minute
- No alarms processed in 5 minutes

## Troubleshooting

See [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) for detailed guide.

### Common Issues

**Alarms not processing**
```bash
# Check scheduler status
curl http://localhost:9090/health

# Check logs
tail -f /var/log/weather-scheduler.log

# Verify alarm configuration
mysql -e "SELECT * FROM weather_alarms WHERE is_enabled=1"
```

**High latency**
```bash
# Check cache hit rate
curl http://localhost:9090/metrics | grep cache_hits

# Check Redis connection
redis-cli ping

# Check database queries
mysql -e "SHOW PROCESSLIST"
```

**FCM failures**
```bash
# Verify credentials
test -f /path/to/firebase-credentials.json

# Check FCM quota
# Visit Firebase Console → Cloud Messaging

# Verify tokens
mysql -e "SELECT COUNT(*) FROM weather_service_tokens"
```

## Contributing

### Development Workflow

1. Fork repository
2. Create feature branch: `git checkout -b feature/alarm-snooze`
3. Make changes with tests
4. Run tests: `make test`
5. Run linter: `make lint`
6. Commit with conventional commits: `feat: add alarm snooze`
7. Push and create pull request

### Code Review Checklist

- [ ] All tests pass
- [ ] Code coverage maintained (>80%)
- [ ] Linter passes
- [ ] Documentation updated
- [ ] Changelog updated
- [ ] No breaking changes (or documented)

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Example:
```
feat(scheduler): add alarm snooze feature

Allow users to snooze alarms for 10 minutes.

Closes #123
```

## License

Copyright (c) 2025 Joker Backend Team

## Support

- GitHub Issues: https://github.com/JokerTrickster/joker_backend/issues
- Documentation: [docs/](docs/)
- Email: support@example.com

## Changelog

### v1.0.0 (2025-11-11)

- Initial release
- Scheduler with 1-minute granularity
- Redis caching with 10-minute TTL
- FCM batch notifications
- Prometheus metrics
- Health check endpoint
- Graceful shutdown
