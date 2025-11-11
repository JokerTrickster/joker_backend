#!/bin/bash

# =============================================================================
# Docker Build Script for Weather Service
# Handles shared module copying for build context
# =============================================================================

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')}"
IMAGE_NAME="${IMAGE_NAME:-weather-scheduler}"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

main() {
    cd "${PROJECT_ROOT}"

    log_info "Building Docker image: ${IMAGE_NAME}:${VERSION}"

    # Check if shared module exists
    if [ ! -d "../../shared" ]; then
        log_error "Shared module not found at ../../shared"
        exit 1
    fi

    # Create temporary build directory
    TEMP_DIR=$(mktemp -d)
    trap "rm -rf ${TEMP_DIR}" EXIT

    log_info "Preparing build context in ${TEMP_DIR}"

    # Copy service files
    cp -r . "${TEMP_DIR}/"

    # Copy shared module to build context
    mkdir -p "${TEMP_DIR}/shared"
    cp -r ../../shared/* "${TEMP_DIR}/shared/"

    # Update go.mod to use local shared module
    cd "${TEMP_DIR}"

    # Build Docker image
    log_info "Building image..."
    docker build \
        --build-arg VERSION="${VERSION}" \
        -t "${IMAGE_NAME}:${VERSION}" \
        -t "${IMAGE_NAME}:latest" \
        -f Dockerfile \
        .

    if [ $? -eq 0 ]; then
        log_info "Docker image built successfully"

        # Show image size
        IMAGE_SIZE=$(docker images "${IMAGE_NAME}:${VERSION}" --format "{{.Size}}")
        log_info "Image size: ${IMAGE_SIZE}"

        # Extract size in MB for comparison
        SIZE_NUM=$(echo "${IMAGE_SIZE}" | grep -oE '[0-9.]+')
        SIZE_UNIT=$(echo "${IMAGE_SIZE}" | grep -oE '[A-Z]+')

        if [ "${SIZE_UNIT}" = "MB" ]; then
            if (( $(echo "$SIZE_NUM < 50" | bc -l) )); then
                log_info "✓ Image size is under 50MB target"
            else
                log_warn "⚠ Image size exceeds 50MB target"
            fi
        elif [ "${SIZE_UNIT}" = "GB" ]; then
            log_error "✗ Image size is too large (>1GB)"
        fi
    else
        log_error "Docker build failed"
        exit 1
    fi

    log_info "Build complete!"
    log_info "Run with: docker run -d --name weather-scheduler ${IMAGE_NAME}:${VERSION}"
}

main "$@"
