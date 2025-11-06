#!/bin/bash

# Deployment script for Joker Backend services
# Usage: ./scripts/deploy.sh [service-name] [port]

set -e

SERVICE_NAME=${1:-joker-backend}
SERVICE_PORT=${2:-6000}
DB_PORT=$((SERVICE_PORT + 3309 - 6000))

DEPLOY_DIR="/home/runner/services/${SERVICE_NAME}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "🚀 Deploying ${SERVICE_NAME} on port ${SERVICE_PORT}..."

# Create deployment directory
mkdir -p "${DEPLOY_DIR}"

# Copy project files
echo "📦 Copying project files..."
rsync -av --exclude='.git' --exclude='bin' --exclude='.claude*' \
  --exclude='node_modules' --exclude='.env' \
  "${PROJECT_ROOT}/" "${DEPLOY_DIR}/"

# Create .env file if not exists
if [ ! -f "${DEPLOY_DIR}/.env" ]; then
  echo "📝 Creating .env file..."
  cat > "${DEPLOY_DIR}/.env" << EOF
SERVICE_NAME=${SERVICE_NAME}
PORT=${SERVICE_PORT}
DB_PORT=${DB_PORT}
DB_HOST=mysql
DB_USER=joker_user
DB_PASSWORD=${DB_PASSWORD:-change_me}
DB_NAME=${SERVICE_NAME//-/_}
MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD:-change_me}
LOG_LEVEL=info
ENV=production
EOF
fi

# Stop existing containers
echo "🛑 Stopping existing containers..."
cd "${DEPLOY_DIR}"
docker compose -f docker-compose.prod.yml down || true

# Build and start containers
echo "🔨 Building and starting containers..."
docker compose -f docker-compose.prod.yml up -d --build

# Wait for services
echo "⏳ Waiting for services to be ready..."
sleep 10

# Check MySQL health
for i in {1..30}; do
  if docker compose -f docker-compose.prod.yml ps mysql | grep -q "healthy"; then
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
docker compose -f docker-compose.prod.yml ps

# Test health endpoint
if curl -f http://localhost:${SERVICE_PORT}/health > /dev/null 2>&1; then
  echo "✅ Deployment successful!"
  echo "📊 Service: ${SERVICE_NAME}"
  echo "🌐 Port: ${SERVICE_PORT}"
  echo "🔗 Health: http://localhost:${SERVICE_PORT}/health"
  exit 0
else
  echo "❌ Deployment failed - Health check failed"
  docker compose -f docker-compose.prod.yml logs api
  exit 1
fi
