# =============================================================================
# Multi-stage build for Weather Service Scheduler
# Final image size target: < 50MB
# =============================================================================

# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Copy source code (includes go.mod replace directive)
COPY . .

# Note: Build requires shared module at ../../shared
# For Docker build, run from project root or copy shared to build context

# Build the scheduler application with optimizations
# -ldflags="-w -s" removes debug info and symbol table
# -trimpath removes file system paths
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags="-w -s -X main.version=${VERSION:-dev}" \
    -trimpath \
    -o scheduler ./cmd/scheduler

# Verify binary size
RUN ls -lh scheduler

# =============================================================================
# Runtime stage - minimal Alpine image
# =============================================================================
FROM alpine:latest

# Add non-root user for security
RUN addgroup -g 1000 weather && \
    adduser -D -u 1000 -G weather weather

WORKDIR /app

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata && \
    rm -rf /var/cache/apk/*

# Copy binary from builder
COPY --from=builder /build/scheduler .

# Change ownership to non-root user
RUN chown -R weather:weather /app

# Switch to non-root user
USER weather

# Expose metrics port
EXPOSE 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9090/health || exit 1

# Run the scheduler
CMD ["./scheduler"]
