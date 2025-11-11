# Operations Runbook

Operational procedures for the Weather Data Collector Service.

**Version:** 1.0.0
**Last Updated:** 2025-11-11

## Table of Contents

- [Restart Procedures](#restart-procedures)
- [Scaling Strategies](#scaling-strategies)
- [Debugging Steps](#debugging-steps)
- [Common Issues](#common-issues)
- [Monitoring Checklist](#monitoring-checklist)
- [Emergency Contacts](#emergency-contacts)

## Restart Procedures

### Graceful Restart

**Use Case:** Normal maintenance, configuration updates

**Steps:**

1. **Check Service Health**
   ```bash
   curl http://localhost:9090/health
   # Expected: {"status": "healthy"}
   ```

2. **Send Stop Signal (SIGTERM)**
   ```bash
   # Systemd
   sudo systemctl stop weather-scheduler

   # Docker
   docker stop --time=30 weather-scheduler

   # Kubernetes
   kubectl delete pod -l app=weather-scheduler

   # Direct process
   kill -SIGTERM $(pidof scheduler)
   ```

3. **Wait for Shutdown (max 30 seconds)**
   ```bash
   # Monitor logs for shutdown message
   tail -f /var/log/weather-scheduler.log | grep "shutdown"
   # Expected: "Shutdown completed gracefully"
   ```

4. **Verify No In-Flight Operations**
   ```bash
   # Check metrics for pending alarms
   curl http://localhost:9090/metrics | grep weather_alarms_processing
   # Expected: 0
   ```

5. **Start Service**
   ```bash
   # Systemd
   sudo systemctl start weather-scheduler

   # Docker
   docker start weather-scheduler

   # Kubernetes (new pod auto-created)
   # kubectl delete pod triggers new pod

   # Direct binary
   ./scheduler &
   ```

6. **Verify Health**
   ```bash
   # Wait 10 seconds for initialization
   sleep 10

   # Check health
   curl http://localhost:9090/health

   # Check metrics
   curl http://localhost:9090/metrics | grep weather_alarms_processed_total
   ```

**Timeline:** 45-60 seconds total

---

### Emergency Restart

**Use Case:** Service unresponsive, memory leak, crash loop

**Steps:**

1. **Force Stop (SIGKILL)**
   ```bash
   # Systemd
   sudo systemctl kill --signal=SIGKILL weather-scheduler

   # Docker
   docker kill weather-scheduler

   # Kubernetes
   kubectl delete pod -l app=weather-scheduler --grace-period=0 --force

   # Direct process
   kill -9 $(pidof scheduler)
   ```

2. **Verify Process Terminated**
   ```bash
   ps aux | grep scheduler
   # Expected: no process
   ```

3. **Check for Resource Leaks**
   ```bash
   # Check open connections
   lsof -p $(pidof scheduler) 2>/dev/null || echo "Process terminated"

   # Check memory usage
   free -h
   ```

4. **Start Service Immediately**
   ```bash
   sudo systemctl start weather-scheduler
   ```

5. **Monitor Logs for Errors**
   ```bash
   journalctl -u weather-scheduler -f --since "1 minute ago"
   ```

**Timeline:** 10-20 seconds total

**Note:** Emergency restart may drop in-flight alarms. Check logs for failed alarms and manually retry if critical.

---

### Rollback Procedure

**Use Case:** New version causes issues

**Steps:**

1. **Identify Last Working Version**
   ```bash
   # Docker
   docker images | grep weather-scheduler

   # Git
   git log --oneline -10
   ```

2. **Stop Current Version**
   ```bash
   sudo systemctl stop weather-scheduler
   ```

3. **Deploy Previous Version**
   ```bash
   # Docker
   docker stop weather-scheduler
   docker rm weather-scheduler
   docker run -d --name weather-scheduler weather-scheduler:1.0.0

   # Kubernetes
   kubectl set image deployment/weather-scheduler scheduler=weather-scheduler:1.0.0
   kubectl rollout status deployment/weather-scheduler

   # Binary
   cp /opt/backups/scheduler-1.0.0 /usr/local/bin/scheduler
   sudo systemctl start weather-scheduler
   ```

4. **Verify Rollback**
   ```bash
   curl http://localhost:9090/health | jq '.version'
   # Expected: "1.0.0"
   ```

5. **Check Metrics**
   ```bash
   # Wait 2 minutes
   sleep 120

   # Verify alarms processing
   curl http://localhost:9090/metrics | grep weather_alarms_processed_total
   ```

**Timeline:** 2-5 minutes

---

### Post-Restart Verification

**Checklist:**

- [ ] Health endpoint returns 200 OK
- [ ] Version matches expected
- [ ] Database connection healthy
- [ ] Redis connection healthy
- [ ] Scheduler running
- [ ] Metrics updating
- [ ] Alarms processing within 2 minutes
- [ ] No errors in logs
- [ ] FCM notifications sending

**Commands:**
```bash
# Health
curl http://localhost:9090/health

# Version
curl http://localhost:9090/health | jq '.version'

# Dependencies
curl http://localhost:9090/health | jq '.dependencies'

# Metrics
curl http://localhost:9090/metrics | grep -E 'weather_alarms_processed|weather_fcm_sent'

# Logs (last 50 lines)
journalctl -u weather-scheduler -n 50 --no-pager

# Test alarm processing (wait 2 minutes)
sleep 120 && curl http://localhost:9090/metrics | grep weather_alarms_processed_total
```

---

## Scaling Strategies

### Vertical Scaling

**Increase Resources (CPU/Memory)**

**When to Scale:**
- CPU usage > 80% consistently
- Memory usage > 80% consistently
- Processing latency > 500ms

**How to Scale:**

**Docker:**
```bash
docker update --cpus=2 --memory=1g weather-scheduler
docker restart weather-scheduler
```

**Kubernetes:**
```yaml
# k8s/deployment.yaml
resources:
  requests:
    cpu: 200m      # Increased from 100m
    memory: 256Mi  # Increased from 128Mi
  limits:
    cpu: 1000m     # Increased from 500m
    memory: 1Gi    # Increased from 512Mi
```
```bash
kubectl apply -f k8s/deployment.yaml
```

**Systemd (VM):**
```bash
# Edit cgroup limits in /etc/systemd/system/weather-scheduler.service
CPUQuota=200%      # Increased from 100%
MemoryMax=1G       # Increased from 512M

sudo systemctl daemon-reload
sudo systemctl restart weather-scheduler
```

**Recommended Thresholds:**

| Alarms/Min | CPU | Memory |
|------------|-----|--------|
| < 100 | 100m | 128Mi |
| 100-500 | 200m | 256Mi |
| 500-1000 | 500m | 512Mi |
| > 1000 | 1000m | 1Gi |

---

### Horizontal Scaling

**Multiple Scheduler Instances**

**Warning:** Multiple instances will process same alarms. Implement distributed locking or database transactions to prevent duplicate notifications.

**Prerequisites:**
- Distributed locking (Redis SETNX or database transactions)
- Shared state (database, not in-memory)

**Implementation Options:**

**Option 1: Database Row Locking**
```sql
-- Use SELECT FOR UPDATE
SELECT * FROM weather_alarms
WHERE alarm_time = ?
  AND is_enabled = true
FOR UPDATE SKIP LOCKED
```

**Option 2: Redis Distributed Lock**
```go
// Acquire lock before processing alarm
lockKey := fmt.Sprintf("lock:alarm:%d", alarm.ID)
ok, err := redisClient.SetNX(ctx, lockKey, "1", 60*time.Second).Result()
if !ok {
    // Another instance is processing this alarm
    return nil
}
defer redisClient.Del(ctx, lockKey)
```

**Deployment:**
```yaml
# k8s/deployment.yaml
replicas: 3  # Increased from 1
```

**Not Recommended Unless:**
- Alarm volume > 1000/minute
- Single instance can't handle load
- HA requirement > duplicate send prevention

---

### Database Scaling

**Read Replicas**

**When to Scale:**
- Database CPU > 70%
- Query latency > 50ms
- Connection pool exhaustion

**Strategy:**
```go
// Use read replica for GetAlarmsToNotify, GetFCMTokens
// Use master for UpdateLastSent

masterDB := gorm.Open(mysql.Open(masterDSN))
replicaDB := gorm.Open(mysql.Open(replicaDSN))

repo := &SchedulerWeatherRepository{
    readDB:  replicaDB,  // SELECT queries
    writeDB: masterDB,   // UPDATE queries
}
```

**Connection Pool Tuning:**
```go
sqlDB, _ := db.DB()
sqlDB.SetMaxIdleConns(20)   // Increased from 10
sqlDB.SetMaxOpenConns(200)  // Increased from 100
sqlDB.SetConnMaxLifetime(30 * time.Minute)
```

---

### Redis Scaling

**Redis Cluster**

**When to Scale:**
- Cache hit rate < 50%
- Redis CPU > 70%
- Memory usage > 80%

**Strategy:**
```go
import "github.com/redis/go-redis/v9"

// Change from single instance to cluster
redisClient := redis.NewClusterClient(&redis.ClusterOptions{
    Addrs: []string{
        "redis-node-1:6379",
        "redis-node-2:6379",
        "redis-node-3:6379",
    },
})
```

**Memory Optimization:**
```bash
# Redis config
maxmemory 2gb
maxmemory-policy allkeys-lru
```

---

### Performance Tuning

**Optimize Scheduler Interval**

**Trade-off:** Lower interval = lower latency but higher resource usage

```bash
# .env
SCHEDULER_INTERVAL=30s  # Reduced from 1m (not recommended)
SCHEDULER_INTERVAL=2m   # Increased from 1m (lower precision)
```

**Optimize Cache TTL**

**Trade-off:** Longer TTL = fewer crawler calls but staler data

```go
// cache/weather.go
const CacheTTL = 15 * time.Minute  // Increased from 10 minutes
```

**Batch Database Queries**

```go
// Instead of N queries for N alarms
for _, alarm := range alarms {
    tokens, _ := repo.GetFCMTokens(ctx, alarm.UserID)
}

// Use single query with IN clause
userIDs := extractUserIDs(alarms)
tokenMap, _ := repo.GetFCMTokensBatch(ctx, userIDs)
```

---

## Debugging Steps

### Step 1: Check Service Status

```bash
# Health endpoint
curl http://localhost:9090/health

# Systemd status
sudo systemctl status weather-scheduler

# Docker status
docker ps | grep weather-scheduler

# Kubernetes status
kubectl get pods -l app=weather-scheduler
kubectl describe pod -l app=weather-scheduler
```

**Expected Output:**
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

---

### Step 2: Check Logs

```bash
# Systemd
journalctl -u weather-scheduler -f --since "10 minutes ago"

# Docker
docker logs -f --tail=100 weather-scheduler

# Kubernetes
kubectl logs -l app=weather-scheduler -f --tail=100

# File-based
tail -f /var/log/weather-scheduler.log
```

**Look for:**
- Error messages
- Stack traces
- Connection failures
- Processing delays

---

### Step 3: Check Metrics

```bash
# All metrics
curl http://localhost:9090/metrics

# Specific metrics
curl http://localhost:9090/metrics | grep -E 'weather_alarms_processed|weather_alarms_failed|weather_cache_hits|weather_fcm_sent'
```

**Key Metrics:**
- `weather_alarms_processed_total`: Should increment every minute
- `weather_alarms_failed_total`: Should be < 10% of processed
- `weather_cache_hits_total / weather_cache_misses_total`: Hit rate should be > 70%
- `weather_fcm_sent_total`: Should match alarms processed
- `weather_fcm_failed_total`: Should be < 5% of sent

---

### Step 4: Check Database Connection

```bash
# Test database connectivity
mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD -e "SELECT 1"

# Check alarm configuration
mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME -e "
  SELECT COUNT(*) as total_alarms,
         SUM(is_enabled) as enabled_alarms,
         COUNT(DISTINCT user_id) as unique_users
  FROM weather_alarms
"

# Check alarms for current time
mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME -e "
  SELECT id, user_id, region, alarm_time, is_enabled, last_sent
  FROM weather_alarms
  WHERE alarm_time = DATE_FORMAT(NOW(), '%H:%i')
    AND is_enabled = true
  LIMIT 10
"
```

---

### Step 5: Check Redis Connection

```bash
# Test Redis connectivity
redis-cli -h $REDIS_HOST -p $REDIS_PORT ping
# Expected: PONG

# Check cached data
redis-cli -h $REDIS_HOST -p $REDIS_PORT KEYS 'weather:*'

# Check specific region cache
redis-cli -h $REDIS_HOST -p $REDIS_PORT GET 'weather:서울'

# Check TTL
redis-cli -h $REDIS_HOST -p $REDIS_PORT TTL 'weather:서울'
# Expected: 0-600 seconds
```

---

### Step 6: Check FCM Credentials

```bash
# Verify credentials file exists
test -f $FCM_CREDENTIALS_PATH && echo "Credentials found" || echo "Credentials missing"

# Validate JSON format
cat $FCM_CREDENTIALS_PATH | jq . > /dev/null && echo "Valid JSON" || echo "Invalid JSON"

# Check file permissions
ls -l $FCM_CREDENTIALS_PATH
# Expected: -rw------- (600) or -rw-r--r-- (644)
```

---

### Step 7: Test Individual Components

**Test Crawler:**
```bash
# Run crawler test
cd services/weatherService
go test -v ./features/weather/crawler/... -run TestNaverWeatherCrawler_Fetch
```

**Test Cache:**
```bash
# Run cache test
go test -v ./features/weather/cache/... -run TestWeatherCache
```

**Test Repository:**
```bash
# Run repository test
go test -v ./features/weather/repository/... -run TestSchedulerWeatherRepository
```

---

## Common Issues

### Issue: Alarms Not Processing

**Symptoms:**
- Scheduler running but no alarms processed
- Metrics show `weather_alarms_processed_total = 0`
- No log entries for "Processing alarms"

**Diagnosis:**

1. **Check scheduler status:**
   ```bash
   curl http://localhost:9090/health | jq '.dependencies.scheduler'
   ```

2. **Check alarm configuration:**
   ```bash
   mysql -e "SELECT * FROM weather_alarms WHERE is_enabled=1 LIMIT 10"
   ```

3. **Verify alarm timing:**
   ```bash
   # Current time
   date '+%H:%M'

   # Alarms for current time + 1 minute
   mysql -e "SELECT * FROM weather_alarms WHERE alarm_time = DATE_FORMAT(DATE_ADD(NOW(), INTERVAL 1 MINUTE), '%H:%i')"
   ```

**Solutions:**

- **No alarms in database:**
  ```sql
  INSERT INTO weather_alarms (user_id, region, alarm_time, is_enabled)
  VALUES (1, '서울', '09:00', true);
  ```

- **Alarms disabled:**
  ```sql
  UPDATE weather_alarms SET is_enabled = true WHERE id = ?;
  ```

- **Already sent today:**
  ```sql
  UPDATE weather_alarms SET last_sent = NULL WHERE id = ?;
  ```

- **Wrong time format:**
  ```sql
  -- Ensure HH:MM format (09:00, not 9:00)
  UPDATE weather_alarms SET alarm_time = '09:00' WHERE alarm_time = '9:00';
  ```

---

### Issue: High Latency

**Symptoms:**
- Processing duration > 500ms
- Metrics show `weather_processing_duration_seconds > 0.5`
- Slow response times

**Diagnosis:**

1. **Check cache hit rate:**
   ```bash
   curl http://localhost:9090/metrics | grep -E 'weather_cache_hits|weather_cache_misses'
   # Calculate hit rate: hits / (hits + misses)
   ```

2. **Check database query performance:**
   ```bash
   mysql -e "SHOW PROCESSLIST"
   mysql -e "SHOW FULL PROCESSLIST WHERE Time > 1"
   ```

3. **Check Redis latency:**
   ```bash
   redis-cli --latency
   redis-cli --latency-history
   ```

**Solutions:**

- **Low cache hit rate (< 70%):**
  - Increase cache TTL
  - Check Redis memory limits
  - Verify cache is working: `redis-cli KEYS 'weather:*'`

- **Slow database queries:**
  - Add indexes:
    ```sql
    CREATE INDEX idx_alarm_time_enabled ON weather_alarms(alarm_time, is_enabled);
    CREATE INDEX idx_last_sent ON weather_alarms(last_sent);
    ```
  - Optimize queries with EXPLAIN
  - Increase connection pool

- **Slow crawler:**
  - Check Naver API response time
  - Reduce retries: `maxRetries = 2`
  - Increase timeout: `timeout = 15 * time.Second`

---

### Issue: FCM Notifications Not Sending

**Symptoms:**
- Metrics show `weather_fcm_failed_total` increasing
- Users not receiving notifications
- Logs show FCM errors

**Diagnosis:**

1. **Check FCM credentials:**
   ```bash
   test -f $FCM_CREDENTIALS_PATH && echo "Found" || echo "Missing"
   cat $FCM_CREDENTIALS_PATH | jq .project_id
   ```

2. **Check FCM tokens:**
   ```bash
   mysql -e "SELECT COUNT(*) FROM weather_service_tokens"
   mysql -e "SELECT * FROM weather_service_tokens LIMIT 5"
   ```

3. **Check logs for FCM errors:**
   ```bash
   journalctl -u weather-scheduler | grep -i fcm | tail -20
   ```

**Solutions:**

- **Invalid credentials:**
  - Download new credentials from Firebase Console
  - Update `FCM_CREDENTIALS_PATH`
  - Restart service

- **No tokens:**
  ```sql
  INSERT INTO weather_service_tokens (user_id, fcm_token)
  VALUES (1, 'valid_fcm_token_here');
  ```

- **Invalid tokens:**
  - FCM returns error for invalid tokens
  - Clean up invalid tokens from database
  - Users need to re-register

- **FCM quota exceeded:**
  - Check Firebase Console → Cloud Messaging
  - Implement rate limiting
  - Batch sends more efficiently

---

### Issue: High Memory Usage

**Symptoms:**
- Memory usage > 80%
- OOM kills
- Slow performance

**Diagnosis:**

1. **Check memory usage:**
   ```bash
   # Container
   docker stats weather-scheduler --no-stream

   # Process
   ps aux | grep scheduler
   top -p $(pidof scheduler)
   ```

2. **Check for goroutine leaks:**
   ```bash
   # Enable pprof in scheduler
   curl http://localhost:6060/debug/pprof/goroutine?debug=1
   ```

**Solutions:**

- **Memory leak:**
  - Check for goroutine leaks
  - Profile with pprof
  - Restart service temporarily

- **Insufficient resources:**
  - Increase memory limit
  - Vertical scaling (see above)

- **Too many connections:**
  - Reduce database connection pool
  - Reduce Redis connections

---

### Issue: Database Connection Errors

**Symptoms:**
- Logs show "failed to connect to database"
- Health check shows database unhealthy
- Errors: "connection refused", "too many connections"

**Diagnosis:**

1. **Test connectivity:**
   ```bash
   mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD -e "SELECT 1"
   ```

2. **Check connection count:**
   ```bash
   mysql -e "SHOW STATUS LIKE 'Threads_connected'"
   mysql -e "SHOW VARIABLES LIKE 'max_connections'"
   ```

**Solutions:**

- **Connection refused:**
  - Check MySQL is running: `systemctl status mysql`
  - Check firewall: `telnet $DB_HOST 3306`
  - Verify credentials

- **Too many connections:**
  - Reduce connection pool: `SetMaxOpenConns(50)`
  - Increase MySQL max_connections: `SET GLOBAL max_connections=200`
  - Close unused connections

- **Connection timeout:**
  - Increase `ConnMaxLifetime`
  - Check network latency
  - Use connection pooling

---

## Monitoring Checklist

### Real-Time Monitoring

Check every 5 minutes:

- [ ] Service health: `curl http://localhost:9090/health`
- [ ] Alarms processing: `weather_alarms_processed_total` incrementing
- [ ] Error rate: `weather_alarms_failed_total < 10% of processed`
- [ ] Cache hit rate: `> 70%`
- [ ] FCM success rate: `> 95%`

### Daily Monitoring

Check daily:

- [ ] Total alarms processed: Compare to baseline
- [ ] Average processing latency: `< 500ms`
- [ ] Database connection pool usage: `< 80%`
- [ ] Redis memory usage: `< 80%`
- [ ] Log file size: `< 1GB`
- [ ] Disk space: `> 20% free`

### Weekly Monitoring

Check weekly:

- [ ] Alarm growth rate: Track user adoption
- [ ] Cache performance trends: Hit rate stable
- [ ] FCM token growth: Correlates with users
- [ ] Error trends: No increasing error rate
- [ ] Resource usage trends: Plan for scaling

---

### Key Metrics

| Metric | Threshold | Action |
|--------|-----------|--------|
| `weather_alarms_processed_total` | Incrementing every minute | Alert if static for 5 minutes |
| `weather_cache_hit_rate` | > 70% | Alert if < 50%, investigate cache |
| `weather_fcm_success_rate` | > 95% | Alert if < 90%, check tokens |
| `weather_processing_duration_seconds` | < 0.5s | Alert if > 1s, investigate latency |
| `weather_alarms_failed_total` | < 10/min | Alert if > 20/min, investigate errors |
| CPU usage | < 80% | Alert if > 90%, scale vertically |
| Memory usage | < 80% | Alert if > 90%, scale vertically |
| Database connections | < 80 | Alert if > 90, increase pool |

---

## Emergency Contacts

### On-Call Rotation

- **Primary:** [Name] - [Phone] - [Email]
- **Secondary:** [Name] - [Phone] - [Email]
- **Manager:** [Name] - [Phone] - [Email]

### Escalation Path

1. **Level 1:** On-call engineer (0-15 minutes)
2. **Level 2:** Team lead (15-30 minutes)
3. **Level 3:** Engineering manager (30-60 minutes)

### External Dependencies

- **Naver Weather:** No direct contact (public API)
- **Firebase/FCM:** https://firebase.google.com/support
- **MySQL Support:** [DBA Team Contact]
- **Redis Support:** [Infrastructure Team Contact]

---

## Version History

- **v1.0.0** (2025-11-11): Initial release
