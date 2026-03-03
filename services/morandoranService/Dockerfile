FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app/services/morandoranService

COPY shared /app/shared
COPY migrations /app/migrations
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

FROM alpine:latest

WORKDIR /root/

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/services/morandoranService/main .
COPY --from=builder /app/migrations ./migrations

EXPOSE 18083

CMD ["./main"]
