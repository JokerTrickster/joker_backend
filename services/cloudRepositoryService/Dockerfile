FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
COPY shared/ ./shared/

RUN go mod download

# Copy source code
COPY services/cloudRepositoryService/ ./services/cloudRepositoryService/

# Build the application
WORKDIR /app/services/cloudRepositoryService
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/services/cloudRepositoryService/main .

EXPOSE 8080

CMD ["./main"]
