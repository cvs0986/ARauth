.PHONY: build run test lint clean migrate-up migrate-down docker-up docker-down help

# Variables
BINARY_NAME=iam-api
MAIN_PATH=./cmd/server
MIGRATIONS_PATH=./migrations
DATABASE_URL?=postgres://iam_user:change-me@localhost:5432/iam?sslmode=disable

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: bin/$(BINARY_NAME)"

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	$(GOCMD) run $(MAIN_PATH)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linters
lint:
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it with: make install-tools"; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	fi

# Run database migrations up
migrate-up:
	@echo "Running database migrations up..."
	@if command -v migrate > /dev/null; then \
		migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" up; \
	else \
		echo "migrate tool not found. Install it with: make install-tools"; \
	fi

# Run database migrations down
migrate-down:
	@echo "Rolling back database migrations..."
	@if command -v migrate > /dev/null; then \
		migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down; \
	else \
		echo "migrate tool not found. Install it with: make install-tools"; \
	fi

# Create new migration
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $$name

# Start Docker Compose services
docker-up:
	@echo "Starting Docker Compose services..."
	docker-compose up -d
	@echo "Services started. Waiting for health checks..."
	@sleep 5
	@echo "Services are ready!"

# Stop Docker Compose services
docker-down:
	@echo "Stopping Docker Compose services..."
	docker-compose down

# Stop and remove volumes
docker-clean:
	@echo "Stopping Docker Compose services and removing volumes..."
	docker-compose down -v

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@bash scripts/install-dev-tools.sh

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	$(GOCMD) clean

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Verify dependencies
verify:
	@echo "Verifying dependencies..."
	$(GOMOD) verify

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  run            - Run the application"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  lint           - Run linters"
	@echo "  fmt            - Format code"
	@echo "  migrate-up     - Run database migrations up"
	@echo "  migrate-down   - Rollback database migrations"
	@echo "  migrate-create - Create a new migration"
	@echo "  docker-up      - Start Docker Compose services"
	@echo "  docker-down    - Stop Docker Compose services"
	@echo "  docker-clean   - Stop services and remove volumes"
	@echo "  install-tools  - Install development tools"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Download dependencies"
	@echo "  verify         - Verify dependencies"
	@echo "  help           - Show this help message"

