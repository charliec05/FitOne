# FitONEX Backend Makefile

.PHONY: help run build test clean migrate-up migrate-down docker-build docker-run seed-dev lint run-jobs

# Default target
help:
	@echo "Available commands:"
	@echo "  run          - Run the development server"
	@echo "  build        - Build the application"
	@echo "  test         - Run tests"
	@echo "  lint         - Run linters"
	@echo "  seed-dev     - Seed the development database"
	@echo "  run-jobs     - Run background jobs"
	@echo "  clean        - Clean build artifacts"
	@echo "  migrate-up   - Run database migrations up"
	@echo "  migrate-down - Rollback database migrations"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"

# Development server
run:
	@echo "Starting development server..."
	go run cmd/api/main.go

# Build the application
build:
	@echo "Building application..."
	go build -o bin/api cmd/api/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

lint:
	@echo "Running linters..."
	go vet ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Database migrations
migrate-up:
	@echo "Running database migrations up..."
	@cd infra && docker-compose up -d postgres
	@sleep 5
	@DATABASE_URL="postgres://fitonex:password@localhost:5432/fitonex?sslmode=disable" go run cmd/api/main.go migrate-up

migrate-down:
	@echo "Rolling back database migrations..."
	@DATABASE_URL="postgres://fitonex:password@localhost:5432/fitonex?sslmode=disable" go run cmd/api/main.go migrate-down

# Docker commands
docker-build:
	@echo "Building Docker image..."
	docker build -t fitonex-api .

docker-run:
	@echo "Starting services with Docker Compose..."
	@cd infra && docker-compose up --build

# Development setup
setup:
	@echo "Setting up development environment..."
	@cp env.example .env
	@cd infra && docker-compose up -d postgres redis minio
	@sleep 15
	@make migrate-up
	@make seed
	@echo "Development environment ready!"

# Seed database with sample data
seed:
	@$(MAKE) seed-dev

seed-dev:
	@echo "Seeding development database..."
	go run cmd/admin/seed.go

run-jobs:
	@echo "Starting background jobs..."
	go run cmd/jobs/pricing.go

# Stop services
stop:
	@echo "Stopping services..."
	@cd infra && docker-compose down

# View logs
logs:
	@cd infra && docker-compose logs -f api
