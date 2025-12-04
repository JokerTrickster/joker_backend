FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory to match project structure
WORKDIR /app/services/cloudRepositoryService

# Copy shared module to match replace directive path (../../shared)
COPY shared /app/shared

# Copy go mod files
COPY services/cloudRepositoryService/go.mod services/cloudRepositoryService/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY services/cloudRepositoryService/ .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS and FFmpeg for video processing
RUN apk --no-cache add ca-certificates ffmpeg

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/services/cloudRepositoryService/main .

# Port is configured via environment variable
EXPOSE 18080

CMD ["./main"]
