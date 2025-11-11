# Troubleshooting Guide

Symptom-based troubleshooting for the Weather Data Collector Service.

**Version:** 1.0.0
**Last Updated:** 2025-11-11

## Table of Contents

- [Service Won't Start](#service-wont-start)
- [Alarms Not Processing](#alarms-not-processing)
- [Notifications Not Arriving](#notifications-not-arriving)
- [High Latency](#high-latency)
- [High Memory Usage](#high-memory-usage)
- [High CPU Usage](#high-cpu-usage)
- [Database Errors](#database-errors)
- [Redis Errors](#redis-errors)
- [FCM Errors](#fcm-errors)
- [Cache Issues](#cache-issues)

## Service Won't Start

### Symptom 1: Configuration Error on Startup

**Error Message:**
```
FATAL  Invalid configuration  error="DB_HOST is required"
```

**Diagnosis:**

1. Check environment variables:
   ```bash
   env | grep -E 'DB_|REDIS_|FCM_'
   ```

2. Check `.env` file exists:
   ```bash
   test -f .env && echo "Found" || echo "Missing"
   cat .env
   ```

**Solution:**

Create or fix `.env` file:
```bash
cp .env.example .env
vim .env
# Set required variables:
# DB_HOST=localhost
# DB_USER=root
# DB_NAME=joker
# REDIS_HOST=localhost
# FCM_CREDENTIALS_PATH=/path/to/credentials.json
```

---

### Symptom 2: FCM Credentials Not Found

**Error Message:**
```
FATAL  Failed to initialize FCM notifier  error="FCM credentials file not found at /path/to/credentials.json"
```

**Diagnosis:**

1. Check file exists:
   ```bash
   ls -l /path/to/firebase-credentials.json
   ```

2. Check path in environment:
   ```bash
   echo $FCM_CREDENTIALS_PATH
   ```

**Solution:**

**Option 1: Fix path**
```bash
export FCM_CREDENTIALS_PATH=/correct/path/to/firebase-credentials.json
```

**Option 2: Download credentials**
1. Go to Firebase Console → Project Settings → Service Accounts
2. Click "Generate New Private Key"
3. Save to expected location
4. Verify:
   ```bash
   cat $FCM_CREDENTIALS_PATH | jq .project_id
   ```

---

### Symptom 3: Database Connection Failed

**Error Message:**
```
FATAL  Failed to initialize database  error="failed to connect to database: dial tcp: connection refused"
```

**Diagnosis:**

1. Check MySQL is running:
   ```bash
   systemctl status mysql
   # or
   docker ps | grep mysql
   ```

2. Check connectivity:
   ```bash
   telnet $DB_HOST $DB_PORT
   # or
   nc -zv $DB_HOST $DB_PORT
   ```

3. Test credentials:
   ```bash
   mysql -h $DB_HOST -P $DB_PORT -u $DB_USER -p$DB_PASSWORD -e "SELECT 1"
   ```

**Solution:**

**If MySQL not running:**
```bash
sudo systemctl start mysql
# or
docker start mysql
```

**If wrong credentials:**
```bash
# Update .env
DB_USER=correct_user
DB_PASSWORD=correct_password

# Test again
mysql -h $DB_HOST -P $DB_PORT -u $DB_USER -p$DB_PASSWORD -e "SELECT 1"
```

**If firewall blocking:**
```bash
# Allow MySQL port
sudo firewall-cmd --add-port=3306/tcp --permanent
sudo firewall-cmd --reload
```

---

### Symptom 4: Redis Connection Failed

**Error Message:**
```
FATAL  Failed to connect to Redis  error="dial tcp: connection refused"
```

**Diagnosis:**

1. Check Redis is running:
   ```bash
   systemctl status redis
   # or
   docker ps | grep redis
   ```

2. Test connectivity:
   ```bash
   redis-cli -h $REDIS_HOST -p $REDIS_PORT ping
   ```

**Solution:**

**If Redis not running:**
```bash
sudo systemctl start redis
# or
docker start redis
```

**If wrong host/port:**
```bash
# Update .env
REDIS_HOST=correct_host
REDIS_PORT=6379

# Test again
redis-cli -h $REDIS_HOST -p $REDIS_PORT ping
```

---

## Alarms Not Processing

### Symptom 1: Scheduler Running but No Alarms Processed

**Symptoms:**
- Health endpoint shows `"scheduler": "healthy"`
- Metrics show `weather_alarms_processed_total = 0`
- No log entries for "Processing alarms"

**Diagnosis:**

1. Check scheduler status:
   ```bash
   curl http://localhost:9090/health | jq '.dependencies.scheduler'
   ```

2. Check for alarms in database:
   ```bash
   mysql -e "
   SELECT COUNT(*) as total,
          SUM(is_enabled) as enabled
   FROM weather_alarms
   "
   ```

3. Check alarms for current time:
   ```bash
   mysql -e "
   SELECT *
   FROM weather_alarms
   WHERE alarm_time = DATE_FORMAT(DATE_ADD(NOW(), INTERVAL 1 MINUTE), '%H:%i')
     AND is_enabled = true
   LIMIT 10
   "
   ```

**Solution:**

**No alarms in database:**
```sql
-- Create test alarm
INSERT INTO weather_alarms (user_id, region, alarm_time, is_enabled, created_at, updated_at)
VALUES (1, '서울', '09:00', true, NOW(), NOW());
```

**Alarms disabled:**
```sql
UPDATE weather_alarms SET is_enabled = true WHERE id = ?;
```

**Alarms already sent today:**
```sql
-- Reset last_sent to test
UPDATE weather_alarms SET last_sent = NULL WHERE id = ?;
```

**Wrong time format:**
```sql
-- Ensure HH:MM format (09:00, not 9:00)
UPDATE weather_alarms
SET alarm_time = LPAD(alarm_time, 5, '0')
WHERE LENGTH(alarm_time) = 4;
```

---

### Symptom 2: Alarms Found but Not Notifying

**Symptoms:**
- Logs show "Processing alarm"
- No FCM send logs
- Metrics show `weather_fcm_sent_total = 0`

**Diagnosis:**

1. Check logs for errors:
   ```bash
   journalctl -u weather-scheduler -n 100 | grep -i error
   ```

2. Check FCM tokens:
   ```bash
   mysql -e "
   SELECT u.id as user_id, COUNT(t.id) as token_count
   FROM weather_alarms a
   JOIN users u ON a.user_id = u.id
   LEFT JOIN weather_service_tokens t ON u.id = t.user_id
   WHERE a.is_enabled = true
   GROUP BY u.id
   "
   ```

**Solution:**

**No FCM tokens:**
```sql
-- Add test token
INSERT INTO weather_service_tokens (user_id, fcm_token, created_at)
VALUES (1, 'test_token_abc123', NOW());
```

**Invalid tokens:**
- Check logs for "Token send failed"
- Remove invalid tokens from database
- Users need to re-register

---

### Symptom 3: Alarms Processing Very Slowly

**Symptoms:**
- Metrics show `weather_processing_duration_seconds > 1.0`
- Delays between alarm time and notification
- Logs show slow operations

**Diagnosis:**

1. Check processing duration:
   ```bash
   curl http://localhost:9090/metrics | grep weather_processing_duration_seconds
   ```

2. Check cache hit rate:
   ```bash
   hits=$(curl -s http://localhost:9090/metrics | grep weather_cache_hits_total | awk '{print $2}')
   misses=$(curl -s http://localhost:9090/metrics | grep weather_cache_misses_total | awk '{print $2}')
   echo "Hit rate: $(echo "scale=2; $hits / ($hits + $misses) * 100" | bc)%"
   ```

3. Check database performance:
   ```bash
   mysql -e "SHOW PROCESSLIST"
   ```

**Solution:**

See [High Latency](#high-latency) section below.

---

## Notifications Not Arriving

### Symptom 1: FCM Send Success but No Notification

**Symptoms:**
- Logs show "Successfully processed alarm"
- Metrics show `weather_fcm_sent_total` increasing
- Users report no notifications

**Diagnosis:**

1. Check FCM response:
   ```bash
   journalctl -u weather-scheduler | grep "Batch completed"
   ```

2. Check notification payload:
   ```bash
   journalctl -u weather-scheduler | grep "Sending FCM notification" -A 10
   ```

3. Test FCM directly:
   ```bash
   # Use Firebase Console → Cloud Messaging → Send test message
   # Or use curl with FCM API
   ```

**Solution:**

**Invalid tokens:**
- Logs show individual token failures
- Clean up invalid tokens:
  ```sql
  DELETE FROM weather_service_tokens
  WHERE fcm_token IN (
    -- List of failed tokens from logs
  );
  ```

**App not handling notification:**
- Check mobile app notification handler
- Verify notification permissions enabled
- Check app is in foreground/background

**FCM quota exceeded:**
- Check Firebase Console → Usage & Quotas
- Implement rate limiting
- Optimize notification frequency

---

### Symptom 2: All FCM Sends Failing

**Error Message:**
```
ERROR  all notification sends failed: 150 failures
```

**Diagnosis:**

1. Check FCM credentials:
   ```bash
   cat $FCM_CREDENTIALS_PATH | jq .project_id
   ```

2. Check Firebase project:
   - Go to Firebase Console
   - Verify project ID matches credentials
   - Check Cloud Messaging is enabled

3. Check network connectivity:
   ```bash
   curl -I https://fcm.googleapis.com
   ```

**Solution:**

**Invalid credentials:**
1. Download new credentials from Firebase Console
2. Update credentials file
3. Restart service

**Network issues:**
```bash
# Check firewall
sudo iptables -L -n | grep 443

# Allow HTTPS
sudo firewall-cmd --add-service=https --permanent
sudo firewall-cmd --reload
```

**FCM API disabled:**
- Enable Cloud Messaging API in Firebase Console

---

## High Latency

### Symptom: Processing Taking > 500ms

**Symptoms:**
- Metrics show `weather_processing_duration_seconds > 0.5`
- Slow notification delivery
- Logs show delays

**Diagnosis:**

1. **Check cache hit rate:**
   ```bash
   curl http://localhost:9090/metrics | grep -E 'weather_cache_hits|weather_cache_misses'
   ```

2. **Check database latency:**
   ```bash
   mysql -e "SHOW PROCESSLIST"
   mysql -e "SHOW FULL PROCESSLIST WHERE Time > 1"
   ```

3. **Check Redis latency:**
   ```bash
   redis-cli --latency
   redis-cli --latency-history
   ```

4. **Check crawler latency:**
   ```bash
   journalctl -u weather-scheduler | grep "duration" | tail -20
   ```

**Solution:**

**Low cache hit rate (< 70%):**

1. Check Redis is working:
   ```bash
   redis-cli KEYS 'weather:*'
   ```

2. Check cache TTL:
   ```bash
   redis-cli TTL 'weather:서울'
   ```

3. Increase cache TTL (if acceptable):
   ```go
   // cache/weather.go
   const CacheTTL = 15 * time.Minute  // Increased from 10
   ```

4. Check Redis memory:
   ```bash
   redis-cli INFO memory
   ```

5. Increase Redis memory:
   ```bash
   # redis.conf
   maxmemory 4gb  # Increased from 2gb
   ```

**Slow database queries:**

1. Add indexes:
   ```sql
   CREATE INDEX idx_alarm_time_enabled
   ON weather_alarms(alarm_time, is_enabled);

   CREATE INDEX idx_last_sent
   ON weather_alarms(last_sent);
   ```

2. Optimize queries with EXPLAIN:
   ```sql
   EXPLAIN SELECT * FROM weather_alarms
   WHERE alarm_time = '09:00'
     AND is_enabled = true;
   ```

3. Increase connection pool:
   ```go
   sqlDB.SetMaxOpenConns(200)  // Increased from 100
   ```

**Slow crawler:**

1. Check Naver API response time:
   ```bash
   time curl -s "https://search.naver.com/search.naver?query=날씨 서울" > /dev/null
   ```

2. Reduce retries:
   ```go
   crawler := crawler.NewNaverWeatherCrawler(
       10*time.Second,  // timeout
       2,               // Reduced from 3 retries
   )
   ```

3. Increase timeout:
   ```go
   crawler := crawler.NewNaverWeatherCrawler(
       15*time.Second,  // Increased from 10s
       3,
   )
   ```

---

## High Memory Usage

### Symptom: Memory > 80%

**Symptoms:**
- Container/process using > 80% of allocated memory
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
   # Enable pprof (add to main.go)
   import _ "net/http/pprof"
   go func() {
       log.Println(http.ListenAndServe("localhost:6060", nil))
   }()

   # Check goroutines
   curl http://localhost:6060/debug/pprof/goroutine?debug=1
   ```

3. **Check heap allocation:**
   ```bash
   curl http://localhost:6060/debug/pprof/heap > heap.prof
   go tool pprof heap.prof
   ```

**Solution:**

**Memory leak:**

1. Profile with pprof:
   ```bash
   go tool pprof http://localhost:6060/debug/pprof/heap
   # Use 'top' command to see top allocations
   ```

2. Check for unclosed resources:
   ```bash
   # Check open files
   lsof -p $(pidof scheduler)
   ```

3. Restart service temporarily:
   ```bash
   sudo systemctl restart weather-scheduler
   ```

**Insufficient resources:**

1. Increase memory limit:
   ```bash
   # Docker
   docker update --memory=1g weather-scheduler

   # Kubernetes
   kubectl set resources deployment weather-scheduler \
     --limits=memory=1Gi
   ```

2. Vertical scaling (see RUNBOOK.md)

**Too many connections:**

1. Reduce database pool:
   ```go
   sqlDB.SetMaxOpenConns(50)  // Reduced from 100
   ```

2. Reduce Redis connections:
   ```go
   redisClient := redis.NewClient(&redis.Options{
       PoolSize: 5,  // Reduced from 10
   })
   ```

---

## High CPU Usage

### Symptom: CPU > 80%

**Symptoms:**
- High CPU usage
- Slow processing
- System unresponsive

**Diagnosis:**

1. **Check CPU usage:**
   ```bash
   top -p $(pidof scheduler)
   ```

2. **Check active goroutines:**
   ```bash
   curl http://localhost:6060/debug/pprof/goroutine?debug=1 | grep "^goroutine" | wc -l
   ```

3. **Profile CPU:**
   ```bash
   curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof
   go tool pprof cpu.prof
   ```

**Solution:**

**Too many alarms:**

1. Check alarm count:
   ```bash
   mysql -e "SELECT COUNT(*) FROM weather_alarms WHERE is_enabled=true"
   ```

2. Increase scheduler interval:
   ```bash
   SCHEDULER_INTERVAL=2m  # Reduced frequency
   ```

3. Vertical scaling (see RUNBOOK.md)

**Inefficient code:**

1. Profile and optimize hot paths
2. Batch operations where possible
3. Use goroutine pools to limit concurrency

---

## Database Errors

### Symptom: Connection Pool Exhausted

**Error Message:**
```
ERROR  failed to get alarms  error="pq: sorry, too many clients already"
```

**Solution:**

1. **Increase MySQL max_connections:**
   ```sql
   SET GLOBAL max_connections = 200;
   ```

2. **Reduce application pool:**
   ```go
   sqlDB.SetMaxOpenConns(50)
   ```

3. **Close unused connections:**
   ```go
   sqlDB.SetConnMaxLifetime(10 * time.Minute)
   ```

---

### Symptom: Query Timeout

**Error Message:**
```
ERROR  failed to get alarms  error="context deadline exceeded"
```

**Solution:**

1. **Add indexes:**
   ```sql
   CREATE INDEX idx_alarm_time ON weather_alarms(alarm_time);
   ```

2. **Optimize queries:**
   ```sql
   EXPLAIN SELECT * FROM weather_alarms WHERE alarm_time = '09:00';
   ```

3. **Increase context timeout:**
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
   ```

---

## Redis Errors

### Symptom: Connection Refused

**Error Message:**
```
ERROR  cache error  error="dial tcp: connection refused"
```

**Solution:**

See [Redis Connection Failed](#symptom-4-redis-connection-failed) above.

---

### Symptom: OOM Command Not Allowed

**Error Message:**
```
WARN  Failed to cache weather data  error="OOM command not allowed when used memory > 'maxmemory'"
```

**Solution:**

1. **Increase Redis memory:**
   ```bash
   # redis.conf
   maxmemory 4gb
   ```

2. **Enable eviction:**
   ```bash
   # redis.conf
   maxmemory-policy allkeys-lru
   ```

3. **Clear old keys:**
   ```bash
   redis-cli --scan --pattern 'weather:*' | xargs redis-cli DEL
   ```

---

## FCM Errors

### Symptom: Authentication Error

**Error Message:**
```
ERROR  FCM notification failed  error="failed to initialize Firebase app: credentials error"
```

**Solution:**

1. **Verify credentials format:**
   ```bash
   cat $FCM_CREDENTIALS_PATH | jq .
   ```

2. **Download new credentials:**
   - Firebase Console → Project Settings → Service Accounts
   - Generate New Private Key

3. **Restart service:**
   ```bash
   sudo systemctl restart weather-scheduler
   ```

---

### Symptom: Invalid Registration Token

**Error Message:**
```
WARN  Token send failed  error="invalid registration token"
```

**Solution:**

1. **Remove invalid token:**
   ```sql
   DELETE FROM weather_service_tokens WHERE fcm_token = 'invalid_token_here';
   ```

2. **User needs to re-register:**
   - Mobile app should request new token
   - Update database with new token

---

## Cache Issues

### Symptom: Low Cache Hit Rate

**Symptoms:**
- Metrics show hit rate < 70%
- High crawler usage
- Slow processing

**Diagnosis:**

```bash
hits=$(curl -s http://localhost:9090/metrics | grep weather_cache_hits_total | awk '{print $2}')
misses=$(curl -s http://localhost:9090/metrics | grep weather_cache_misses_total | awk '{print $2}')
echo "Hits: $hits, Misses: $misses"
echo "Hit rate: $(echo "scale=2; $hits / ($hits + $misses) * 100" | bc)%"
```

**Solution:**

1. **Increase TTL:**
   ```go
   const CacheTTL = 15 * time.Minute
   ```

2. **Check Redis is working:**
   ```bash
   redis-cli KEYS 'weather:*'
   redis-cli GET 'weather:서울'
   ```

3. **Warm cache proactively:**
   ```go
   // Pre-fetch popular regions
   popularRegions := []string{"서울", "부산", "대구"}
   for _, region := range popularRegions {
       data, _ := crawler.Fetch(ctx, region)
       cache.Set(ctx, region, data)
   }
   ```

---

## Version History

- **v1.0.0** (2025-11-11): Initial release
