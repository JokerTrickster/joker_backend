FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app/services/lottoDefenseService
COPY shared /app/shared
COPY services/lottoDefenseService/go.mod services/lottoDefenseService/go.sum ./
RUN go mod download
COPY services/lottoDefenseService ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/services/lottoDefenseService/main .

EXPOSE 18082
CMD ["./main"]
