# Configuration Reference

Complete configuration documentation for the Weather Data Collector Service.

**Version:** 1.0.0
**Last Updated:** 2025-11-11

## Table of Contents

- [Environment Variables](#environment-variables)
- [Database Configuration](#database-configuration)
- [Redis Configuration](#redis-configuration)
- [Firebase Configuration](#firebase-configuration)
- [Scheduler Configuration](#scheduler-configuration)
- [Logging Configuration](#logging-configuration)
- [Metrics Configuration](#metrics-configuration)
- [Configuration Examples](#configuration-examples)

## Environment Variables

### Overview

All configuration is managed through environment variables for 12-factor app compliance.

**Configuration File:** `.env` (development) or Kubernetes ConfigMap/Secret (production)

### Required Variables

#### DB_HOST

- **Required:** Yes
- **Type:** String
- **Default:** None
- **Description:** MySQL database hostname or IP address
- **Example:** `localhost`, `mysql.example.com`, `10.0.1.100`
- **Validation:** Must be reachable on DB_PORT

**Usage:**
```bash
DB_HOST=localhost           # Local development
DB_HOST=mysql.example.com   # Production
DB_HOST=10.0.1.100         # IP address
```

---

#### DB_PORT

- **Required:** No
- **Type:** Integer
- **Default:** `3306`
- **Description:** MySQL database port
- **Example:** `3306`, `3307`
- **Validation:** Must be 1-65535

**Usage:**
```bash
DB_PORT=3306  # Default MySQL port
DB_PORT=3307  # Custom port
```

---

#### DB_USER

- **Required:** Yes
- **Type:** String
- **Default:** None
- **Description:** MySQL username for database authentication
- **Example:** `root`, `weather_user`
- **Validation:** Must have SELECT, UPDATE privileges on `weather_alarms` and `weather_service_tokens`

**Usage:**
```bash
DB_USER=root           # Development
DB_USER=weather_user   # Production (recommended: dedicated user)
```

---

#### DB_PASSWORD

- **Required:** No (required for production)
- **Type:** String
- **Default:** Empty string
- **Description:** MySQL password for database authentication
- **Example:** `secret123`
- **Validation:** None (but use strong password in production)
- **Security:** Store in Kubernetes Secret, not ConfigMap

**Usage:**
```bash
DB_PASSWORD=                # Development (no password)
DB_PASSWORD=secret123       # Production

# Kubernetes Secret
kubectl create secret generic scheduler-secrets \
  --from-literal=db-password='secret123'
```

---

#### DB_NAME

- **Required:** Yes
- **Type:** String
- **Default:** None
- **Description:** MySQL database name
- **Example:** `joker`, `weather_db`
- **Validation:** Database must exist with required tables

**Usage:**
```bash
DB_NAME=joker       # Default database
DB_NAME=weather_db  # Custom database
```

**Required Tables:**
- `weather_alarms`
- `weather_service_tokens`

---

#### REDIS_HOST

- **Required:** Yes
- **Type:** String
- **Default:** None
- **Description:** Redis hostname or IP address
- **Example:** `localhost`, `redis.example.com`, `10.0.1.200`
- **Validation:** Must be reachable on REDIS_PORT

**Usage:**
```bash
REDIS_HOST=localhost           # Local development
REDIS_HOST=redis.example.com   # Production
REDIS_HOST=10.0.1.200         # IP address
```

---

#### REDIS_PORT

- **Required:** No
- **Type:** Integer
- **Default:** `6379`
- **Description:** Redis port
- **Example:** `6379`, `6380`
- **Validation:** Must be 1-65535

**Usage:**
```bash
REDIS_PORT=6379  # Default Redis port
REDIS_PORT=6380  # Custom port
```

---

#### REDIS_PASSWORD

- **Required:** No
- **Type:** String
- **Default:** Empty string
- **Description:** Redis password for authentication
- **Example:** `redis_secret`
- **Validation:** None
- **Security:** Recommended for production

**Usage:**
```bash
REDIS_PASSWORD=              # Development (no auth)
REDIS_PASSWORD=redis_secret  # Production

# Kubernetes Secret
kubectl create secret generic scheduler-secrets \
  --from-literal=redis-password='redis_secret'
```

---

#### FCM_CREDENTIALS_PATH

- **Required:** Yes
- **Type:** String (file path)
- **Default:** None
- **Description:** Absolute path to Firebase service account credentials JSON file
- **Example:** `/etc/firebase/credentials.json`, `/opt/config/firebase.json`
- **Validation:** File must exist and contain valid Firebase credentials

**Usage:**
```bash
FCM_CREDENTIALS_PATH=/etc/firebase/credentials.json  # Production
FCM_CREDENTIALS_PATH=./firebase-credentials.json     # Development
```

**File Format:**
```json
{
  "type": "service_account",
  "project_id": "your-project-id",
  "private_key_id": "...",
  "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
  "client_email": "firebase-adminsdk-...@your-project-id.iam.gserviceaccount.com",
  "client_id": "...",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "..."
}
```

**Kubernetes Mount:**
```yaml
volumeMounts:
  - name: firebase-credentials
    mountPath: /etc/firebase
    readOnly: true

volumes:
  - name: firebase-credentials
    secret:
      secretName: firebase-credentials
```

---

### Optional Variables

#### SCHEDULER_INTERVAL

- **Required:** No
- **Type:** Duration (Go duration string)
- **Default:** `1m`
- **Description:** Interval between alarm processing runs
- **Example:** `30s`, `1m`, `2m`
- **Validation:** Must be parseable by `time.ParseDuration()`
- **Recommendation:** `1m` for minute-granularity alarms

**Usage:**
```bash
SCHEDULER_INTERVAL=1m   # Check every minute (recommended)
SCHEDULER_INTERVAL=30s  # Check every 30 seconds (higher load)
SCHEDULER_INTERVAL=2m   # Check every 2 minutes (lower precision)
```

**Trade-offs:**
- Lower interval: Lower latency, higher resource usage
- Higher interval: Lower resource usage, lower precision

---

#### METRICS_PORT

- **Required:** No
- **Type:** Integer
- **Default:** `9090`
- **Description:** HTTP port for metrics and health endpoints
- **Example:** `9090`, `8080`, `3000`
- **Validation:** Must be 1-65535, not in use by another process

**Usage:**
```bash
METRICS_PORT=9090  # Default Prometheus port
METRICS_PORT=8080  # Alternative port
```

**Endpoints:**
- `/metrics` - Prometheus metrics
- `/health` - Health check
- `/` - Service info

---

#### LOG_LEVEL

- **Required:** No
- **Type:** String (enum)
- **Default:** `info`
- **Description:** Logging verbosity level
- **Example:** `debug`, `info`, `warn`, `error`
- **Validation:** Must be one of: `debug`, `info`, `warn`, `error`

**Usage:**
```bash
LOG_LEVEL=info   # Production (default)
LOG_LEVEL=debug  # Development (verbose)
LOG_LEVEL=warn   # Production (minimal)
LOG_LEVEL=error  # Production (errors only)
```

**Log Levels:**
- `debug`: All logs (very verbose)
- `info`: Informational messages
- `warn`: Warning messages
- `error`: Error messages only

---

#### LOG_OUTPUT

- **Required:** No
- **Type:** String
- **Default:** `stdout`
- **Description:** Log output destination
- **Example:** `stdout`, `/var/log/weather-scheduler.log`
- **Validation:** `stdout` or writable file path

**Usage:**
```bash
LOG_OUTPUT=stdout                           # Docker/Kubernetes (default)
LOG_OUTPUT=/var/log/weather-scheduler.log  # Systemd/VM
```

**Note:** Use `stdout` for containerized deployments (Docker, Kubernetes) for centralized logging.

---

#### ENV

- **Required:** No
- **Type:** String (enum)
- **Default:** `development`
- **Description:** Environment name (affects log format)
- **Example:** `development`, `production`
- **Validation:** Any string

**Usage:**
```bash
ENV=development  # Development (human-readable logs)
ENV=production   # Production (JSON logs)
```

**Log Format:**
- `development`: Console encoder (colored, human-readable)
- `production`: JSON encoder (structured, parseable)

---

## Database Configuration

### Connection String

Constructed from environment variables:

```
DSN = "{DB_USER}:{DB_PASSWORD}@tcp({DB_HOST}:{DB_PORT})/{DB_NAME}?charset=utf8mb4&parseTime=True&loc=Local"
```

**Example:**
```
root:secret123@tcp(localhost:3306)/joker?charset=utf8mb4&parseTime=True&loc=Local
```

### Connection Pool

**Settings:**
```go
sqlDB.SetMaxIdleConns(10)              // Idle connections in pool
sqlDB.SetMaxOpenConns(100)             // Max open connections
sqlDB.SetConnMaxLifetime(time.Hour)    // Connection lifetime
```

**Tuning:**
- **MaxIdleConns:** Increase if many short queries
- **MaxOpenConns:** Increase if high concurrency
- **ConnMaxLifetime:** Decrease if connection errors

**Example:**
```go
// High load configuration
sqlDB.SetMaxIdleConns(20)
sqlDB.SetMaxOpenConns(200)
sqlDB.SetConnMaxLifetime(30 * time.Minute)
```

### Required Permissions

**Minimal Privileges:**
```sql
CREATE USER 'weather_user'@'%' IDENTIFIED BY 'secret123';

GRANT SELECT, UPDATE ON joker.weather_alarms TO 'weather_user'@'%';
GRANT SELECT ON joker.weather_service_tokens TO 'weather_user'@'%';

FLUSH PRIVILEGES;
```

**Recommended Indexes:**
```sql
CREATE INDEX idx_alarm_time_enabled ON weather_alarms(alarm_time, is_enabled);
CREATE INDEX idx_last_sent ON weather_alarms(last_sent);
CREATE INDEX idx_user_id_enabled ON weather_alarms(user_id, is_enabled);
CREATE INDEX idx_token_user_id ON weather_service_tokens(user_id);
```

---

## Redis Configuration

### Connection String

Constructed from environment variables:

```
ADDR = "{REDIS_HOST}:{REDIS_PORT}"
```

**Example:**
```
localhost:6379
redis.example.com:6379
```

### Connection Options

```go
redisClient := redis.NewClient(&redis.Options{
    Addr:     "{REDIS_HOST}:{REDIS_PORT}",
    Password: "{REDIS_PASSWORD}",
    DB:       0,                    // Default database
    PoolSize: 10,                   // Connection pool size
})
```

### Cache Configuration

**TTL:** 10 minutes (600 seconds)

```go
const CacheTTL = 10 * time.Minute
```

**Key Format:**
```
weather:{region}
```

**Example:**
```
weather:서울
weather:부산
```

### Memory Management

**Recommended Settings:**
```bash
# redis.conf
maxmemory 2gb
maxmemory-policy allkeys-lru
```

**Monitoring:**
```bash
# Check memory usage
redis-cli INFO memory

# Check keys
redis-cli DBSIZE

# Check specific key TTL
redis-cli TTL weather:서울
```

---

## Firebase Configuration

### Credentials File

Download from Firebase Console:
1. Go to Project Settings → Service Accounts
2. Click "Generate New Private Key"
3. Save as `firebase-credentials.json`
4. Set `FCM_CREDENTIALS_PATH` to file location

### Initialization

```go
opt := option.WithCredentialsFile(credentialsPath)
app, err := firebase.NewApp(ctx, nil, opt)
```

### Batch Configuration

**Maximum tokens per batch:** 500 (FCM limit)

```go
const MaxTokensPerBatch = 500
```

### Error Handling

**Retry Strategy:**
- Retry failed batches once
- Log individual token failures
- Return error only if all sends fail

---

## Scheduler Configuration

### Interval

**Default:** 1 minute

```bash
SCHEDULER_INTERVAL=1m
```

**How it works:**
- Ticker fires every `SCHEDULER_INTERVAL`
- Processes alarms for `current_time + SCHEDULER_INTERVAL`
- Example: At 09:00:00, processes alarms for 09:01:00

### Graceful Shutdown

**Timeout:** 30 seconds

```go
shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
```

**Process:**
1. Send SIGTERM signal
2. Cancel context (stops ticker)
3. Wait for in-flight operations (max 30s)
4. Close connections
5. Exit

---

## Logging Configuration

### Log Levels

**Development:**
```go
config := zap.NewDevelopmentConfig()
config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
```

**Production:**
```go
config := zap.NewProductionConfig()
config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
```

### Log Format

**Development (Console):**
```
2025-11-11T10:30:00.123+09:00  INFO  scheduler  Processing alarms  {"target_time": "2025-11-11T10:31:00+09:00"}
```

**Production (JSON):**
```json
{
  "level": "info",
  "ts": "2025-11-11T10:30:00.123+09:00",
  "caller": "scheduler/scheduler.go:150",
  "msg": "Processing alarms",
  "target_time": "2025-11-11T10:31:00+09:00"
}
```

### Log Rotation

**Systemd (journald):**
```bash
# Configure in /etc/systemd/journald.conf
SystemMaxUse=1G
MaxFileSec=1week
```

**File-based:**
```bash
# Use logrotate
# /etc/logrotate.d/weather-scheduler
/var/log/weather-scheduler.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0640 scheduler scheduler
}
```

---

## Metrics Configuration

### Prometheus Metrics

**Endpoint:** `http://localhost:{METRICS_PORT}/metrics`

**Exposed Metrics:**
- `weather_alarms_processed_total`: Counter
- `weather_alarms_failed_total`: Counter
- `weather_cache_hits_total`: Counter
- `weather_cache_misses_total`: Counter
- `weather_fcm_sent_total`: Counter
- `weather_fcm_failed_total`: Counter
- `weather_processing_duration_seconds`: Histogram

### Scrape Configuration

**Prometheus config:**
```yaml
scrape_configs:
  - job_name: 'weather-scheduler'
    static_configs:
      - targets: ['localhost:9090']
    scrape_interval: 15s
    scrape_timeout: 10s
```

**Kubernetes ServiceMonitor:**
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: weather-scheduler
spec:
  selector:
    matchLabels:
      app: weather-scheduler
  endpoints:
    - port: metrics
      interval: 15s
```

---

## Configuration Examples

### Development (.env)

```bash
# Logging
LOG_LEVEL=debug
LOG_OUTPUT=stdout
ENV=development

# Metrics
METRICS_PORT=9090

# Scheduler
SCHEDULER_INTERVAL=1m

# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=
DB_NAME=joker

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Firebase
FCM_CREDENTIALS_PATH=./firebase-credentials.json
```

### Production (Systemd)

**Environment file:** `/etc/weather-scheduler/config.env`

```bash
# Logging
LOG_LEVEL=info
LOG_OUTPUT=/var/log/weather-scheduler.log
ENV=production

# Metrics
METRICS_PORT=9090

# Scheduler
SCHEDULER_INTERVAL=1m

# Database
DB_HOST=mysql.internal.example.com
DB_PORT=3306
DB_USER=weather_user
DB_PASSWORD=secure_password_here
DB_NAME=joker

# Redis
REDIS_HOST=redis.internal.example.com
REDIS_PORT=6379
REDIS_PASSWORD=redis_password_here

# Firebase
FCM_CREDENTIALS_PATH=/etc/firebase/credentials.json
```

**Systemd service:** `/etc/systemd/system/weather-scheduler.service`

```ini
[Unit]
Description=Weather Scheduler Service
After=network.target mysql.service redis.service

[Service]
Type=simple
User=scheduler
Group=scheduler
EnvironmentFile=/etc/weather-scheduler/config.env
ExecStart=/usr/local/bin/scheduler
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

### Production (Kubernetes)

**ConfigMap:** `k8s/configmap.yaml`

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: scheduler-config
data:
  LOG_LEVEL: "info"
  LOG_OUTPUT: "stdout"
  ENV: "production"
  METRICS_PORT: "9090"
  SCHEDULER_INTERVAL: "1m"
  DB_HOST: "mysql.default.svc.cluster.local"
  DB_PORT: "3306"
  DB_USER: "weather_user"
  DB_NAME: "joker"
  REDIS_HOST: "redis.default.svc.cluster.local"
  REDIS_PORT: "6379"
```

**Secret:** `k8s/secret.yaml`

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: scheduler-secrets
type: Opaque
stringData:
  DB_PASSWORD: "secure_password_here"
  REDIS_PASSWORD: "redis_password_here"
  firebase-credentials.json: |
    {
      "type": "service_account",
      "project_id": "your-project-id",
      ...
    }
```

**Deployment:** `k8s/deployment.yaml`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: weather-scheduler
spec:
  replicas: 1
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
        image: weather-scheduler:1.0.0
        ports:
        - containerPort: 9090
          name: metrics
        envFrom:
        - configMapRef:
            name: scheduler-config
        env:
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: scheduler-secrets
              key: DB_PASSWORD
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: scheduler-secrets
              key: REDIS_PASSWORD
        - name: FCM_CREDENTIALS_PATH
          value: /etc/firebase/credentials.json
        volumeMounts:
        - name: firebase-credentials
          mountPath: /etc/firebase
          readOnly: true
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
      volumes:
      - name: firebase-credentials
        secret:
          secretName: scheduler-secrets
          items:
          - key: firebase-credentials.json
            path: credentials.json
```

---

## Validation

### Configuration Validation

**At startup, the service validates:**

1. **Required variables present:**
   - `DB_HOST`, `DB_USER`, `DB_NAME`
   - `REDIS_HOST`
   - `FCM_CREDENTIALS_PATH`

2. **FCM credentials file exists:**
   ```go
   if _, err := os.Stat(credentialsPath); os.IsNotExist(err) {
       return fmt.Errorf("FCM credentials file not found")
   }
   ```

3. **Database connection:**
   ```go
   if err := sqlDB.Ping(); err != nil {
       return fmt.Errorf("failed to ping database: %w", err)
   }
   ```

4. **Redis connection:**
   ```go
   if err := redisClient.Ping(ctx).Err(); err != nil {
       return fmt.Errorf("failed to connect to Redis: %w", err)
   }
   ```

**Service will fail fast if any validation fails.**

---

## Version History

- **v1.0.0** (2025-11-11): Initial release
