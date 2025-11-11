# Weather Service Quick Start

Get the Weather Data Collector up and running in 5 minutes.

## Prerequisites

- Go 1.24+
- Docker & Docker Compose
- MySQL 8.0+
- Redis 7+

## Quick Setup

### 1. Clone and Navigate

```bash
cd services/weatherService
```

### 2. Configure Environment

```bash
# Copy example configuration
cp .env.example .env

# Edit with your settings
vim .env
```

**Minimum required:**
```bash
DB_PASSWORD=your_mysql_password
FCM_CREDENTIALS_PATH=/path/to/fcm-credentials.json
```

### 3. Start Dependencies

```bash
# Start MySQL and Redis
make docker-test-up

# Wait for services to be ready (about 10 seconds)
sleep 10
```

### 4. Run Scheduler

```bash
# Development mode (uses test containers)
make dev-scheduler
```

### 5. Verify Running

Open another terminal:

```bash
# Check health
curl http://localhost:9090/health

# Check metrics
curl http://localhost:9090/metrics
```

You should see:
```json
{
  "status": "healthy",
  "version": "dev",
  "uptime": "10s",
  "checks": {
    "database": "ok",
    "redis": "ok",
    "scheduler": "ok"
  }
}
```

## What's Running?

- **Scheduler Service:** Port 9090 (metrics/health)
- **MySQL:** Port 3307 (mapped from 3306)
- **Redis:** Port 6380 (mapped from 6379)

## Common Commands

```bash
# Build binary
make build-scheduler

# Run tests
make test

# Stop dependencies
make docker-test-down

# View logs
docker-compose -f docker-compose.test.yml logs -f
```

## Production Deployment

### Docker Compose

```bash
# Configure production environment
cp .env.example .env
# Edit .env with production settings

# Start all services
make deploy-docker

# View logs
docker-compose logs -f weather-scheduler
```

### Systemd (Linux)

```bash
# Build
make build-scheduler

# Deploy (requires sudo)
make deploy-systemd

# Start service
sudo systemctl start weather-scheduler

# Check status
sudo systemctl status weather-scheduler
```

### Kubernetes

```bash
# Update secrets in deploy/k8s/secret.yaml

# Deploy
make deploy-k8s

# Check status
kubectl get pods -n joker-backend -l app=weather-scheduler
```

## Troubleshooting

### Service won't start

```bash
# Check configuration
./bin/scheduler  # Will show validation errors

# Check database connection
mysql -h localhost -P 3307 -u test_user -ptest_password joker_test

# Check Redis connection
redis-cli -h localhost -p 6380 ping
```

### Health check fails

```bash
# View logs
make docker-test-up
make dev-scheduler  # Check console output

# Test dependencies
curl http://localhost:9090/health | jq .
```

### FCM errors

```bash
# Verify credentials file
cat $FCM_CREDENTIALS_PATH | jq .

# Check file path is correct in .env
grep FCM_CREDENTIALS_PATH .env
```

## Next Steps

1. **Configure Alarms:** Add weather alarms via API
2. **Set up Monitoring:** Import Grafana dashboard
3. **Review Logs:** Check application behavior
4. **Scale Up:** Deploy to production environment

## Documentation

- **Full Deployment Guide:** [DEPLOYMENT.md](./DEPLOYMENT.md)
- **Operations Runbook:** [OPERATIONS.md](./OPERATIONS.md)
- **Monitoring Setup:** [MONITORING.md](./MONITORING.md)
- **API Documentation:** [API.md](./API.md)

## Support

- GitHub Issues: https://github.com/JokerTrickster/joker_backend/issues
- Check logs: `make docker-logs`
- Health endpoint: http://localhost:9090/health
