# Makefile for Jackpot Bet Admin Stats API

# Variables
APP_NAME=casino-analytics
API_ENTRY=cmd/api/main.go
SEEDER_ENTRY=cmd/seeder/main.go
BINARY_DIR=bin

.PHONY: all build run seed test clean docker-up docker-down tidy help

help:
	@echo "Available commands:"
	@echo "  make run         - Run the API"
	@echo "  make seed        - Run the data seeder"
	@echo "  make build       - Build the binaries"
	@echo "  make test        - Run tests"
	@echo "  make tidy        - Tidy go modules"
	@echo "  make clean       - Remove built binaries"
	@echo "  make docker-run   - Start docker containers (API + DB)"
	@echo "  make docker-seed  - Run the data seeder in Docker"
	@echo "  make docker-down  - Stop docker containers"
	@echo ""
	@echo "Troubleshooting:"
	@echo "  If docker containers fail to start, ensure Docker Desktop is running and ports 5432/8080 are free."

all: help

build:
	@echo "Building binaries..."
	@if [ ! -d $(BINARY_DIR) ]; then mkdir $(BINARY_DIR); fi
	go build -o $(BINARY_DIR)/api $(API_ENTRY)
	go build -o $(BINARY_DIR)/seeder $(SEEDER_ENTRY)

run:
	@echo "Running API..."
	go run $(API_ENTRY)

seed:
	@echo "Running Seeder..."
	go run $(SEEDER_ENTRY)

test:
	@echo "Running tests..."
	go test ./...

tidy:
	@echo "Tidying go modules..."
	go mod tidy

clean:
	@echo "Cleaning up..."
	@if [ -d $(BINARY_DIR) ]; then rm -rf $(BINARY_DIR); fi

docker-run:
	@echo "Starting Docker containers..."
	docker-compose up -d api

# Alias for backward compatibility
docker-up: docker-run

docker-seed:
	@echo "Running Seeder in Docker..."
	docker-compose run --rm seeder

docker-down:
	@echo "Stopping Docker containers..."
	docker-compose down
