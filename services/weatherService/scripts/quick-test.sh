#!/bin/bash
set -e

echo "================================================"
echo "Weather Data Collector - Quick Test Script"
echo "================================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}Error: Docker is not running${NC}"
    exit 1
fi

# Function to cleanup on exit
cleanup() {
    echo ""
    echo -e "${YELLOW}Cleaning up...${NC}"
    cd /Users/luxrobo/project/joker_backend/services/weatherService
    docker-compose -f docker-compose.test.yml down -v > /dev/null 2>&1 || true
}

trap cleanup EXIT

# Navigate to service directory
cd /Users/luxrobo/project/joker_backend/services/weatherService

echo "Step 1: Starting test containers (MySQL + Redis)..."
docker-compose -f docker-compose.test.yml up -d

echo "Waiting for services to be ready..."
sleep 10

# Check if containers are running
if ! docker ps | grep -q "joker_mysql_test"; then
    echo -e "${RED}Error: MySQL container failed to start${NC}"
    exit 1
fi

if ! docker ps | grep -q "joker_redis_test"; then
    echo -e "${RED}Error: Redis container failed to start${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Containers started successfully${NC}"
echo ""

echo "Step 2: Running integration tests..."
echo ""

# Run integration tests with verbose output
if go test -v -timeout 5m ./features/weather/integration/...; then
    echo ""
    echo -e "${GREEN}================================================${NC}"
    echo -e "${GREEN}All tests passed!${NC}"
    echo -e "${GREEN}================================================${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}================================================${NC}"
    echo -e "${RED}Tests failed!${NC}"
    echo -e "${RED}================================================${NC}"
    exit 1
fi
