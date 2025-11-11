# Weather Service Operations Runbook

Operational procedures and incident response guide for the Weather Data Collector service.

## Table of Contents

- [Service Overview](#service-overview)
- [Start/Stop Procedures](#startstop-procedures)
- [Scaling Guidelines](#scaling-guidelines)
- [Backup and Restore](#backup-and-restore)
- [Monitoring Setup](#monitoring-setup)
- [Incident Response](#incident-response)
- [Maintenance Windows](#maintenance-windows)
- [Disaster Recovery](#disaster-recovery)

## Service Overview

### Service Information

- **Service Name:** Weather Data Collector Scheduler
- **Purpose:** Automated weather data collection and user notifications
- **Type:** Background service (scheduler)
- **Dependencies:** MySQL, Redis, FCM
- **SLA:** 99.5% uptime

### Architecture

```
┌─────────────────┐
│   Scheduler     │ ← Metrics (9090)
│   (Go Service)  │
└────────┬────────┘
         │
    ┌────┴────┬──────────┬──────────┐
    │         │          │          │
    v         v          v          v
┌─────┐  ┌─────┐   ┌──────┐   ┌─────┐
│MySQL│  │Redis│   │Naver │   │ FCM │
└─────┘  └─────┘   │Weather│  └─────┘
                   └──────┘
```

### Critical Paths

1. **Alarm Check** → Database query
2. **Weather Fetch** → Naver crawler
3. **Cache Check** → Redis lookup
4. **Notification** → FCM send

## Start/Stop Procedures

### Starting the Service

#### Development

```bash
# Start dependencies
make docker-test-up

# Start scheduler
make dev-scheduler
```

#### Docker

```bash
# Start all services
docker-compose up -d

# Verify startup
docker-compose ps
docker-compose logs -f weather-scheduler
```

#### Systemd

```bash
# Start service
sudo systemctl start weather-scheduler

# Verify status
sudo systemctl status weather-scheduler

# Check logs
sudo journalctl -u weather-scheduler -f --since "5 minutes ago"
```

#### Kubernetes

```bash
# Apply deployment
kubectl apply -f deploy/k8s/

# Wait for ready
kubectl wait --for=condition=ready pod \
  -l app=weather-scheduler \
  -n joker-backend \
  --timeout=60s

# Verify
kubectl get pods -n joker-backend -l app=weather-scheduler
```

### Stopping the Service

#### Development

```bash
# Ctrl+C (sends SIGTERM)
# Or
kill -TERM <pid>
```

#### Docker

```bash
# Stop gracefully (30s timeout)
docker-compose stop weather-scheduler

# Force stop
docker-compose down
```

#### Systemd

```bash
# Stop service
sudo systemctl stop weather-scheduler

# Verify stopped
sudo systemctl status weather-scheduler
```

#### Kubernetes

```bash
# Scale down to zero
kubectl scale deployment weather-scheduler -n joker-backend --replicas=0

# Or delete deployment
kubectl delete deployment weather-scheduler -n joker-backend
```

### Restart Procedures

**Planned Restart:**

```bash
# Systemd
sudo systemctl restart weather-scheduler

# Docker
docker-compose restart weather-scheduler

# Kubernetes (rolling restart)
kubectl rollout restart deployment weather-scheduler -n joker-backend
```

**Emergency Restart:**

```bash
# Systemd (force)
sudo systemctl stop weather-scheduler
sleep 5
sudo systemctl start weather-scheduler

# Docker (recreate)
docker-compose down
docker-compose up -d

# Kubernetes (force delete)
kubectl delete pod -n joker-backend -l app=weather-scheduler --force
```

## Scaling Guidelines

### Vertical Scaling

Increase resources for single instance.

#### Systemd (Update resource limits)

```bash
# Edit service file
sudo systemctl edit weather-scheduler

# Add limits
[Service]
MemoryLimit=2G
CPUQuota=200%

# Reload and restart
sudo systemctl daemon-reload
sudo systemctl restart weather-scheduler
```

#### Kubernetes (Update resources)

```bash
# Edit deployment
kubectl edit deployment weather-scheduler -n joker-backend

# Update resources
resources:
  requests:
    cpu: 1000m
    memory: 512Mi
  limits:
    cpu: 2000m
    memory: 1Gi
```

### Horizontal Scaling

**Note:** Weather scheduler is designed for single-instance operation. Multi-instance requires distributed locking.

#### Current Limitation

The scheduler uses in-memory state and is not designed for horizontal scaling. Running multiple instances will cause:
- Duplicate notifications
- Race conditions
- Wasted resources

#### Future Scaling Strategy

To enable horizontal scaling:

1. Implement distributed locking (Redis)
2. Partition alarms by location/user
3. Add leader election
4. Coordinate job distribution

### Auto-Scaling (Kubernetes HPA)

**Not recommended** for scheduler service. HPA is configured but should maintain `minReplicas: 1, maxReplicas: 1`.

For genuine scaling needs, consider:
- Increase scheduler interval
- Implement job queue (RabbitMQ, Kafka)
- Split into worker pool architecture

### Performance Tuning

#### Database Connection Pool

```bash
# In main.go
sqlDB.SetMaxIdleConns(10)   # Increase for high load
sqlDB.SetMaxOpenConns(100)  # Increase for concurrency
sqlDB.SetConnMaxLifetime(time.Hour)
```

#### Scheduler Interval

```bash
# Lower load
SCHEDULER_INTERVAL=5m

# Higher frequency
SCHEDULER_INTERVAL=30s

# Balance: 1m (default)
```

#### Cache TTL

```bash
# Longer cache (less crawler load)
WEATHER_CACHE_TTL=1h

# Shorter cache (fresher data)
WEATHER_CACHE_TTL=15m

# Balance: 30m (default)
```

## Backup and Restore

### Database Backup

#### Manual Backup

```bash
# Full database
mysqldump -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME > backup_$(date +%Y%m%d).sql

# Alarms table only
mysqldump -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME alarms > alarms_backup.sql

# Compressed backup
mysqldump -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME | gzip > backup.sql.gz
```

#### Automated Backup (Cron)

```bash
# Add to crontab
0 2 * * * /usr/local/bin/backup-weather-db.sh

# backup-weather-db.sh
#!/bin/bash
BACKUP_DIR=/backups/weather
DATE=$(date +\%Y\%m\%d_\%H\%M\%S)
mysqldump -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME | gzip > $BACKUP_DIR/backup_$DATE.sql.gz
find $BACKUP_DIR -name "backup_*.sql.gz" -mtime +7 -delete
```

#### Backup Verification

```bash
# Test restore to temporary database
gunzip -c backup.sql.gz | mysql -h localhost -u root -p test_restore_db
mysql -h localhost -u root -p test_restore_db -e "SELECT COUNT(*) FROM alarms;"
```

### Database Restore

#### Full Restore

```bash
# Stop scheduler first
sudo systemctl stop weather-scheduler

# Restore
gunzip -c backup.sql.gz | mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME

# Verify
mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME -e "SELECT COUNT(*) FROM alarms;"

# Start scheduler
sudo systemctl start weather-scheduler
```

#### Partial Restore (Table)

```bash
# Extract table from backup
sed -n '/CREATE TABLE `alarms`/,/UNLOCK TABLES;/p' backup.sql > alarms_only.sql

# Restore table
mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME < alarms_only.sql
```

### Redis Backup

#### Backup

```bash
# Trigger BGSAVE
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD BGSAVE

# Copy RDB file
cp /var/lib/redis/dump.rdb /backups/redis/dump_$(date +%Y%m%d).rdb
```

#### Restore

```bash
# Stop Redis
sudo systemctl stop redis

# Replace RDB file
cp /backups/redis/dump_20251111.rdb /var/lib/redis/dump.rdb
chown redis:redis /var/lib/redis/dump.rdb

# Start Redis
sudo systemctl start redis
```

### Configuration Backup

```bash
# Backup configuration
cp /opt/weather-scheduler/.env /backups/config/.env_$(date +%Y%m%d)
cp /opt/weather-scheduler/config/fcm-credentials.json /backups/config/
```

## Monitoring Setup

### Prometheus Configuration

Add to `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'weather-scheduler'
    static_configs:
      - targets: ['localhost:9090']
    metric_relabel_configs:
      - source_labels: [__name__]
        regex: 'weather_.*'
        action: keep
```

### Grafana Dashboard

Import dashboard from `grafana-dashboard.json`:

1. Open Grafana
2. Click "+" → Import
3. Upload `grafana-dashboard.json`
4. Select Prometheus data source
5. Click Import

### Key Metrics

| Metric | Alert Threshold | Action |
|--------|----------------|--------|
| `weather_scheduler_errors_total` | > 10/min | Check logs |
| `weather_cache_hit_rate` | < 50% | Review cache TTL |
| `weather_crawler_duration_seconds` | > 30s | Check network |
| `weather_notifications_failed_total` | > 5/min | Check FCM |

### Alert Rules

Create `alert_rules.yml`:

```yaml
groups:
  - name: weather_scheduler
    interval: 30s
    rules:
      - alert: WeatherSchedulerDown
        expr: up{job="weather-scheduler"} == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Weather scheduler is down"

      - alert: HighErrorRate
        expr: rate(weather_scheduler_errors_total[5m]) > 0.1
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High error rate in weather scheduler"

      - alert: LowCacheHitRate
        expr: weather_cache_hit_rate < 0.5
        for: 30m
        labels:
          severity: info
        annotations:
          summary: "Cache hit rate below 50%"
```

### Health Check Monitoring

```bash
# Add to monitoring script
#!/bin/bash
STATUS=$(curl -s http://localhost:9090/health | jq -r '.status')
if [ "$STATUS" != "healthy" ]; then
    echo "ALERT: Weather scheduler unhealthy"
    # Send alert
fi
```

## Incident Response

### Severity Levels

- **P0 (Critical):** Service down, no notifications
- **P1 (High):** Degraded performance, some failures
- **P2 (Medium):** Minor issues, no user impact
- **P3 (Low):** Cosmetic issues, future improvements

### Incident Response Procedures

#### P0: Service Down

**Symptoms:**
- Health check failing
- No metrics reported
- Service not running

**Immediate Actions:**

1. **Check service status:**
   ```bash
   sudo systemctl status weather-scheduler
   curl http://localhost:9090/health
   ```

2. **Review logs:**
   ```bash
   sudo journalctl -u weather-scheduler -n 100
   ```

3. **Attempt restart:**
   ```bash
   sudo systemctl restart weather-scheduler
   ```

4. **If restart fails:**
   - Check configuration validation
   - Verify database connectivity
   - Verify Redis connectivity
   - Check FCM credentials
   - Review disk space

5. **Escalation:**
   - If service doesn't recover in 5 minutes
   - Contact on-call engineer
   - Prepare for rollback

#### P1: High Error Rate

**Symptoms:**
- `weather_scheduler_errors_total` increasing
- Some notifications failing
- Slow response times

**Actions:**

1. **Identify error pattern:**
   ```bash
   sudo journalctl -u weather-scheduler | grep ERROR | tail -50
   ```

2. **Check dependencies:**
   ```bash
   # Database
   mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD -e "SELECT 1"

   # Redis
   redis-cli -h $REDIS_HOST ping

   # Crawler
   curl -I https://weather.naver.com
   ```

3. **Review metrics:**
   ```bash
   curl http://localhost:9090/metrics | grep error
   ```

4. **Mitigation:**
   - Increase scheduler interval (reduce load)
   - Clear Redis cache if corrupted
   - Restart service if memory leak

#### P2: Performance Degradation

**Symptoms:**
- Slow metrics endpoint
- High memory usage
- Slow notifications

**Actions:**

1. **Monitor resources:**
   ```bash
   # Memory
   ps aux | grep scheduler

   # Goroutines
   curl http://localhost:9090/debug/pprof/goroutine?debug=1
   ```

2. **Check database queries:**
   ```bash
   mysql -e "SHOW PROCESSLIST;"
   ```

3. **Review cache performance:**
   ```bash
   redis-cli INFO stats
   ```

4. **Optimization:**
   - Increase cache TTL
   - Optimize database indexes
   - Review scheduler interval

### Common Issues

#### Issue: FCM Notifications Not Sending

**Diagnosis:**
```bash
# Check FCM logs
journalctl -u weather-scheduler | grep FCM

# Verify credentials
cat $FCM_CREDENTIALS_PATH | jq .
```

**Solutions:**
1. Verify FCM credentials validity
2. Check device token format
3. Test with FCM console
4. Review FCM quotas

#### Issue: Database Connection Pool Exhausted

**Diagnosis:**
```bash
# Check connections
mysql -e "SHOW STATUS LIKE 'Threads_connected';"

# Review pool settings
grep -i "max.*conn" main.go
```

**Solutions:**
1. Increase pool size
2. Reduce scheduler frequency
3. Fix connection leaks
4. Increase MySQL max_connections

#### Issue: Redis Cache Misses

**Diagnosis:**
```bash
# Cache stats
redis-cli INFO stats | grep -i hit

# TTL check
redis-cli --scan --pattern "weather:*" | xargs -L1 redis-cli TTL
```

**Solutions:**
1. Increase cache TTL
2. Review cache key patterns
3. Verify Redis persistence
4. Check memory limits

## Maintenance Windows

### Planned Maintenance

**Preparation:**
1. Schedule during low-traffic period
2. Notify stakeholders 24h advance
3. Prepare rollback plan
4. Backup database and configuration

**Procedure:**

```bash
# 1. Stop service
sudo systemctl stop weather-scheduler

# 2. Perform maintenance
# - Update binaries
# - Update configuration
# - Database migrations
# - Dependency updates

# 3. Verify health
curl http://localhost:9090/health

# 4. Start service
sudo systemctl start weather-scheduler

# 5. Monitor for 15 minutes
sudo journalctl -u weather-scheduler -f
```

### Database Migrations

```bash
# Run migrations
migrate -path ./migrations -database "mysql://$DB_USER:$DB_PASSWORD@tcp($DB_HOST:$DB_PORT)/$DB_NAME" up

# Verify
mysql -e "SELECT * FROM schema_migrations;"

# Rollback if needed
migrate -path ./migrations -database "..." down 1
```

### Configuration Updates

```bash
# Backup current config
cp .env .env.backup

# Update configuration
vim .env

# Test configuration
./scheduler --validate-config

# Apply changes
sudo systemctl restart weather-scheduler

# Monitor
sudo journalctl -u weather-scheduler -f
```

## Disaster Recovery

### Complete Service Failure

**Recovery Steps:**

1. **Assess damage:**
   - Database state
   - Redis state
   - Service state

2. **Restore from backup:**
   ```bash
   # Database
   gunzip -c backup_latest.sql.gz | mysql -h $DB_HOST -u $DB_USER -p $DB_NAME

   # Redis
   cp backup_latest.rdb /var/lib/redis/dump.rdb
   sudo systemctl restart redis

   # Configuration
   cp .env.backup /opt/weather-scheduler/.env
   ```

3. **Redeploy service:**
   ```bash
   # Reinstall binary
   cp bin/scheduler_backup /opt/weather-scheduler/bin/scheduler

   # Restart
   sudo systemctl restart weather-scheduler
   ```

4. **Verify recovery:**
   ```bash
   curl http://localhost:9090/health
   sudo journalctl -u weather-scheduler -f
   ```

### Data Corruption

**Recovery:**

1. **Identify corrupted data:**
   ```sql
   SELECT * FROM alarms WHERE updated_at > NOW() - INTERVAL 1 HOUR;
   ```

2. **Restore from point-in-time backup:**
   ```bash
   # Find backup before corruption
   ls -lt /backups/weather/

   # Restore specific table
   gunzip -c backup_pre_corruption.sql.gz | mysql $DB_NAME
   ```

3. **Verify data integrity:**
   ```sql
   SELECT COUNT(*) FROM alarms WHERE active = 1;
   ```

### Regional Outage

**Multi-region setup (future):**

1. DNS failover to secondary region
2. Restore database from replication
3. Deploy service in secondary region
4. Monitor recovery

## Runbook Checklist

### Daily Operations

- [ ] Check service health
- [ ] Review error logs
- [ ] Monitor metrics dashboard
- [ ] Verify backup completion

### Weekly Operations

- [ ] Review performance metrics
- [ ] Check disk space
- [ ] Test backup restoration
- [ ] Update dependencies

### Monthly Operations

- [ ] Rotate logs
- [ ] Review alert thresholds
- [ ] Performance optimization review
- [ ] Security updates

---

**Last Updated:** 2025-11-11
**Version:** 1.0.0
**On-Call:** [Your Team Contact]
