.PHONY: help build run test clean deps dev setup-db

# Colors for output
GREEN  := \033[0;32m
YELLOW := \033[0;33m
NC     := \033[0m # No Color

help: ## Show this help message
	@echo '$(GREEN)Karbos Server - Available Commands:$(NC)'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-15s$(NC) %s\n", $$1, $$2}'

deps: ## Install Go dependencies
	@echo "$(GREEN)Installing dependencies...$(NC)"
	go mod download
	go mod tidy

build: ## Build the server binary
	@echo "$(GREEN)Building server...$(NC)"
	go build -o bin/karbos-server cmd/api/main.go
	@echo "$(GREEN)✓ Build complete: bin/karbos-server$(NC)"

build-worker: ## Build the worker binary
	@echo "$(GREEN)Building worker...$(NC)"
	go build -o bin/karbos-worker cmd/worker/main.go
	@echo "$(GREEN)✓ Build complete: bin/karbos-worker$(NC)"

build-all: build build-worker ## Build both server and worker binaries
	@echo "$(GREEN)✓ All binaries built$(NC)"

build-seeder: ## Build the demo data seeder
	@echo "$(GREEN)Building demo data seeder...$(NC)"
	go build -o bin/karbos-seeder cmd/seeder/main.go
	@echo "$(GREEN)✓ Build complete: bin/karbos-seeder$(NC)"

run: ## Run the server
	@echo "$(GREEN)Starting Karbos server...$(NC)"
	go run cmd/api/main.go

run-worker: ## Run the worker
	@echo "$(GREEN)Starting Karbos worker...$(NC)"
	go run cmd/worker/main.go

seed-demo: ## Seed demo data (50 historical jobs + 5 active jobs)
	@echo "$(GREEN)Seeding demo data...$(NC)"
	@echo "$(YELLOW)Note: Ensure database is set up and API server will be running$(NC)"
	go run cmd/seeder/main.go

dev: ## Run with auto-reload (requires air: go install github.com/cosmtrek/air@latest)
	@echo "$(GREEN)Starting development server with auto-reload...$(NC)"
	air

test: ## Run tests
	@echo "$(GREEN)Running tests...$(NC)"
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report: coverage.html$(NC)"

clean: ## Clean build artifacts
	@echo "$(GREEN)Cleaning...$(NC)"
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "$(GREEN)✓ Clean complete$(NC)"

setup-db: ## Set up the PostgreSQL database schema
	@echo "$(GREEN)Setting up database...$(NC)"
	@if [ -z "$$DATABASE_URL" ]; then \
		echo "$(YELLOW)Warning: DATABASE_URL not set. Please set it first.$(NC)"; \
		exit 1; \
	fi
	psql "$$DATABASE_URL" -f database/schema.sql
	@echo "$(GREEN)✓ Database schema created$(NC)"

check-redis: ## Check if Redis is running
	@echo "$(GREEN)Checking Redis connection...$(NC)"
	@redis-cli ping > /dev/null 2>&1 && echo "$(GREEN)✓ Redis is running$(NC)" || echo "$(YELLOW)⚠ Redis is not running. Start it with: redis-server$(NC)"

check-postgres: ## Check if PostgreSQL is accessible
	@echo "$(GREEN)Checking PostgreSQL connection...$(NC)"
	@if [ -z "$$DATABASE_URL" ]; then \
		echo "$(YELLOW)⚠ DATABASE_URL not set$(NC)"; \
	else \
		psql "$$DATABASE_URL" -c "SELECT 1;" > /dev/null 2>&1 && echo "$(GREEN)✓ PostgreSQL is accessible$(NC)" || echo "$(YELLOW)⚠ Cannot connect to PostgreSQL$(NC)"; \
	fi

check: check-redis check-postgres ## Check if all services are running

docker-redis: ## Start Redis in Docker
	@echo "$(GREEN)Starting Redis in Docker...$(NC)"
	docker run -d --name karbos-redis -p 6379:6379 redis:alpine
	@echo "$(GREEN)✓ Redis started$(NC)"

docker-redis-stop: ## Stop Redis Docker container
	@echo "$(GREEN)Stopping Redis...$(NC)"
	docker stop karbos-redis
	docker rm karbos-redis
	@echo "$(GREEN)✓ Redis stopped$(NC)"

fmt: ## Format Go code
	@echo "$(GREEN)Formatting code...$(NC)"
	go fmt ./...
	@echo "$(GREEN)✓ Code formatted$(NC)"

lint: ## Run linter (requires golangci-lint)
	@echo "$(GREEN)Running linter...$(NC)"
	golangci-lint run

mod-update: ## Update Go dependencies
	@echo "$(GREEN)Updating dependencies...$(NC)"
	go get -u ./...
	go mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(NC)"
