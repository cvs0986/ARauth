.PHONY: help build run test test-unit test-integration test-coverage lint clean docker-build docker-up docker-down migrate-up migrate-down

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	go build -o bin/iam-api ./cmd/server

run: ## Run the application
	go run ./cmd/server

test: ## Run all tests
	go test ./... -v

test-unit: ## Run unit tests only
	go test ./... -v -short

test-integration: ## Run integration tests
	go test ./... -v -tags=integration

test-coverage: ## Run tests with coverage
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run linter
	golangci-lint run

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

docker-build: ## Build Docker image
	docker build -t nuage-identity/iam-api:latest .

docker-up: ## Start Docker Compose services
	docker-compose up -d

docker-down: ## Stop Docker Compose services
	docker-compose down

migrate-up: ## Run database migrations up
	migrate -path migrations -database "$$DATABASE_URL" up

migrate-down: ## Run database migrations down
	migrate -path migrations -database "$$DATABASE_URL" down

benchmark: ## Run benchmarks
	go test ./... -bench=. -benchmem
