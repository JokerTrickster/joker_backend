#!/bin/bash

# Deployment Testing Script
# Tests the deployed cloudRepositoryService endpoints

set -e

# Configuration
CLOUD_REPO_PORT=${1:-18080}
AUTH_PORT=${2:-18081}
BASE_URL="http://localhost:${CLOUD_REPO_PORT}"
AUTH_URL="http://localhost:${AUTH_PORT}"

echo "🧪 Testing Deployment..."
echo "📍 Cloud Repository Service: ${BASE_URL}"
echo "📍 Auth Service: ${AUTH_URL}"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function to test endpoint
test_endpoint() {
  local name=$1
  local url=$2
  local expected_status=${3:-200}

  echo -n "Testing ${name}... "

  response=$(curl -s -w "\n%{http_code}" "${url}" 2>/dev/null || echo "000")
  http_code=$(echo "$response" | tail -n1)
  body=$(echo "$response" | sed '$d')

  if [ "$http_code" = "$expected_status" ]; then
    echo -e "${GREEN}✓ PASS${NC} (HTTP ${http_code})"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    return 0
  else
    echo -e "${RED}✗ FAIL${NC} (HTTP ${http_code}, expected ${expected_status})"
    echo "  Response: ${body}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    return 1
  fi
}

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1. Health Check Tests"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

test_endpoint "Cloud Repository Health" "${BASE_URL}/health" 200
test_endpoint "Auth Service Health" "${AUTH_URL}/health" 200

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2. API Endpoint Tests"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Test Swagger documentation
test_endpoint "Swagger Documentation" "${BASE_URL}/swagger/index.html" 200

# Test API endpoints (should return 401 without auth)
test_endpoint "Files List (No Auth)" "${BASE_URL}/api/v1/files" 401
test_endpoint "Upload URL (No Auth)" "${BASE_URL}/api/v1/files/upload-url" 401

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3. Container Status"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo "Cloud Repository Service:"
docker ps --filter "name=cloudrepositoryservice" --format "  Status: {{.Status}}\n  Ports: {{.Ports}}"

echo ""
echo "Redis:"
docker ps --filter "publish=6379" --format "  Status: {{.Status}}\n  Ports: {{.Ports}}"

echo ""
echo "MySQL:"
docker ps --filter "publish=3306" --format "  Status: {{.Status}}\n  Ports: {{.Ports}}"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "4. Service Logs (Last 20 lines)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

CONTAINER_NAME=$(docker ps --filter "name=cloudrepositoryservice" --format "{{.Names}}" | head -1)
if [ -n "$CONTAINER_NAME" ]; then
  echo "Recent logs from ${CONTAINER_NAME}:"
  docker logs "$CONTAINER_NAME" --tail 20 2>&1 | grep -E "(level|error|fatal|Server starting|WebSocket|Redis)" || echo "  No relevant logs found"
else
  echo -e "${RED}Container not found${NC}"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "5. Redis Connection Test"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

REDIS_CONTAINER=$(docker ps --filter "publish=6379" --format "{{.Names}}" | head -1)
if [ -n "$REDIS_CONTAINER" ]; then
  echo -n "Testing Redis PING... "
  if docker exec "$REDIS_CONTAINER" redis-cli ping 2>&1 | grep -q "PONG"; then
    echo -e "${GREEN}✓ PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
  else
    echo -e "${RED}✗ FAIL${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
  fi
else
  echo -e "${RED}Redis container not found${NC}"
  TESTS_FAILED=$((TESTS_FAILED + 1))
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Test Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "Tests Passed: ${GREEN}${TESTS_PASSED}${NC}"
echo -e "Tests Failed: ${RED}${TESTS_FAILED}${NC}"
echo "Total Tests: $((TESTS_PASSED + TESTS_FAILED))"

if [ $TESTS_FAILED -eq 0 ]; then
  echo ""
  echo -e "${GREEN}✅ All tests passed!${NC}"
  exit 0
else
  echo ""
  echo -e "${RED}❌ Some tests failed${NC}"
  exit 1
fi
