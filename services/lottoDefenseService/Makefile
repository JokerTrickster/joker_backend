.PHONY: build run test clean

build:
	go build -o bin/lottoDefenseService ./cmd/main.go

run:
	go run ./cmd/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/

fmt:
	go fmt ./...

tidy:
	go mod tidy
