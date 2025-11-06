#!/bin/bash

# Deployment script for Joker Backend services
# Usage: ./scripts/deploy.sh [service-name] [port]

set -e

SERVICE_NAME=${1:-joker-backend}
SERVICE_PORT=${2:-6000}
DB_NAME=${3:-backend_dev}  # 모든 서비스가 동일한 DB 사용

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

# Copy project files
echo "📦 Copying project files..."
rsync -av --exclude='.git' --exclude='bin' --exclude='.claude*' \
  --exclude='node_modules' --exclude='.env' \
  "${PROJECT_ROOT}/" "${DEPLOY_DIR}/"

# Check and use existing MySQL
MYSQL_EXISTS=false
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
    MYSQL_CONTAINER_NAME="mysql"
    MYSQL_EXISTS=true
  fi
else
  echo "🆕 No MySQL found on port 3306, creating new container..."
  cd "${DEPLOY_DIR}"
  docker-compose -f docker-compose.prod.yml up -d mysql

  # Wait for MySQL to be ready
  for i in {1..10}; do
    if docker ps --filter "name=joker_mysql" --filter "status=running" | grep -q joker_mysql; then
      echo "MySQL container started"
      break
    fi
    echo "Waiting for MySQL container... ($i/10)"
    sleep 2
  done

  MYSQL_CONTAINER_NAME="joker_mysql"
  MYSQL_NETWORK="joker_network"
  MYSQL_EXISTS=false
fi

# Create .env file if not exists
if [ ! -f "${DEPLOY_DIR}/.env" ]; then
  echo "📝 Creating .env file..."
  cat > "${DEPLOY_DIR}/.env" << EOF
SERVICE_NAME=${SERVICE_NAME}
PORT=${SERVICE_PORT}
DB_HOST=${MYSQL_CONTAINER_NAME}
DB_PORT=3306
DB_USER=joker_user
DB_PASSWORD=${DB_PASSWORD:-change_me}
DB_NAME=${DB_NAME}
MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD:-change_me}
LOG_LEVEL=info
ENV=production
EOF
fi

# Stop existing API container
echo "🛑 Stopping existing API container..."
cd "${DEPLOY_DIR}"
docker stop ${SERVICE_NAME}_api 2>/dev/null || true
docker rm ${SERVICE_NAME}_api 2>/dev/null || true

# Build and start API container
echo "🔨 Building API container..."
docker-compose -f docker-compose.prod.yml build api

# Get the image name that was just built
IMAGE_NAME=$(docker-compose -f docker-compose.prod.yml images -q api | head -1)
if [ -z "$IMAGE_NAME" ]; then
  IMAGE_NAME="joker-backend-api"
fi
echo "📦 Using image: $IMAGE_NAME"

# Start API container and connect to MySQL network
if [ "$MYSQL_EXISTS" == "true" ]; then
  echo "🔗 Connecting API to MySQL network: $MYSQL_NETWORK"
  docker run -d \
    --name ${SERVICE_NAME}_api \
    --network $MYSQL_NETWORK \
    --env-file .env \
    -p ${SERVICE_PORT}:${SERVICE_PORT} \
    --restart unless-stopped \
    $IMAGE_NAME
else
  # Use docker-compose if we created our own MySQL
  echo "🚀 Starting API with docker-compose..."
  docker-compose -f docker-compose.prod.yml up -d --no-deps api
fi

# Wait for services
echo "⏳ Waiting for services to be ready..."
sleep 10

# Check MySQL health directly using docker exec
for i in {1..30}; do
  if docker exec ${MYSQL_CONTAINER_NAME} mysqladmin ping -h localhost --silent; then
    echo "✅ MySQL is healthy"
    break
  fi
  echo "Waiting for MySQL... ($i/30)"
  sleep 2
done

# Check API health
for i in {1..30}; do
  if curl -f http://localhost:${SERVICE_PORT}/health > /dev/null 2>&1; then
    echo "✅ API is healthy"
    break
  fi
  echo "Waiting for API... ($i/30)"
  sleep 2
done

# Verify deployment
echo "🔍 Verifying deployment..."
docker-compose -f docker-compose.prod.yml ps

# Test health endpoint
if curl -f http://localhost:${SERVICE_PORT}/health > /dev/null 2>&1; then
  echo "✅ Deployment successful!"
  echo "📊 Service: ${SERVICE_NAME}"
  echo "🌐 Port: ${SERVICE_PORT}"
  echo "🔗 Health: http://localhost:${SERVICE_PORT}/health"
  exit 0
else
  echo "❌ Deployment failed - Health check failed"
  docker-compose -f docker-compose.prod.yml logs api
  exit 1
fi
