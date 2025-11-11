# Weather Service Deployment Guide

Complete deployment guide for the Weather Data Collector service across different environments.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Configuration](#configuration)
- [Local Development](#local-development)
- [Docker Deployment](#docker-deployment)
- [Systemd Deployment](#systemd-deployment)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### Required

- Go 1.24+ (for building from source)
- MySQL 8.0+
- Redis 7+
- Firebase Cloud Messaging credentials

### Optional

- Docker & Docker Compose (for containerized deployment)
- Kubernetes cluster (for K8s deployment)
- systemd (for Linux service deployment)

## Configuration

### Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

#### Required Variables

```bash
# Database (REQUIRED)
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=joker

# Redis (REQUIRED)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# FCM (REQUIRED)
FCM_CREDENTIALS_PATH=/path/to/fcm-credentials.json
```

#### Optional Variables

```bash
# Scheduler
SCHEDULER_INTERVAL=1m

# Crawler
CRAWLER_TIMEOUT=10s
CRAWLER_RETRY_COUNT=3

# Logging
LOG_LEVEL=info
LOG_OUTPUT=stdout

# Metrics
METRICS_PORT=9090

# Cache
WEATHER_CACHE_TTL=30m
```

### Configuration Validation

The service validates critical configuration on startup:
- Database connection parameters
- Redis connection parameters
- FCM credentials file existence
- All required environment variables

If validation fails, the service will exit with an error message.

## Local Development

### Quick Start

1. **Install dependencies:**

```bash
go mod download
```

2. **Start test dependencies:**

```bash
make docker-test-up
```

3. **Configure environment:**

```bash
export DB_HOST=localhost
export DB_PORT=3307
export DB_USER=test_user
export DB_PASSWORD=test_password
export DB_NAME=joker_test
export REDIS_HOST=localhost
export REDIS_PORT=6380
export FCM_CREDENTIALS_PATH=./fcm-credentials.json
```

4. **Run scheduler:**

```bash
make run-scheduler
```

Or use the development target:

```bash
make dev-scheduler
```

### Build from Source

```bash
# Build scheduler binary
make build-scheduler

# Binary will be at: bin/scheduler
./bin/scheduler
```

### Running Tests

```bash
# Unit tests
make test-unit

# Integration tests (requires Docker)
make test-integration

# All tests
make test

# With coverage
make test-coverage
```

## Docker Deployment

### Using Docker Compose

1. **Configure environment:**

Create `.env` file with production settings.

2. **Ensure FCM credentials are available:**

```bash
# Place your FCM credentials file
cp /path/to/your/firebase-credentials.json ./fcm-credentials.json
```

3. **Start all services:**

```bash
docker-compose up -d
```

4. **View logs:**

```bash
docker-compose logs -f weather-scheduler
```

5. **Stop services:**

```bash
docker-compose down
```

### Building Docker Image

```bash
# Build image
docker build -t weather-scheduler:latest .

# Verify image size (should be < 50MB)
docker images weather-scheduler

# Run container
docker run -d \
  --name weather-scheduler \
  --env-file .env \
  -v $(pwd)/fcm-credentials.json:/app/config/fcm-credentials.json:ro \
  -p 9090:9090 \
  weather-scheduler:latest
```

### Docker Image Features

- **Multi-stage build** for minimal size
- **Non-root user** (weather:1000)
- **Health check** endpoint
- **Alpine-based** runtime
- **Final size:** < 50MB

## Systemd Deployment

For production Linux servers using systemd.

### Installation Steps

1. **Build binary:**

```bash
./scripts/build.sh linux
```

2. **Create deployment directory:**

```bash
sudo mkdir -p /opt/weather-scheduler/{bin,config,logs}
```

3. **Install binary:**

```bash
sudo cp bin/linux_amd64/scheduler /opt/weather-scheduler/bin/
sudo chmod +x /opt/weather-scheduler/bin/scheduler
```

4. **Configure environment:**

```bash
sudo cp .env /opt/weather-scheduler/.env
sudo cp fcm-credentials.json /opt/weather-scheduler/config/
```

5. **Create user:**

```bash
sudo useradd -r -s /bin/false weather
sudo chown -R weather:weather /opt/weather-scheduler
```

6. **Install systemd service:**

```bash
sudo cp deploy/systemd/weather-scheduler.service /etc/systemd/system/
sudo systemctl daemon-reload
```

7. **Enable and start:**

```bash
sudo systemctl enable weather-scheduler
sudo systemctl start weather-scheduler
```

### Service Management

```bash
# Check status
sudo systemctl status weather-scheduler

# View logs
sudo journalctl -u weather-scheduler -f

# Restart service
sudo systemctl restart weather-scheduler

# Stop service
sudo systemctl stop weather-scheduler
```

### Directory Structure

```
/opt/weather-scheduler/
├── bin/
│   └── scheduler           # Binary
├── config/
│   └── fcm-credentials.json
├── logs/                   # Optional log directory
└── .env                    # Environment configuration
```

## Kubernetes Deployment

For production Kubernetes deployments.

### Prerequisites

- Kubernetes cluster 1.24+
- kubectl configured
- Namespace created: `joker-backend`

### Deployment Steps

1. **Create namespace:**

```bash
kubectl create namespace joker-backend
```

2. **Create FCM credentials secret:**

```bash
kubectl create secret generic fcm-credentials \
  --from-file=fcm-credentials.json=/path/to/fcm-credentials.json \
  -n joker-backend
```

3. **Update secrets:**

Edit `deploy/k8s/secret.yaml` with your database credentials:

```yaml
stringData:
  DB_USER: "your_user"
  DB_PASSWORD: "your_password"
```

4. **Apply Kubernetes manifests:**

```bash
# ConfigMap
kubectl apply -f deploy/k8s/configmap.yaml

# Secret
kubectl apply -f deploy/k8s/secret.yaml

# Deployment
kubectl apply -f deploy/k8s/deployment.yaml

# Service
kubectl apply -f deploy/k8s/service.yaml

# HPA (optional)
kubectl apply -f deploy/k8s/hpa.yaml
```

5. **Verify deployment:**

```bash
# Check pods
kubectl get pods -n joker-backend -l app=weather-scheduler

# Check logs
kubectl logs -n joker-backend -l app=weather-scheduler -f

# Check service
kubectl get svc -n joker-backend weather-scheduler-metrics
```

### Kubernetes Features

- **Rolling updates** with zero downtime
- **Health checks** (liveness, readiness, startup)
- **Auto-scaling** via HPA (2-5 replicas)
- **Resource limits** (CPU: 1 core, Memory: 512Mi)
- **Security context** (non-root, read-only filesystem)
- **Pod anti-affinity** for high availability

### Scaling

```bash
# Manual scaling
kubectl scale deployment weather-scheduler -n joker-backend --replicas=3

# Auto-scaling (via HPA)
kubectl get hpa -n joker-backend weather-scheduler-hpa
```

### Monitoring

```bash
# Access metrics endpoint
kubectl port-forward -n joker-backend svc/weather-scheduler-metrics 9090:9090

# Visit: http://localhost:9090/metrics
```

## Build and Release Scripts

### Build Script

Build binaries for multiple platforms:

```bash
# Build for all platforms
./scripts/build.sh

# Build for specific platform
./scripts/build.sh linux          # Linux AMD64
./scripts/build.sh linux-arm64    # Linux ARM64
./scripts/build.sh darwin         # macOS AMD64
./scripts/build.sh darwin-arm64   # macOS ARM64 (Apple Silicon)
./scripts/build.sh windows        # Windows AMD64

# Clean builds
./scripts/build.sh clean
```

Binaries are output to `bin/<platform>_<arch>/scheduler`.

### Release Script

Build binaries and Docker image:

```bash
# Full release (build + Docker + push)
VERSION=1.0.0 ./scripts/release.sh

# Build without pushing to registry
./scripts/release.sh --skip-push

# Use custom registry
./scripts/release.sh --registry=ghcr.io/jokertrickster

# Skip binary build (Docker only)
./scripts/release.sh --skip-build
```

## Health Checks

The service exposes health check endpoints on the metrics port (default: 9090).

### Endpoints

- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics

### Health Check Response

```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": "2h15m30s",
  "checks": {
    "database": "ok",
    "redis": "ok",
    "scheduler": "ok"
  },
  "last_error": null
}
```

### Testing Health

```bash
# Local
curl http://localhost:9090/health

# Docker
curl http://localhost:9090/health

# Kubernetes
kubectl port-forward -n joker-backend svc/weather-scheduler-metrics 9090:9090
curl http://localhost:9090/health
```

## Graceful Shutdown

The service implements graceful shutdown with a 30-second timeout:

1. **Signal received** (SIGTERM, SIGINT)
2. **Stop scheduler** (no new jobs)
3. **Wait for in-flight jobs** (max 30s)
4. **Close connections** (DB, Redis)
5. **Exit**

### Testing Shutdown

```bash
# Send SIGTERM
kill -TERM <pid>

# Or use systemd
sudo systemctl stop weather-scheduler

# Docker
docker stop weather-scheduler

# Kubernetes
kubectl delete pod -n joker-backend <pod-name>
```

## Monitoring and Observability

### Prometheus Metrics

Metrics are exposed at `http://localhost:9090/metrics`:

- `weather_scheduler_collections_total` - Total weather collections
- `weather_scheduler_notifications_total` - Total notifications sent
- `weather_scheduler_errors_total` - Total errors
- `weather_cache_hits_total` - Cache hits
- `weather_cache_misses_total` - Cache misses
- `weather_crawler_duration_seconds` - Crawler duration

### Log Levels

Set via `LOG_LEVEL` environment variable:

- `debug` - Detailed debugging information
- `info` - General information (default)
- `warn` - Warning messages
- `error` - Error messages only

### Structured Logging

Logs are output in JSON format (production) or console format (development):

```json
{
  "level": "info",
  "ts": "2025-11-11T10:30:00.000Z",
  "msg": "Weather data collected",
  "location_id": 123,
  "duration_ms": 1234
}
```

## Troubleshooting

### Common Issues

#### Service won't start

**Symptom:** Service exits immediately

**Solutions:**
1. Check configuration validation:
   ```bash
   # Verify all required env vars are set
   ./scheduler  # Will show validation errors
   ```

2. Check FCM credentials:
   ```bash
   # Verify file exists and is valid JSON
   cat $FCM_CREDENTIALS_PATH | jq .
   ```

3. Check database connection:
   ```bash
   mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME
   ```

#### High memory usage

**Symptom:** Memory consumption increases over time

**Solutions:**
1. Check cache TTL settings
2. Review scheduler interval
3. Monitor goroutine leaks:
   ```bash
   curl http://localhost:9090/debug/pprof/goroutine
   ```

#### Notifications not sending

**Symptom:** No FCM notifications sent

**Solutions:**
1. Check FCM credentials validity
2. Verify device tokens in database
3. Check FCM logs for errors
4. Test FCM connectivity

#### Database connection failures

**Symptom:** Connection pool exhausted

**Solutions:**
1. Increase connection pool size
2. Check MySQL max_connections
3. Review query performance
4. Monitor active connections

### Debug Mode

Enable debug logging:

```bash
export LOG_LEVEL=debug
./scheduler
```

### Getting Support

1. Check logs: `journalctl -u weather-scheduler -f`
2. Check health endpoint: `curl http://localhost:9090/health`
3. Review metrics: `curl http://localhost:9090/metrics`
4. Open GitHub issue with logs and configuration

## Security Considerations

1. **Non-root execution** - Always run as dedicated user
2. **Credentials management** - Use secrets management (Vault, K8s Secrets)
3. **Network isolation** - Firewall rules for database/redis access
4. **TLS/SSL** - Enable for database and redis connections
5. **Resource limits** - Set memory and CPU limits
6. **Read-only filesystem** - Use in containers when possible

## Performance Tuning

### Database

```bash
# Connection pool
DB_MAX_OPEN_CONNS=100
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=1h
```

### Redis

```bash
# Cache TTL
WEATHER_CACHE_TTL=30m

# Connection pool
REDIS_POOL_SIZE=10
```

### Scheduler

```bash
# Adjust interval based on load
SCHEDULER_INTERVAL=1m  # Faster updates
SCHEDULER_INTERVAL=5m  # Lower load
```

## Backup and Recovery

### Database Backup

```bash
# Backup alarms table
mysqldump -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME alarms > backup.sql

# Restore
mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME < backup.sql
```

### Redis Backup

```bash
# Redis persistence is automatic with AOF
# Backup RDB file
cp /var/lib/redis/dump.rdb /backup/

# Restore
cp /backup/dump.rdb /var/lib/redis/
```

## Migration Guide

### From Manual to Docker

1. Export current configuration
2. Create `.env` file
3. Build Docker image
4. Test with docker-compose
5. Migrate production

### From Systemd to Kubernetes

1. Create K8s secrets from .env
2. Build and push Docker image
3. Apply K8s manifests
4. Test deployment
5. Update DNS/load balancer
6. Stop systemd service

## Appendix

### Environment Profiles

#### Development
```bash
ENV=development
LOG_LEVEL=debug
SCHEDULER_INTERVAL=30s
```

#### Staging
```bash
ENV=staging
LOG_LEVEL=info
SCHEDULER_INTERVAL=1m
```

#### Production
```bash
ENV=production
LOG_LEVEL=warn
SCHEDULER_INTERVAL=1m
```

### Resource Requirements

| Environment | CPU | Memory | Storage |
|------------|-----|--------|---------|
| Development | 0.5 | 256MB | 1GB |
| Staging | 1.0 | 512MB | 5GB |
| Production | 2.0 | 1GB | 20GB |

### Port Reference

| Port | Service | Description |
|------|---------|-------------|
| 9090 | Metrics | Prometheus metrics and health |
| 3306 | MySQL | Database |
| 6379 | Redis | Cache |

---

**Last Updated:** 2025-11-11
**Version:** 1.0.0
