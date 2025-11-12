#!/bin/bash

# Swagger Documentation Initialization Script
# This script initializes/regenerates swagger documentation for specified service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check if swag is installed
check_swag_installed() {
    if ! command -v swag &> /dev/null; then
        print_error "swag is not installed"
        print_info "Install swag with: go install github.com/swaggo/swag/cmd/swag@latest"
        exit 1
    fi
    print_info "swag version: $(swag --version)"
}

# Get service name from argument or default to authService
SERVICE_NAME="${1:-authService}"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SERVICE_PATH="$PROJECT_ROOT/services/$SERVICE_NAME"

print_info "Initializing Swagger documentation for: $SERVICE_NAME"
print_info "Project root: $PROJECT_ROOT"
print_info "Service path: $SERVICE_PATH"

# Check if service directory exists
if [ ! -d "$SERVICE_PATH" ]; then
    print_error "Service directory not found: $SERVICE_PATH"
    print_info "Available services:"
    ls -1 "$PROJECT_ROOT/services" 2>/dev/null || echo "No services directory found"
    exit 1
fi

# Check if cmd/main.go exists
MAIN_FILE="$SERVICE_PATH/cmd/main.go"
if [ ! -f "$MAIN_FILE" ]; then
    print_error "Main file not found: $MAIN_FILE"
    exit 1
fi

# Check if swag is installed
check_swag_installed

# Change to service directory
cd "$SERVICE_PATH"
print_info "Changed to directory: $(pwd)"

# Create docs directory if it doesn't exist
DOCS_DIR="$SERVICE_PATH/cmd/docs"
if [ ! -d "$DOCS_DIR" ]; then
    print_warning "Docs directory not found, creating: $DOCS_DIR"
    mkdir -p "$DOCS_DIR"
fi

# Run swag init
print_info "Running swag init..."
swag init \
    --generalInfo cmd/main.go \
    --output cmd/docs \
    --parseDependency \
    --parseInternal \
    --parseDepth 1

# Check if swagger files were generated
if [ -f "$DOCS_DIR/docs.go" ] && [ -f "$DOCS_DIR/swagger.json" ] && [ -f "$DOCS_DIR/swagger.yaml" ]; then
    print_info "Swagger documentation generated successfully!"
    print_info "Generated files:"
    ls -lh "$DOCS_DIR"
    print_info ""
    print_info "Files created:"
    print_info "  - $DOCS_DIR/docs.go"
    print_info "  - $DOCS_DIR/swagger.json"
    print_info "  - $DOCS_DIR/swagger.yaml"
else
    print_error "Failed to generate swagger documentation"
    exit 1
fi

print_info "✓ Swagger initialization completed for $SERVICE_NAME"
