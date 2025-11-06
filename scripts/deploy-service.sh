#!/bin/bash

# Multi-service deployment script
# Usage: ./scripts/deploy-service.sh <service-name> <port>
# Example: ./scripts/deploy-service.sh auth-service 6000

set -e

SERVICE_NAME=$1
SERVICE_PORT=$2
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

# Create .env file
echo "📝 Creating .env file..."
cat > "${DEPLOY_DIR}/.env" << EOF
SERVICE_NAME=${SERVICE_NAME}
PORT=${SERVICE_PORT}
DB_HOST=${MYSQL_CONTAINER_NAME}
DB_PORT=3306
DB_USER=${DB_USER:-joker_user}
DB_PASSWORD=${DB_PASSWORD:-change_me}
DB_NAME=backend_dev
LOG_LEVEL=info
ENV=${ENV:-production}
CORS_ALLOWED_ORIGINS=${CORS_ALLOWED_ORIGINS:-}
MIGRATIONS_PATH=./migrations
EOF

# Stop existing container
echo "🛑 Stopping existing ${SERVICE_NAME} container..."
cd "${DEPLOY_DIR}"
docker stop ${SERVICE_NAME}_api 2>/dev/null || true
docker rm ${SERVICE_NAME}_api 2>/dev/null || true

# Build Docker image
echo "🔨 Building ${SERVICE_NAME} Docker image..."
docker build -t ${SERVICE_NAME}:latest .

# Start container and connect to MySQL network
echo "🚀 Starting ${SERVICE_NAME} container..."
if [ "$MYSQL_EXISTS" == "true" ]; then
  echo "🔗 Connecting to MySQL network: $MYSQL_NETWORK"
  docker run -d \
    --name ${SERVICE_NAME}_api \
    --network $MYSQL_NETWORK \
    --env-file .env \
    -p ${SERVICE_PORT}:${SERVICE_PORT} \
    --restart unless-stopped \
    ${SERVICE_NAME}:latest
else
  docker run -d \
    --name ${SERVICE_NAME}_api \
    --env-file .env \
    -p ${SERVICE_PORT}:${SERVICE_PORT} \
    --restart unless-stopped \
    ${SERVICE_NAME}:latest
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
docker ps --filter "name=${SERVICE_NAME}_api"

# Test health endpoint
if curl -f http://localhost:${SERVICE_PORT}/health > /dev/null 2>&1; then
  echo "✅ Deployment successful!"
  echo "📊 Service: ${SERVICE_NAME}"
  echo "🌐 Port: ${SERVICE_PORT}"
  echo "🔗 Health: http://localhost:${SERVICE_PORT}/health"
  exit 0
else
  echo "❌ Deployment failed - Health check failed"
  docker logs ${SERVICE_NAME}_api
  exit 1
fi
