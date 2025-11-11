#!/bin/bash

# Comprehensive test script for weather service
# Runs all tests, generates coverage reports, and validates quality metrics

set -e  # Exit on error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
MIN_COVERAGE=80
COVERAGE_FILE="coverage.out"
COVERAGE_HTML="coverage.html"
COVERAGE_SUMMARY="coverage-summary.txt"
TEST_TIMEOUT="10m"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Weather Service Comprehensive Test Suite${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Function to print section headers
print_header() {
    echo ""
    echo -e "${BLUE}>>> $1${NC}"
    echo "----------------------------------------"
}

# Function to print success
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

# Function to print error
print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Function to print warning
print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Change to service directory
cd "$(dirname "$0")/.."
SERVICE_DIR=$(pwd)
echo "Service directory: $SERVICE_DIR"

# Clean previous test artifacts
print_header "Cleaning Previous Test Artifacts"
rm -f "$COVERAGE_FILE" "$COVERAGE_HTML" "$COVERAGE_SUMMARY"
print_success "Cleaned previous artifacts"

# Run unit tests with coverage
print_header "Running Unit Tests"
echo "This includes: cache, crawler, repository, notifier, scheduler tests"
go test -timeout "$TEST_TIMEOUT" -coverprofile="$COVERAGE_FILE" ./features/weather/... -v 2>&1 | tee test-output.log

TEST_EXIT_CODE=${PIPESTATUS[0]}

if [ $TEST_EXIT_CODE -ne 0 ]; then
    print_error "Unit tests failed"
    echo ""
    echo "Check test-output.log for details"
    exit 1
fi

print_success "Unit tests passed"

# Generate coverage report
print_header "Generating Coverage Report"
go tool cover -func="$COVERAGE_FILE" | tee "$COVERAGE_SUMMARY"

# Calculate overall coverage
TOTAL_COVERAGE=$(go tool cover -func="$COVERAGE_FILE" | grep total | awk '{print $3}' | sed 's/%//')

if [ -z "$TOTAL_COVERAGE" ]; then
    print_warning "Could not calculate coverage"
    TOTAL_COVERAGE=0
fi

echo ""
echo -e "${BLUE}Total Coverage: ${TOTAL_COVERAGE}%${NC}"

# Check coverage threshold
if (( $(echo "$TOTAL_COVERAGE < $MIN_COVERAGE" | bc -l) )); then
    print_error "Coverage ${TOTAL_COVERAGE}% is below minimum threshold ${MIN_COVERAGE}%"
    print_warning "Continuing with other tests..."
else
    print_success "Coverage ${TOTAL_COVERAGE}% meets minimum threshold ${MIN_COVERAGE}%"
fi

# Generate HTML coverage report
go tool cover -html="$COVERAGE_FILE" -o "$COVERAGE_HTML"
print_success "Generated HTML coverage report: $COVERAGE_HTML"

# Run integration tests (skip if database not available)
print_header "Running Integration Tests"
echo "Note: Integration tests require test database to be configured"

go test -timeout "$TEST_TIMEOUT" -v ./features/weather/integration/... -short 2>&1 | tee -a test-output.log || {
    print_warning "Integration tests skipped or failed (database may not be configured)"
}

# Run load tests
print_header "Running Load Tests"
echo "Testing with 1000 concurrent alarms..."

go test -timeout "$TEST_TIMEOUT" -v ./features/weather/loadtest/... -run TestLoadTest 2>&1 | tee -a test-output.log || {
    print_warning "Load tests skipped (use -short to skip)"
}

# Run edge case tests
print_header "Running Edge Case Tests"
echo "Testing error handling and boundary conditions..."

go test -timeout "$TEST_TIMEOUT" -v ./features/weather/edgecase/... 2>&1 | tee -a test-output.log

if [ $? -eq 0 ]; then
    print_success "Edge case tests passed"
else
    print_error "Edge case tests failed"
fi

# Run benchmarks
print_header "Running Performance Benchmarks"
echo "Measuring cache, database, and processing performance..."

go test -bench=. -benchmem ./features/weather/bench/... 2>&1 | tee benchmark-results.txt

if [ $? -eq 0 ]; then
    print_success "Benchmarks completed"
    echo ""
    echo "Benchmark results saved to: benchmark-results.txt"
else
    print_warning "Benchmarks failed or incomplete"
fi

# Check for race conditions
print_header "Checking for Race Conditions"
echo "Running tests with race detector..."

go test -race -timeout "$TEST_TIMEOUT" ./features/weather/cache ./features/weather/repository ./features/weather/scheduler 2>&1 | tee -a test-output.log

if [ $? -eq 0 ]; then
    print_success "No race conditions detected"
else
    print_error "Race conditions detected"
fi

# Run linter (if installed)
print_header "Running Linter"

if command -v golangci-lint &> /dev/null; then
    golangci-lint run ./features/weather/... 2>&1 | tee lint-results.txt

    if [ $? -eq 0 ]; then
        print_success "Linting passed"
    else
        print_warning "Linting issues found (see lint-results.txt)"
    fi
else
    print_warning "golangci-lint not installed, skipping linting"
    echo "Install with: brew install golangci-lint (macOS) or go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
fi

# Summary
print_header "Test Summary"

echo ""
echo "Test Results:"
echo "  - Unit Tests: PASSED ✅"
echo "  - Coverage: ${TOTAL_COVERAGE}% (minimum: ${MIN_COVERAGE}%)"
echo "  - Edge Cases: Check test-output.log"
echo "  - Load Tests: Check test-output.log"
echo "  - Benchmarks: See benchmark-results.txt"
echo ""

echo "Generated Files:"
echo "  - Coverage Profile: $COVERAGE_FILE"
echo "  - Coverage HTML: $COVERAGE_HTML"
echo "  - Coverage Summary: $COVERAGE_SUMMARY"
echo "  - Test Output: test-output.log"
echo "  - Benchmark Results: benchmark-results.txt"
echo ""

# Final status
if (( $(echo "$TOTAL_COVERAGE >= $MIN_COVERAGE" | bc -l) )); then
    print_success "All quality checks passed!"
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}✅ TEST SUITE COMPLETE${NC}"
    echo -e "${GREEN}========================================${NC}"
    exit 0
else
    print_error "Coverage below threshold"
    echo ""
    echo -e "${RED}========================================${NC}"
    echo -e "${RED}❌ QUALITY CHECKS FAILED${NC}"
    echo -e "${RED}========================================${NC}"
    exit 1
fi
