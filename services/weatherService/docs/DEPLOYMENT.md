# Deployment Guide

Deployment instructions for the Weather Data Collector Service across different environments.

**Version:** 1.0.0
**Last Updated:** 2025-11-11

## Table of Contents

- [Local Development](#local-development)
- [Docker](#docker)
- [Docker Compose](#docker-compose)
- [Systemd](#systemd)
- [Kubernetes](#kubernetes)

## Local Development

### Prerequisites

- Go 1.21+
- MySQL 8.0+
- Redis 7.0+
- Firebase credentials

### Steps

1. **Build Binary**
   ```bash
   cd services/weatherService
   make build-scheduler
   ```

2. **Configure Environment**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Run Scheduler**
   ```bash
   ./scheduler
   ```

4. **Verify**
   ```bash
   curl http://localhost:9090/health
   ```

## Docker

### Build Image

```bash
# From services/weatherService directory
docker build -t weather-scheduler:1.0.0 .
```

### Run Container

```bash
docker run -d \
  --name weather-scheduler \
  --env-file .env \
  -p 9090:9090 \
  -v /path/to/firebase-credentials.json:/etc/firebase/credentials.json:ro \
  weather-scheduler:1.0.0
```

### With Environment Variables

```bash
docker run -d \
  --name weather-scheduler \
  -e DB_HOST=mysql.example.com \
  -e DB_PORT=3306 \
  -e DB_USER=weather_user \
  -e DB_PASSWORD=secret123 \
  -e DB_NAME=joker \
  -e REDIS_HOST=redis.example.com \
  -e REDIS_PORT=6379 \
  -e REDIS_PASSWORD=redis_secret \
  -e FCM_CREDENTIALS_PATH=/etc/firebase/credentials.json \
  -e SCHEDULER_INTERVAL=1m \
  -e METRICS_PORT=9090 \
  -e LOG_LEVEL=info \
  -e ENV=production \
  -p 9090:9090 \
  -v /path/to/firebase-credentials.json:/etc/firebase/credentials.json:ro \
  weather-scheduler:1.0.0
```

### Verify Deployment

```bash
# Check container status
docker ps | grep weather-scheduler

# Check logs
docker logs -f weather-scheduler

# Check health
curl http://localhost:9090/health

# Check metrics
curl http://localhost:9090/metrics
```

### Stop and Remove

```bash
docker stop weather-scheduler
docker rm weather-scheduler
```

## Docker Compose

### Create docker-compose.yml

```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: joker
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
      - ../../shared/db/mysql/table.sql:/docker-entrypoint-initdb.d/init.sql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7.0-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  scheduler:
    build: .
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      DB_HOST: mysql
      DB_PORT: 3306
      DB_USER: root
      DB_PASSWORD: password
      DB_NAME: joker
      REDIS_HOST: redis
      REDIS_PORT: 6379
      REDIS_PASSWORD: ""
      FCM_CREDENTIALS_PATH: /etc/firebase/credentials.json
      SCHEDULER_INTERVAL: 1m
      METRICS_PORT: 9090
      LOG_LEVEL: info
      ENV: production
    ports:
      - "9090:9090"
    volumes:
      - ./firebase-credentials.json:/etc/firebase/credentials.json:ro
    restart: unless-stopped

volumes:
  mysql_data:
  redis_data:
```

### Deploy

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f scheduler

# Check status
docker-compose ps

# Stop services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

### Verify Deployment

```bash
# Health check
curl http://localhost:9090/health

# Metrics
curl http://localhost:9090/metrics

# Logs
docker-compose logs scheduler | tail -50
```

## Systemd

For production deployment on Linux VMs.

### Prerequisites

- Linux VM with systemd
- Go binary installed at `/usr/local/bin/scheduler`
- Firebase credentials at `/etc/firebase/credentials.json`
- Configuration at `/etc/weather-scheduler/config.env`

### Setup Steps

1. **Create User**
   ```bash
   sudo useradd -r -s /bin/false scheduler
   ```

2. **Install Binary**
   ```bash
   sudo cp scheduler /usr/local/bin/
   sudo chmod +x /usr/local/bin/scheduler
   sudo chown root:root /usr/local/bin/scheduler
   ```

3. **Create Configuration Directory**
   ```bash
   sudo mkdir -p /etc/weather-scheduler
   sudo mkdir -p /etc/firebase
   sudo mkdir -p /var/log/weather-scheduler
   sudo chown scheduler:scheduler /var/log/weather-scheduler
   ```

4. **Create Configuration File**
   ```bash
   sudo vim /etc/weather-scheduler/config.env
   ```

   ```bash
   # /etc/weather-scheduler/config.env
   LOG_LEVEL=info
   LOG_OUTPUT=/var/log/weather-scheduler/scheduler.log
   ENV=production
   METRICS_PORT=9090
   SCHEDULER_INTERVAL=1m

   DB_HOST=mysql.internal.example.com
   DB_PORT=3306
   DB_USER=weather_user
   DB_PASSWORD=secure_password_here
   DB_NAME=joker

   REDIS_HOST=redis.internal.example.com
   REDIS_PORT=6379
   REDIS_PASSWORD=redis_password_here

   FCM_CREDENTIALS_PATH=/etc/firebase/credentials.json
   ```

5. **Set Permissions**
   ```bash
   sudo chown root:scheduler /etc/weather-scheduler/config.env
   sudo chmod 640 /etc/weather-scheduler/config.env
   ```

6. **Copy Firebase Credentials**
   ```bash
   sudo cp firebase-credentials.json /etc/firebase/credentials.json
   sudo chown root:scheduler /etc/firebase/credentials.json
   sudo chmod 640 /etc/firebase/credentials.json
   ```

7. **Create Systemd Service**
   ```bash
   sudo vim /etc/systemd/system/weather-scheduler.service
   ```

   ```ini
   [Unit]
   Description=Weather Scheduler Service
   After=network.target mysql.service redis.service
   Wants=mysql.service redis.service

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

   # Security hardening
   NoNewPrivileges=true
   PrivateTmp=true
   ProtectSystem=strict
   ProtectHome=true
   ReadWritePaths=/var/log/weather-scheduler

   [Install]
   WantedBy=multi-user.target
   ```

8. **Enable and Start Service**
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable weather-scheduler
   sudo systemctl start weather-scheduler
   ```

9. **Verify Deployment**
   ```bash
   # Check status
   sudo systemctl status weather-scheduler

   # Check logs
   sudo journalctl -u weather-scheduler -f

   # Check health
   curl http://localhost:9090/health

   # Check metrics
   curl http://localhost:9090/metrics
   ```

### Log Rotation

Create logrotate configuration:

```bash
sudo vim /etc/logrotate.d/weather-scheduler
```

```bash
/var/log/weather-scheduler/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0640 scheduler scheduler
    sharedscripts
    postrotate
        systemctl reload weather-scheduler > /dev/null 2>&1 || true
    endscript
}
```

### Firewall Configuration

```bash
# Allow metrics port (internal only)
sudo firewall-cmd --permanent --add-port=9090/tcp --zone=internal
sudo firewall-cmd --reload
```

## Kubernetes

For production deployment on Kubernetes clusters.

### Prerequisites

- Kubernetes cluster (1.21+)
- kubectl configured
- Docker image pushed to registry

### Configuration Files

**1. Namespace**
```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: weather-service
```

**2. ConfigMap**
```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: scheduler-config
  namespace: weather-service
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

**3. Secret**
```yaml
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: scheduler-secrets
  namespace: weather-service
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

**4. Deployment**
```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: weather-scheduler
  namespace: weather-service
spec:
  replicas: 1  # Single instance to avoid duplicate sends
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
        image: your-registry.com/weather-scheduler:1.0.0
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
        livenessProbe:
          httpGet:
            path: /health
            port: 9090
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 9090
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: firebase-credentials
        secret:
          secretName: scheduler-secrets
          items:
          - key: firebase-credentials.json
            path: credentials.json
```

**5. Service**
```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: scheduler-metrics
  namespace: weather-service
spec:
  selector:
    app: weather-scheduler
  ports:
  - name: metrics
    port: 9090
    targetPort: 9090
  type: ClusterIP
```

**6. ServiceMonitor (for Prometheus)**
```yaml
# k8s/servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: weather-scheduler
  namespace: weather-service
spec:
  selector:
    matchLabels:
      app: weather-scheduler
  endpoints:
  - port: metrics
    interval: 15s
    path: /metrics
```

### Deployment Steps

1. **Create Namespace**
   ```bash
   kubectl apply -f k8s/namespace.yaml
   ```

2. **Create ConfigMap**
   ```bash
   kubectl apply -f k8s/configmap.yaml
   ```

3. **Create Secret**
   ```bash
   # Option 1: From file
   kubectl create secret generic scheduler-secrets \
     -n weather-service \
     --from-literal=DB_PASSWORD='secure_password' \
     --from-literal=REDIS_PASSWORD='redis_password' \
     --from-file=firebase-credentials.json=./firebase-credentials.json

   # Option 2: From YAML
   kubectl apply -f k8s/secret.yaml
   ```

4. **Deploy Application**
   ```bash
   kubectl apply -f k8s/deployment.yaml
   kubectl apply -f k8s/service.yaml
   ```

5. **Deploy Monitoring (optional)**
   ```bash
   kubectl apply -f k8s/servicemonitor.yaml
   ```

6. **Verify Deployment**
   ```bash
   # Check pods
   kubectl get pods -n weather-service

   # Check logs
   kubectl logs -f -l app=weather-scheduler -n weather-service

   # Check health
   kubectl port-forward -n weather-service svc/scheduler-metrics 9090:9090
   curl http://localhost:9090/health
   ```

### Update Deployment

```bash
# Update image
kubectl set image deployment/weather-scheduler \
  scheduler=your-registry.com/weather-scheduler:1.1.0 \
  -n weather-service

# Check rollout status
kubectl rollout status deployment/weather-scheduler -n weather-service

# Rollback if needed
kubectl rollout undo deployment/weather-scheduler -n weather-service
```

### Scaling (Not Recommended)

```bash
# Scale to multiple replicas (requires distributed locking)
kubectl scale deployment/weather-scheduler --replicas=3 -n weather-service
```

**Note:** Multiple replicas will cause duplicate notifications unless distributed locking is implemented.

---

## Version History

- **v1.0.0** (2025-11-11): Initial release
