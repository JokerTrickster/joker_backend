#!/bin/bash

# =============================================================================
# Weather Service Scheduler Build Script
# Builds binaries for multiple platforms with version tagging
# =============================================================================

set -e

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="${PROJECT_ROOT}/bin"
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')}"
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')

# Build flags
LDFLAGS="-w -s -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Print build info
print_build_info() {
    log_info "Building Weather Scheduler"
    echo "  Version:    ${VERSION}"
    echo "  Build Time: ${BUILD_TIME}"
    echo "  Git Commit: ${GIT_COMMIT}"
    echo ""
}

# Build for specific platform
build_platform() {
    local os=$1
    local arch=$2
    local output_name="scheduler"

    if [ "${os}" = "windows" ]; then
        output_name="scheduler.exe"
    fi

    local output_path="${BIN_DIR}/${os}_${arch}/${output_name}"

    log_info "Building for ${os}/${arch}..."

    mkdir -p "$(dirname "${output_path}")"

    CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} go build \
        -a -installsuffix cgo \
        -ldflags="${LDFLAGS}" \
        -trimpath \
        -o "${output_path}" \
        ./cmd/scheduler

    if [ $? -eq 0 ]; then
        local size=$(du -h "${output_path}" | cut -f1)
        log_info "Built ${os}/${arch}: ${output_path} (${size})"
    else
        log_error "Failed to build ${os}/${arch}"
        return 1
    fi
}

# Clean previous builds
clean() {
    log_info "Cleaning previous builds..."
    rm -rf "${BIN_DIR}"
}

# Main build process
main() {
    cd "${PROJECT_ROOT}"

    print_build_info

    # Parse arguments
    if [ "$1" = "clean" ]; then
        clean
        exit 0
    fi

    # Clean before building
    clean

    # Build for target platforms
    if [ -z "$1" ]; then
        # Build for all platforms
        log_info "Building for all platforms..."
        build_platform linux amd64
        build_platform linux arm64
        build_platform darwin amd64
        build_platform darwin arm64
        build_platform windows amd64
    else
        # Build for specific platform
        case "$1" in
            linux)
                build_platform linux amd64
                ;;
            linux-arm64)
                build_platform linux arm64
                ;;
            darwin)
                build_platform darwin amd64
                ;;
            darwin-arm64)
                build_platform darwin arm64
                ;;
            windows)
                build_platform windows amd64
                ;;
            *)
                log_error "Unknown platform: $1"
                echo "Usage: $0 [linux|linux-arm64|darwin|darwin-arm64|windows|clean]"
                exit 1
                ;;
        esac
    fi

    log_info "Build completed successfully!"
    log_info "Binaries available in: ${BIN_DIR}"
}

main "$@"
