.PHONY: build run test clean docker-build docker-run

# Build the application
build:
	go build -o bin/cloudRepositoryService ./cmd/main.go

# Run the application
run:
	go run ./cmd/main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Build docker image
docker-build:
	docker build -t cloud-repository-service:latest -f Dockerfile ../..

# Run docker container
docker-run:
	docker run -p 8080:8080 --env-file .env cloud-repository-service:latest

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run with hot reload (requires air)
dev:
	air
