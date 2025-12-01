.PHONY: help build run test clean install migrate seed

# Variables
BINARY_NAME=neonexframework
MAIN_PATH=./main.go

help: ## แสดงคำสั่งที่ใช้ได้
	@echo "Available commands:"
	@echo "  make build      - Build the application"
	@echo "  make run        - Run the application"
	@echo "  make test       - Run tests"
	@echo "  make clean      - Clean build files"
	@echo "  make install    - Install dependencies"
	@echo "  make migrate    - Run database migrations"
	@echo "  make seed       - Seed the database"
	@echo "  make dev        - Run with hot reload"

build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) $(MAIN_PATH)

run: ## Run the application
	@echo "Running $(BINARY_NAME)..."
	go run $(MAIN_PATH)

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build files
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -rf tmp/

install: ## Install dependencies
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

migrate: ## Run database migrations
	@echo "Running migrations..."
	go run $(MAIN_PATH) migrate

seed: ## Seed the database
	@echo "Seeding database..."
	go run $(MAIN_PATH) seed

dev: ## Run with hot reload (requires air)
	@echo "Starting development server with hot reload..."
	@if ! command -v air > /dev/null; then \
		echo "Installing air..."; \
		go install github.com/cosmtrek/air@latest; \
	fi
	air

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env $(BINARY_NAME):latest

docker-compose-up: ## Start with Docker Compose
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

docker-compose-down: ## Stop Docker Compose services
	@echo "Stopping services..."
	docker-compose down

lint: ## Run linter
	@echo "Running linter..."
	@if ! command -v golangci-lint > /dev/null; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	golangci-lint run

format: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

mod-update: ## Update dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

security-check: ## Run security check
	@echo "Running security check..."
	@if ! command -v gosec > /dev/null; then \
		echo "Installing gosec..."; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
	fi
	gosec ./...
