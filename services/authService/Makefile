.PHONY: help build run test test-e2e test-coverage test-watch clean docker-up docker-down docker-logs

help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the Go application
	go build -o bin/server ./cmd/main.go

run: ## Run the application locally
	go run ./cmd/main.go

test: ## Run all tests
	go test -v ./...

test-e2e: ## Run E2E tests only
	go test -v ./tests/e2e/...

test-coverage: ## Run tests with coverage report
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-watch: ## Run tests in watch mode (requires entr)
	find . -name '*.go' | entr -c make test-e2e

clean: ## Clean build artifacts
	rm -rf bin/ coverage.out coverage.html

docker-up: ## Start Docker Compose services
	docker-compose up -d

docker-down: ## Stop Docker Compose services
	docker-compose down

docker-logs: ## View Docker Compose logs
	docker-compose logs -f

docker-rebuild: ## Rebuild and restart Docker services
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d

tidy: ## Tidy Go modules
	go mod tidy

fmt: ## Format Go code
	go fmt ./...

lint: ## Run linter
	golangci-lint run

swagger-init: ## Initialize/regenerate Swagger documentation
	../../scripts/swagger-init.sh authService
