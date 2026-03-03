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

# Run unit tests across all service modules
test: test-unit

test-unit:
	@echo "Running unit tests..."
	@failed=0; \
	for dir in services/authService services/cloudRepositoryService services/lottoDefenseService services/morandoranService services/tdService shared; do \
		echo "=== Testing $$dir ==="; \
		(cd $$dir && go test -v -race -short ./...) || failed=1; \
	done; \
	if [ $$failed -eq 1 ]; then exit 1; fi

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
	@rm -f coverage.out
	@for dir in services/authService services/cloudRepositoryService services/lottoDefenseService services/morandoranService services/tdService shared; do \
		echo "=== Coverage: $$dir ==="; \
		(cd $$dir && go test -race -coverprofile=coverage.tmp -covermode=atomic ./... 2>/dev/null) && \
		if [ -f $$dir/coverage.tmp ]; then tail -n +2 $$dir/coverage.tmp >> coverage.out; rm $$dir/coverage.tmp; fi; \
	done
	@if [ -f coverage.out ]; then \
		echo "mode: atomic" | cat - coverage.out > coverage.tmp && mv coverage.tmp coverage.out; \
		go tool cover -html=coverage.out -o coverage.html; \
		echo "Coverage report generated: coverage.html"; \
		go tool cover -func=coverage.out; \
	fi

# Run benchmark tests
test-bench:
	@echo "Running benchmark tests..."
	@for dir in services/authService services/cloudRepositoryService services/lottoDefenseService services/morandoranService services/tdService shared; do \
		echo "=== Bench: $$dir ==="; \
		(cd $$dir && go test -run='^$$' -bench=. -benchmem ./...) || true; \
	done

# Run all test suites
test-all: test-unit test-integration test-e2e

# Run tests in watch mode (requires entr)
test-watch:
	@command -v entr >/dev/null 2>&1 || { echo "entr is required but not installed. Install with: brew install entr"; exit 1; }
	@echo "Running tests in watch mode..."
	@find . -name "*.go" | entr -c make test-unit

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
	@cd services/authService && go test -v -race ./...

test-cloud:
	@echo "Testing cloud repository service..."
	@cd services/cloudRepositoryService && go test -v -race ./...

test-tower:
	@echo "Testing TD service..."
	@cd services/tdService && go test -v -race ./...

test-lotto:
	@echo "Testing lotto defense service..."
	@cd services/lottoDefenseService && go test -v -race ./...

test-morandoran:
	@echo "Testing morandoran service..."
	@cd services/morandoranService && go test -v -race ./...

test-shared:
	@echo "Testing shared module..."
	@cd shared && go test -v -race ./...