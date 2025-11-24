# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory to match project structure
WORKDIR /app/services/authService

# Copy shared module to match replace directive path (../../shared)
COPY shared /app/shared

# Copy migrations
COPY migrations /app/migrations

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /root/

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/services/authService/main .

# Copy migrations from builder
COPY --from=builder /app/migrations ./migrations

# Port is configured via environment variable
EXPOSE 18081

# Run the application
CMD ["./main"]
