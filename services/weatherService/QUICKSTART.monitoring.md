# Weather Data Collector - Monitoring Quick Start

Get up and running with monitoring in 5 minutes.

## Prerequisites

- Go 1.24+
- MySQL database
- Redis server
- Firebase credentials (for FCM)

## Step 1: Install Dependencies

```bash
cd services/weatherService
go mod tidy
```

## Step 2: Configure Environment

```bash
# Copy example configuration
cp .env.example .env

# Edit .env with your settings
vim .env
```

Required settings:
```bash
LOG_LEVEL=info
METRICS_PORT=9090
DB_HOST=localhost
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=joker
REDIS_HOST=localhost
FCM_CREDENTIALS_PATH=/path/to/firebase-credentials.json
```

## Step 3: Run Tests

```bash
# Test metrics
go test ./pkg/metrics/... -v

# Test health checks
go test ./pkg/health/... -v

# Test logger
go test ./pkg/logger/... -v
```

Expected output: All tests PASS

## Step 4: Start the Scheduler

```bash
# Build
go build -o weather-scheduler ./cmd/scheduler

# Run
./weather-scheduler
```

You should see:
```
{"level":"info","msg":"Starting Weather Scheduler Service","version":"1.0.0"}
{"level":"info","msg":"Metrics initialized"}
{"level":"info","msg":"Database connection established"}
{"level":"info","msg":"Redis connection established"}
{"level":"info","msg":"Starting metrics server","address":":9090"}
{"level":"info","msg":"Weather Scheduler Service started successfully"}
```

## Step 5: Verify Monitoring

### Check Health

```bash
curl http://localhost:9090/health | jq
```

Expected response:
```json
{
  "status": "ok",
  "timestamp": "2025-11-11T10:30:00Z",
  "version": "1.0.0",
  "uptime": 60000000000,
  "components": {
    "database": "ok",
    "redis": "ok",
    "scheduler": "running"
  },
  "error_count": 0
}
```

### Check Metrics

```bash
curl http://localhost:9090/metrics
```

Look for these metrics:
- `weather_crawl_requests_total`
- `weather_cache_hits_total`
- `fcm_notifications_sent_total`
- `scheduler_ticks_total`
- `go_goroutines`

## Step 6: Setup Prometheus (Optional)

### Install Prometheus

```bash
# macOS
brew install prometheus

# Linux
wget https://github.com/prometheus/prometheus/releases/download/v2.47.0/prometheus-2.47.0.linux-amd64.tar.gz
tar xvfz prometheus-*.tar.gz
cd prometheus-*
```

### Configure Prometheus

Create `prometheus.yml`:
```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'weather_collector'
    static_configs:
      - targets: ['localhost:9090']
```

### Start Prometheus

```bash
./prometheus --config.file=prometheus.yml
```

Access: http://localhost:9090 (Prometheus UI)

### Test Queries

In Prometheus UI, try:
```promql
rate(weather_crawl_requests_total[5m])
weather_cache_hits_total / (weather_cache_hits_total + weather_cache_misses_total)
rate(scheduler_ticks_total[5m])
```

## Step 7: Setup Grafana (Optional)

### Install Grafana

```bash
# macOS
brew install grafana

# Linux
wget https://dl.grafana.com/oss/release/grafana-10.0.0.linux-amd64.tar.gz
tar -zxvf grafana-*.tar.gz
cd grafana-*
```

### Start Grafana

```bash
# macOS
brew services start grafana

# Linux
./bin/grafana-server
```

Access: http://localhost:3000 (admin/admin)

### Import Dashboard

1. Login to Grafana
2. Go to Dashboards → Import
3. Upload `grafana-dashboard.json`
4. Select "Prometheus" as data source
5. Click Import

You now have 12 monitoring panels!

## Step 8: Test Metrics Collection

### Trigger Some Activity

```bash
# In another terminal, use the weather API
curl -X POST http://localhost:6000/v1/weather/alarm \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "region": "서울",
    "alarm_time": "08:00"
  }'
```

### Watch Metrics Change

```bash
# Watch metrics in real-time
watch -n 2 'curl -s http://localhost:9090/metrics | grep weather_crawl_requests_total'
```

### Check Logs

```bash
# Follow logs
tail -f /var/log/weather-collector.log

# Or if using stdout
./weather-scheduler 2>&1 | tee weather.log
```

## Common Commands

### Health Check
```bash
curl http://localhost:9090/health
```

### Get Specific Metric
```bash
curl http://localhost:9090/metrics | grep scheduler_ticks_total
```

### Count Cache Hits
```bash
curl -s http://localhost:9090/metrics | grep weather_cache_hits_total | awk '{print $2}'
```

### Check Scheduler Status
```bash
curl -s http://localhost:9090/health | jq '.components.scheduler'
```

### View Error Count
```bash
curl -s http://localhost:9090/health | jq '.error_count'
```

## Troubleshooting

### Metrics port already in use
```bash
# Change METRICS_PORT
export METRICS_PORT=9091
./weather-scheduler
```

### Can't connect to database
```bash
# Check MySQL is running
mysql -h localhost -u root -p

# Verify credentials in .env
cat .env | grep DB_
```

### Can't connect to Redis
```bash
# Check Redis is running
redis-cli ping

# Should return: PONG
```

### Health check shows unhealthy
```bash
# Check individual components
curl -s http://localhost:9090/health | jq '.components'

# Check logs for errors
tail -50 /var/log/weather-collector.log | grep error
```

### No metrics showing up
```bash
# Verify metrics endpoint works
curl http://localhost:9090/metrics | head -20

# Check Prometheus scrape config
cat prometheus.yml

# Check Prometheus targets
# Visit: http://localhost:9090/targets
```

## Next Steps

1. **Production Deployment**
   - Read `MONITORING.md` for full documentation
   - Configure log rotation
   - Setup alert rules
   - Configure backup strategies

2. **Alert Configuration**
   - Setup Prometheus Alert Manager
   - Configure notification channels (email, Slack)
   - Test alert delivery

3. **Dashboard Customization**
   - Modify `grafana-dashboard.json`
   - Add custom panels
   - Set alert thresholds

4. **Performance Tuning**
   - Monitor resource usage
   - Adjust cache TTL
   - Optimize database queries
   - Tune scheduler interval

## Resources

- Full Documentation: `MONITORING.md`
- Implementation Summary: `README.monitoring.md`
- Grafana Dashboard: `grafana-dashboard.json`
- Example Config: `.env.example`

## Support

If you encounter issues:

1. Check health endpoint: `curl http://localhost:9090/health`
2. Review logs: `tail -f /var/log/weather-collector.log`
3. Verify configuration: `cat .env`
4. Run tests: `go test ./pkg/...`

## Success Checklist

- [ ] All tests pass
- [ ] Scheduler starts without errors
- [ ] Health endpoint returns 200
- [ ] Metrics endpoint shows data
- [ ] Prometheus scrapes successfully
- [ ] Grafana dashboard displays
- [ ] Logs are properly formatted
- [ ] Database connection works
- [ ] Redis connection works
- [ ] FCM credentials valid

Congratulations! Your monitoring is now operational. 🎉
