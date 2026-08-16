.PHONY: help build run test clean docker-up docker-down lint fmt

help:
	@echo "Available commands:"
	@echo "  make build        - Build the application"
	@echo "  make run          - Run the application"
	@echo "  make test         - Run tests"
	@echo "  make test-integ   - Run integration tests"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make docker-up    - Start Docker containers"
	@echo "  make docker-down  - Stop Docker containers"
	@echo "  make lint         - Run linter"
	@echo "  make fmt          - Format code"
	@echo "  make deps         - Download dependencies"

build:
	@echo "Building application..."
	go build -o bin/api ./cmd/server

run: build
	@echo "Running application..."
	./bin/api

test:
	@echo "Running unit tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-integ:
	@echo "Running integration tests..."
	docker-compose -f docker-compose.yml up -d
	sleep 3
	go test -v -race -tags=integration ./tests/integration/...
	docker-compose -f docker-compose.yml down

clean:
	@echo "Cleaning..."
	rm -rf bin/ coverage.out coverage.html
	go clean

docker-up:
	docker-compose -f docker-compose.yml up -d

docker-down:
	docker-compose -f docker-compose.yml down

lint:
	@echo "Linting..."
	golangci-lint run ./...

fmt:
	@echo "Formatting..."
	go fmt ./...
	goimports -w .

deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy