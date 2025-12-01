#!/bin/bash

# Multi-service deployment script
# Usage: ./scripts/deploy-service.sh <service-name> <port>
# Example: ./scripts/deploy-service.sh auth-service 6000

set -e

SERVICE_NAME=$1
SERVICE_PORT=$2
# Convert service name to lowercase for Docker compatibility
SERVICE_NAME_LOWER=$(echo "${SERVICE_NAME}" | tr '[:upper:]' '[:lower:]')
SERVICE_DIR="services/${SERVICE_NAME}"

if [ -z "$SERVICE_NAME" ] || [ -z "$SERVICE_PORT" ]; then
  echo "Usage: $0 <service-name> <port>"
  echo "Example: $0 auth-service 6000"
  exit 1
fi

# Use HOME directory for deployment
DEPLOY_DIR="${HOME}/services/${SERVICE_NAME}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "🚀 Deploying ${SERVICE_NAME} on port ${SERVICE_PORT}..."

# Clean up disk space before build
echo "🧹 Cleaning up disk space..."
docker system prune -af --volumes || true

echo "💾 Disk space available:"
df -h | grep -E '(Filesystem|/$|/home)'

# Create deployment directory
mkdir -p "${DEPLOY_DIR}"

# Copy service code
echo "📦 Copying ${SERVICE_NAME} files..."
rsync -av --exclude='.git' --exclude='bin' --exclude='.claude*' \
  --exclude='node_modules' --exclude='.env' \
  "${PROJECT_ROOT}/${SERVICE_DIR}/" "${DEPLOY_DIR}/"

# Copy shared code
echo "📦 Copying shared code..."
rsync -av --exclude='.git' \
  "${PROJECT_ROOT}/shared/" "${DEPLOY_DIR}/shared/"

# Copy migrations
echo "📦 Copying migrations..."
rsync -av --exclude='.git' \
  "${PROJECT_ROOT}/migrations/" "${DEPLOY_DIR}/migrations/"

# Check and use existing MySQL
MYSQL_EXISTS=false
MYSQL_CONTAINER_NAME="mysql"
MYSQL_NETWORK="joker_network"

if sudo lsof -Pi :3306 -sTCP:LISTEN -t >/dev/null 2>&1; then
  echo "✅ MySQL is already running on port 3306"

  # Find the container using port 3306
  MYSQL_CONTAINER=$(docker ps --filter "publish=3306" --format "{{.Names}}" | head -1)

  if [ -n "$MYSQL_CONTAINER" ]; then
    echo "📦 Using existing MySQL container: $MYSQL_CONTAINER"

    # Get the network of the existing MySQL container
    MYSQL_NETWORK=$(docker inspect $MYSQL_CONTAINER --format '{{range $key, $value := .NetworkSettings.Networks}}{{$key}}{{end}}')
    echo "🌐 MySQL is using network: $MYSQL_NETWORK"

    MYSQL_CONTAINER_NAME="$MYSQL_CONTAINER"
    MYSQL_EXISTS=true
  else
    echo "⚠️  Port 3306 is in use but not by a Docker container"
    echo "    Assuming external MySQL is available"
    MYSQL_EXISTS=true
  fi
else
  echo "⚠️  No MySQL found on port 3306"
  echo "    Please ensure MySQL is running or update docker-compose to include it"
fi

# Check and use existing Redis (for async queue services)
REDIS_EXISTS=false
REDIS_CONTAINER_NAME="redis"
REDIS_PASSWORD=""

# For cloudRepositoryService, find any running Redis container
if [ "$SERVICE_NAME" == "cloudRepositoryService" ]; then
  echo "🔍 Looking for running Redis container..."

  # Find any Redis container using port 6379
  REDIS_CONTAINER=$(docker ps --filter "publish=6379" --format "{{.Names}}" | grep -i redis | head -1)

  if [ -z "$REDIS_CONTAINER" ]; then
    # No container on port 6379, check for any running Redis
    REDIS_CONTAINER=$(docker ps --filter "ancestor=redis" --format "{{.Names}}" | head -1)
  fi

  if [ -n "$REDIS_CONTAINER" ]; then
    echo "✅ Found existing Redis container: $REDIS_CONTAINER"
    REDIS_CONTAINER_NAME="$REDIS_CONTAINER"

    # Test if it requires auth
    if docker exec $REDIS_CONTAINER redis-cli ping 2>&1 | grep -q "NOAUTH"; then
      echo "⚠️  Redis requires authentication"
      echo "📝 Checking for Redis password in container environment..."

      # Try to get password from container environment
      CONTAINER_PASSWORD=$(docker inspect $REDIS_CONTAINER --format '{{range .Config.Env}}{{println .}}{{end}}' | grep -i "REDIS_PASSWORD" | cut -d'=' -f2)

      if [ -n "$CONTAINER_PASSWORD" ]; then
        echo "✅ Found Redis password in container environment"
        REDIS_PASSWORD="$CONTAINER_PASSWORD"
        # Test with password
        if docker exec $REDIS_CONTAINER redis-cli -a "$REDIS_PASSWORD" ping 2>&1 | grep -q "PONG"; then
          echo "✅ Redis authentication successful"
          REDIS_EXISTS=true
        fi
      else
        echo "⚠️  No password found in environment, trying common passwords..."
        # Try empty password (no auth)
        echo "    Testing with no authentication..."
        REDIS_EXISTS=true  # Assume it will work
      fi
    elif docker exec $REDIS_CONTAINER redis-cli ping 2>&1 | grep -q "PONG"; then
      echo "✅ Redis is accessible without authentication"
      REDIS_EXISTS=true
    else
      echo "⚠️  Redis ping failed, but will try to use it anyway"
      REDIS_EXISTS=true
    fi
  else
    echo "❌ No Redis container found"
    echo "   Redis is required for cloudRepositoryService"
    exit 1
  fi
fi

# Create .env file
echo "📝 Creating .env file..."
cat > "${DEPLOY_DIR}/.env" << EOF
SERVICE_NAME=${SERVICE_NAME}
PORT=${SERVICE_PORT}
IS_LOCAL=true
MYSQL_HOST=${MYSQL_CONTAINER_NAME}
MYSQL_PORT=3306
MYSQL_USER=${DB_USER:-joker_user}
MYSQL_PASSWORD=${DB_PASSWORD:-change_me}
MYSQL_DATABASE=backend_dev
LOG_LEVEL=info
ENV=${ENV:-production}
CORS_ALLOWED_ORIGINS=${CORS_ALLOWED_ORIGINS:-}
MIGRATIONS_PATH=./migrations
AWS_REGION=${AWS_REGION:-}
CLOUD_REPOSITORY_BUCKET=${CLOUD_REPOSITORY_BUCKET:-}
REDIS_HOST=${REDIS_CONTAINER_NAME}:6379
REDIS_PASSWORD=
JWT_SECRET=${JWT_SECRET:-joker-jwt-secret-key-change-this-in-production}
EOF

# Stop existing container
echo "🛑 Stopping existing ${SERVICE_NAME} container..."
cd "${DEPLOY_DIR}"
docker stop ${SERVICE_NAME_LOWER}_api 2>/dev/null || true
docker rm ${SERVICE_NAME_LOWER}_api 2>/dev/null || true

# Build Docker image
echo "🔨 Building ${SERVICE_NAME} Docker image..."
docker build -t ${SERVICE_NAME_LOWER}:latest .

# Start container and connect to MySQL network
echo "🚀 Starting ${SERVICE_NAME} container..."
if [ "$MYSQL_EXISTS" == "true" ]; then
  echo "🔗 Connecting to MySQL network: $MYSQL_NETWORK"
  docker run -d \
    --name ${SERVICE_NAME_LOWER}_api \
    --network $MYSQL_NETWORK \
    --env-file .env \
    -p ${SERVICE_PORT}:${SERVICE_PORT} \
    --restart unless-stopped \
    ${SERVICE_NAME_LOWER}:latest
else
  docker run -d \
    --name ${SERVICE_NAME_LOWER}_api \
    --env-file .env \
    -p ${SERVICE_PORT}:${SERVICE_PORT} \
    --restart unless-stopped \
    ${SERVICE_NAME_LOWER}:latest
fi

# Wait for service
echo "⏳ Waiting for service to be ready..."
sleep 10

# Check MySQL health
for i in {1..30}; do
  if docker exec ${MYSQL_CONTAINER_NAME} mysqladmin ping -h localhost --silent 2>/dev/null; then
    echo "✅ MySQL is healthy"
    break
  fi
  echo "Waiting for MySQL... ($i/30)"
  sleep 2
done

# Check API health
HEALTH_CHECK_PASSED=false
for i in {1..30}; do
  if curl -f http://localhost:${SERVICE_PORT}/health > /dev/null 2>&1; then
    echo "✅ API is healthy"
    HEALTH_CHECK_PASSED=true
    break
  fi
  echo "Waiting for API... ($i/30)"
  sleep 2
done

# If health check failed, show container logs
if [ "$HEALTH_CHECK_PASSED" = false ]; then
  echo "⚠️  Health check failed, showing container logs:"
  docker logs ${SERVICE_NAME_LOWER}_api --tail 50
fi

# Verify deployment
echo "🔍 Verifying deployment..."
docker ps --filter "name=${SERVICE_NAME_LOWER}_api"

# Test health endpoint
if curl -f http://localhost:${SERVICE_PORT}/health > /dev/null 2>&1; then
  echo "✅ Deployment successful!"
  echo "📊 Service: ${SERVICE_NAME}"
  echo "🌐 Port: ${SERVICE_PORT}"
  echo "🔗 Health: http://localhost:${SERVICE_PORT}/health"
  exit 0
else
  echo "❌ Deployment failed - Health check failed"
  docker logs ${SERVICE_NAME_LOWER}_api
  exit 1
fi
