.PHONY: all build test test-coverage lint fmt vet clean e2e-setup e2e-test e2e-teardown e2e-reset e2e-check help

BINARY_NAME=mystisql
MAIN_PATH=./cmd/mystisql
TEST_TIMEOUT=5m
E2E_TEST_TIMEOUT=10m

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

build-linux:
	@echo "Building $(BINARY_NAME) for Linux..."
	@GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

test:
	@echo "Running unit tests..."
	@go test -v -timeout $(TEST_TIMEOUT) ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -cover -coverprofile=coverage.out -timeout $(TEST_TIMEOUT) ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint:
	@echo "Running linter..."
	@golangci-lint run

fmt:
	@echo "Formatting code..."
	@go fmt ./...

vet:
	@echo "Running go vet..."
	@go vet ./...

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

# E2E Testing Targets

e2e-check:
	@echo "Checking e2e test environment..."
	@./scripts/e2e/check-env.sh

e2e-setup:
	@echo "Setting up e2e test environment..."
	@./scripts/e2e/setup-test-env.sh

e2e-test: e2e-check
	@echo "Running e2e tests..."
	@go test -v -tags=e2e -timeout $(E2E_TEST_TIMEOUT) ./test/e2e/...

e2e-test-coverage: e2e-check
	@echo "Running e2e tests with coverage..."
	@go test -v -tags=e2e -cover -coverprofile=e2e-coverage.out -timeout $(E2E_TEST_TIMEOUT) ./test/e2e/...
	@go tool cover -html=e2e-coverage.out -o e2e-coverage.html
	@echo "E2E coverage report generated: e2e-coverage.html"

e2e-teardown:
	@echo "Tearing down e2e test environment..."
	@./scripts/e2e/teardown-test-env.sh

e2e-reset:
	@echo "Resetting e2e test databases..."
	@./scripts/e2e/reset-db.sh

e2e-reset-mysql:
	@echo "Resetting MySQL test database..."
	@./scripts/e2e/reset-db.sh mysql

e2e-reset-postgres:
	@echo "Resetting PostgreSQL test database..."
	@./scripts/e2e/reset-db.sh postgres

# Convenience targets

dev: build
	@echo "Starting development server..."
	@./bin/$(BINARY_NAME) --config config/config.yaml

docker-build:
	@echo "Building Docker image..."
	@docker build -t mystisql:latest .

docker-run:
	@echo "Running Docker container..."
	@docker run -p 8080:8080 mystisql:latest

help:
	@echo "Available targets:"
	@echo "  build            - Build the binary"
	@echo "  build-linux      - Build for Linux AMD64"
	@echo "  test             - Run unit tests"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  lint             - Run linter"
	@echo "  fmt              - Format code"
	@echo "  vet              - Run go vet"
	@echo "  clean            - Clean build artifacts"
	@echo ""
	@echo "E2E Testing:"
	@echo "  e2e-check        - Check e2e test environment"
	@echo "  e2e-setup        - Setup e2e test environment"
	@echo "  e2e-test         - Run e2e tests"
	@echo "  e2e-test-coverage- Run e2e tests with coverage"
	@echo "  e2e-teardown     - Teardown e2e test environment"
	@echo "  e2e-reset        - Reset all test databases"
	@echo "  e2e-reset-mysql  - Reset MySQL test database"
	@echo "  e2e-reset-postgres- Reset PostgreSQL test database"
	@echo ""
	@echo "Development:"
	@echo "  dev              - Build and start development server"
	@echo "  docker-build     - Build Docker image"
	@echo "  docker-run       - Run Docker container"
