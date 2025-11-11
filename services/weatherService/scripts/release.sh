#!/bin/bash

# =============================================================================
# Weather Service Scheduler Release Script
# Builds binaries, creates Docker image, and optionally pushes to registry
# =============================================================================

set -e

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')}"
DOCKER_REGISTRY="${DOCKER_REGISTRY:-}"
IMAGE_NAME="${IMAGE_NAME:-weather-scheduler}"
PLATFORMS="linux/amd64,linux/arm64"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Print release info
print_release_info() {
    log_info "Weather Scheduler Release"
    echo "  Version:        ${VERSION}"
    echo "  Image Name:     ${IMAGE_NAME}"
    echo "  Registry:       ${DOCKER_REGISTRY:-none}"
    echo "  Platforms:      ${PLATFORMS}"
    echo ""
}

# Check prerequisites
check_prerequisites() {
    log_step "Checking prerequisites..."

    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        exit 1
    fi

    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi

    log_info "Prerequisites check passed"
}

# Build binaries
build_binaries() {
    log_step "Building binaries..."

    bash "${PROJECT_ROOT}/scripts/build.sh"

    if [ $? -ne 0 ]; then
        log_error "Binary build failed"
        exit 1
    fi
}

# Build Docker image
build_docker_image() {
    log_step "Building Docker image..."

    cd "${PROJECT_ROOT}"

    # Determine full image name
    local full_image_name="${IMAGE_NAME}:${VERSION}"
    if [ -n "${DOCKER_REGISTRY}" ]; then
        full_image_name="${DOCKER_REGISTRY}/${IMAGE_NAME}:${VERSION}"
    fi

    log_info "Building image: ${full_image_name}"

    # Copy shared module to build context
    if [ ! -d "${PROJECT_ROOT}/shared" ]; then
        log_warn "Copying shared module to build context..."
        cp -r "${PROJECT_ROOT}/../../shared" "${PROJECT_ROOT}/"
    fi

    # Build image with version
    docker build \
        --build-arg VERSION="${VERSION}" \
        -t "${full_image_name}" \
        -t "${IMAGE_NAME}:latest" \
        -f Dockerfile \
        .

    if [ $? -eq 0 ]; then
        log_info "Docker image built successfully"

        # Show image size
        local image_size=$(docker images "${full_image_name}" --format "{{.Size}}")
        log_info "Image size: ${image_size}"

        # Verify size is under 50MB
        local size_mb=$(docker images "${full_image_name}" --format "{{.Size}}" | sed 's/MB//')
        if (( $(echo "$size_mb < 50" | bc -l) )); then
            log_info "Image size is under 50MB target ✓"
        else
            log_warn "Image size exceeds 50MB target"
        fi
    else
        log_error "Docker image build failed"
        exit 1
    fi
}

# Tag Docker image
tag_docker_image() {
    log_step "Tagging Docker image..."

    local base_image="${IMAGE_NAME}:${VERSION}"
    if [ -n "${DOCKER_REGISTRY}" ]; then
        base_image="${DOCKER_REGISTRY}/${IMAGE_NAME}:${VERSION}"
    fi

    # Tag as latest
    docker tag "${base_image}" "${IMAGE_NAME}:latest"

    # Tag with major.minor if version is semver
    if [[ "${VERSION}" =~ ^v?([0-9]+)\.([0-9]+)\.([0-9]+) ]]; then
        local major="${BASH_REMATCH[1]}"
        local minor="${BASH_REMATCH[2]}"

        docker tag "${base_image}" "${IMAGE_NAME}:${major}"
        docker tag "${base_image}" "${IMAGE_NAME}:${major}.${minor}"

        log_info "Tagged as: ${major}, ${major}.${minor}, latest"
    fi
}

# Push Docker image to registry
push_docker_image() {
    if [ -z "${DOCKER_REGISTRY}" ]; then
        log_warn "No registry specified, skipping push"
        return 0
    fi

    log_step "Pushing Docker image to registry..."

    local full_image_name="${DOCKER_REGISTRY}/${IMAGE_NAME}:${VERSION}"

    log_info "Pushing ${full_image_name}..."
    docker push "${full_image_name}"

    log_info "Pushing ${DOCKER_REGISTRY}/${IMAGE_NAME}:latest..."
    docker push "${DOCKER_REGISTRY}/${IMAGE_NAME}:latest"

    if [ $? -eq 0 ]; then
        log_info "Image pushed successfully"
    else
        log_error "Failed to push image"
        exit 1
    fi
}

# Create release artifacts
create_artifacts() {
    log_step "Creating release artifacts..."

    local release_dir="${PROJECT_ROOT}/release/${VERSION}"
    mkdir -p "${release_dir}"

    # Copy binaries
    cp -r "${PROJECT_ROOT}/bin"/* "${release_dir}/"

    # Create checksums
    cd "${release_dir}"
    find . -type f -name "scheduler*" -exec sha256sum {} \; > checksums.txt

    log_info "Release artifacts created in: ${release_dir}"
}

# Main release process
main() {
    cd "${PROJECT_ROOT}"

    print_release_info

    # Parse arguments
    local skip_build=false
    local skip_push=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-build)
                skip_build=true
                shift
                ;;
            --skip-push)
                skip_push=true
                shift
                ;;
            --registry=*)
                DOCKER_REGISTRY="${1#*=}"
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                echo "Usage: $0 [--skip-build] [--skip-push] [--registry=REGISTRY]"
                exit 1
                ;;
        esac
    done

    # Execute release steps
    check_prerequisites

    if [ "${skip_build}" = false ]; then
        build_binaries
        create_artifacts
    fi

    build_docker_image
    tag_docker_image

    if [ "${skip_push}" = false ]; then
        push_docker_image
    fi

    log_info "Release completed successfully!"
    echo ""
    log_info "Summary:"
    echo "  Version:       ${VERSION}"
    echo "  Binaries:      ${PROJECT_ROOT}/bin/"
    echo "  Artifacts:     ${PROJECT_ROOT}/release/${VERSION}/"
    echo "  Docker Image:  ${IMAGE_NAME}:${VERSION}"

    if [ -n "${DOCKER_REGISTRY}" ] && [ "${skip_push}" = false ]; then
        echo "  Registry:      ${DOCKER_REGISTRY}/${IMAGE_NAME}:${VERSION}"
    fi
}

main "$@"
