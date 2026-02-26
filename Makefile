.PHONY: help test test-unit test-integration test-e2e test-coverage test-bench test-all clean docker-up docker-down

# Default target
help:
	@echo "Available targets:"
	@echo "  test              - Run all unit tests"
	@echo "  test-unit         - Run unit tests only"
	@echo "  test-integration  - Run integration tests"
	@echo "  test-e2e          - Run end-to-end tests"
	@echo "  test-coverage     - Run tests with coverage report"
	@echo "  test-bench        - Run benchmark tests"
	@echo "  test-all          - Run all test suites"
	@echo "  test-watch        - Run tests in watch mode"
	@echo "  docker-up         - Start test infrastructure"
	@echo "  docker-down       - Stop test infrastructure"
	@echo "  clean             - Clean test artifacts"

# Test database configuration
TEST_DB_HOST ?= localhost
TEST_DB_PORT ?= 3307
TEST_DB_USER ?= root
TEST_DB_PASSWORD ?= rootpassword

# Export test environment variables
export TEST_DB_HOST
export TEST_DB_PORT
export TEST_DB_USER
export TEST_DB_PASSWORD

# Run unit tests
test: test-unit

test-unit:
	@echo "Running unit tests..."
	@go test -v -race -short ./services/... ./shared/...

# Run integration tests
test-integration: docker-up
	@echo "Running integration tests..."
	@go test -v -race -tags=integration ./tests/integration/...

# Run E2E tests
test-e2e: docker-up
	@echo "Starting services for E2E tests..."
	@./scripts/start-services.sh || true
	@echo "Running E2E tests..."
	@go test -v -tags=e2e ./tests/e2e/...
	@./scripts/stop-services.sh || true

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./services/... ./shared/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out

# Run benchmark tests
test-bench:
	@echo "Running benchmark tests..."
	@go test -run=^$$ -bench=. -benchmem ./services/... ./shared/...

# Run all test suites
test-all: test-unit test-integration test-e2e

# Run tests in watch mode (requires entr)
test-watch:
	@command -v entr >/dev/null 2>&1 || { echo "entr is required but not installed. Install with: brew install entr"; exit 1; }
	@echo "Running tests in watch mode..."
	@find . -name "*.go" | entr -c go test -v -race -short ./services/... ./shared/...

# Docker commands for test infrastructure
docker-up:
	@echo "Starting test infrastructure..."
	@docker-compose -f docker-compose.test.yml up -d
	@echo "Waiting for MySQL to be ready..."
	@for i in {1..30}; do \
		docker exec joker_mysql_test mysqladmin -uroot -prootpassword ping >/dev/null 2>&1 && break || sleep 1; \
	done
	@echo "Test infrastructure ready!"

docker-down:
	@echo "Stopping test infrastructure..."
	@docker-compose -f docker-compose.test.yml down -v

# Clean test artifacts
clean:
	@echo "Cleaning test artifacts..."
	@rm -f coverage.out coverage.html
	@rm -rf tests/tmp
	@rm -rf tests/logs
	@go clean -testcache

# Service-specific test targets
test-auth:
	@echo "Testing auth service..."
	@go test -v -race ./services/authService/...

test-cloud:
	@echo "Testing cloud repository service..."
	@go test -v -race ./services/cloudRepositoryService/...

test-tower:
	@echo "Testing tower defense service..."
	@go test -v -race ./services/towerDefenseService/...

test-lotto:
	@echo "Testing lotto defense service..."
	@go test -v -race ./services/lottoDefenseService/...

# Verbose test output for debugging
test-verbose:
	@echo "Running tests with verbose output..."
	@go test -v -race -count=1 ./services/... ./shared/... 2>&1 | tee test.log

# Quick smoke test
test-smoke:
	@echo "Running smoke tests..."
	@go test -v -race -run="TestSmoke" ./services/... ./shared/... || true