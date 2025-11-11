.PHONY: help build run test test-e2e test-coverage test-watch clean docker-up docker-down docker-logs test-integration test-unit build-scheduler run-scheduler docker-test-up docker-test-down inspect-db inspect-redis

help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the API application
	go build -o bin/server ./cmd/main.go

build-scheduler: ## Build the scheduler application
	go build -o bin/scheduler ./cmd/scheduler/main.go

build-all: build build-scheduler ## Build all applications

run: ## Run the API application locally
	go run ./cmd/main.go

run-scheduler: ## Run the scheduler application (requires env vars)
	@if [ -z "$$FCM_CREDENTIALS_PATH" ]; then \
		echo "Error: FCM_CREDENTIALS_PATH is required"; \
		exit 1; \
	fi
	go run ./cmd/scheduler/main.go

test: ## Run all tests
	go test -v ./...

test-unit: ## Run unit tests only (skip integration)
	go test -v -short ./...

test-integration: docker-test-up ## Run integration tests with real services
	@echo "Waiting for services to be ready..."
	@sleep 10
	@echo "Running integration tests..."
	go test -v -timeout 5m ./features/weather/integration/... || (make docker-test-down && exit 1)
	@make docker-test-down

test-e2e: ## Run E2E tests only
	go test -v ./tests/e2e/...

test-coverage: ## Run tests with coverage report
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-watch: ## Run tests in watch mode (requires entr)
	find . -name '*.go' | entr -c make test-e2e

clean: ## Clean build artifacts and test containers
	rm -rf bin/ coverage.out coverage.html
	docker-compose -f docker-compose.test.yml down -v

docker-up: ## Start production Docker Compose services
	docker-compose up -d

docker-down: ## Stop production Docker Compose services
	docker-compose down

docker-test-up: ## Start test containers (MySQL, Redis)
	docker-compose -f docker-compose.test.yml up -d
	@echo "Test containers starting..."

docker-test-down: ## Stop and remove test containers
	docker-compose -f docker-compose.test.yml down -v

docker-logs: ## View Docker Compose logs
	docker-compose logs -f

docker-rebuild: ## Rebuild and restart Docker services
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d

inspect-db: ## Inspect test database
	docker exec -it joker_mysql_test mysql -utest_user -ptest_password joker_test

inspect-redis: ## Inspect test Redis
	docker exec -it joker_redis_test redis-cli

dev-scheduler: docker-test-up ## Run scheduler in development mode
	@echo "Starting scheduler in development mode..."
	@sleep 5
	ENV=development LOG_LEVEL=debug \
	DB_HOST=localhost DB_PORT=3307 DB_USER=test_user DB_PASSWORD=test_password DB_NAME=joker_test \
	REDIS_HOST=localhost REDIS_PORT=6380 REDIS_PASSWORD="" \
	FCM_CREDENTIALS_PATH=./fcm-credentials.json \
	SCHEDULER_INTERVAL=1m \
	go run ./cmd/scheduler/main.go

tidy: ## Tidy Go modules
	go mod tidy

fmt: ## Format Go code
	go fmt ./...

lint: ## Run linter
	golangci-lint run || go vet ./...

# Deployment targets
deploy-systemd: build-scheduler ## Deploy to systemd (requires sudo)
	@echo "Deploying to systemd..."
	sudo mkdir -p /opt/weather-scheduler/{bin,config,logs}
	sudo cp bin/scheduler /opt/weather-scheduler/bin/
	sudo cp .env /opt/weather-scheduler/
	sudo cp fcm-credentials.json /opt/weather-scheduler/config/ || true
	sudo cp deploy/systemd/weather-scheduler.service /etc/systemd/system/
	sudo systemctl daemon-reload
	@echo "Deployment complete. Run 'sudo systemctl start weather-scheduler' to start"

deploy-k8s: ## Deploy to Kubernetes
	@echo "Deploying to Kubernetes..."
	kubectl apply -f deploy/k8s/namespace.yaml
	kubectl apply -f deploy/k8s/configmap.yaml
	kubectl apply -f deploy/k8s/secret.yaml
	kubectl apply -f deploy/k8s/deployment.yaml
	kubectl apply -f deploy/k8s/service.yaml
	kubectl apply -f deploy/k8s/hpa.yaml
	@echo "Deployment complete. Check status with 'kubectl get pods -n joker-backend'"

deploy-docker: ## Deploy using Docker Compose
	@echo "Deploying with Docker Compose..."
	docker-compose build
	docker-compose up -d
	@echo "Deployment complete. Check logs with 'docker-compose logs -f'"

undeploy-k8s: ## Remove Kubernetes deployment
	kubectl delete -f deploy/k8s/ || true

release: ## Build release artifacts and Docker image
	./scripts/release.sh
