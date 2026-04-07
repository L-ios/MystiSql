.PHONY: all build test test-coverage lint fmt vet clean e2e-setup e2e-test e2e-teardown e2e-reset e2e-check help release build-linux build-darwin build-windows e2e-run e2e-run-frontend e2e-run-backend e2e-run-jdbc e2e-report

BINARY_NAME=mystisql
MAIN_PATH=./cmd/mystisql
TEST_TIMEOUT=5m
E2E_TEST_TIMEOUT=10m
VERSION?=v0.3.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.date=$(BUILD_TIME) -X main.commit=$(GIT_COMMIT)"

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	@go build $(LDFLAGS) -o bin/$(BINARY_NAME) $(MAIN_PATH)

release:
	@echo "Building release binaries for $(VERSION)..."
	@mkdir -p bin/dist
	@echo "Building Linux AMD64..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/dist/$(BINARY_NAME)-$(VERSION)-linux-amd64 $(MAIN_PATH)
	@echo "Building Linux ARM64..."
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/dist/$(BINARY_NAME)-$(VERSION)-linux-arm64 $(MAIN_PATH)
	@echo "Building macOS AMD64..."
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/dist/$(BINARY_NAME)-$(VERSION)-darwin-amd64 $(MAIN_PATH)
	@echo "Building macOS ARM64..."
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/dist/$(BINARY_NAME)-$(VERSION)-darwin-arm64 $(MAIN_PATH)
	@echo "Building Windows AMD64..."
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/dist/$(BINARY_NAME)-$(VERSION)-windows-amd64.exe $(MAIN_PATH)
	@echo "Building Windows ARM64..."
	@GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o bin/dist/$(BINARY_NAME)-$(VERSION)-windows-arm64.exe $(MAIN_PATH)
	@echo "Creating checksums..."
	@cd bin/dist && sha256sum * > checksums.txt
	@echo "Release binaries created in bin/dist/"

build-linux:
	@echo "Building $(BINARY_NAME) for Linux..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

build-darwin:
	@echo "Building $(BINARY_NAME) for macOS..."
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

build-windows:
	@echo "Building $(BINARY_NAME) for Windows..."
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

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

# Unified E2E Testing (using run-e2e-tests.sh)

e2e-run:
	@echo "Running all E2E tests with unified script..."
	@chmod +x scripts/run-e2e-tests.sh
	@./scripts/run-e2e-tests.sh all

e2e-run-frontend:
	@echo "Running frontend E2E tests with unified script..."
	@chmod +x scripts/run-e2e-tests.sh
	@./scripts/run-e2e-tests.sh frontend

e2e-run-backend:
	@echo "Running backend E2E tests with unified script..."
	@chmod +x scripts/run-e2e-tests.sh
	@./scripts/run-e2e-tests.sh backend

e2e-run-jdbc:
	@echo "Running JDBC E2E tests with unified script..."
	@chmod +x scripts/run-e2e-tests.sh
	@./scripts/run-e2e-tests.sh jdbc

e2e-report:
	@echo "Extracting E2E test reports..."
	@if [ -f test-reports/e2e-test-report.tar.gz ]; then \
		tar -xzf test-reports/e2e-test-report.tar.gz -C test-reports/; \
		echo "Reports extracted to test-reports/"; \
		echo "Open test-reports/index.html to view the report"; \
	else \
		echo "No test report found. Run 'make e2e-run' first."; \
	fi

# Legacy E2E Testing Targets (kept for backward compatibility)

e2e-check:
	@echo "Checking e2e test environment..."
	@./scripts/e2e/check-env.sh

e2e-setup:
	@echo "Setting up e2e test environment..."
	@./scripts/e2e/setup-test-env.sh

# Backend E2E tests
e2e-test: e2e-check
	@echo "Running backend e2e tests..."
	@cd e2e-test/backend && go test -v -tags=e2e -timeout $(E2E_TEST_TIMEOUT) ./...

e2e-test-coverage: e2e-check
	@echo "Running backend e2e tests with coverage..."
	@cd e2e-test/backend && go test -v -tags=e2e -cover -coverprofile=e2e-coverage.out -timeout $(E2E_TEST_TIMEOUT) ./...
	@cd e2e-test/backend && go tool cover -html=e2e-coverage.out -o ../../e2e-coverage.html
	@echo "E2E coverage report generated: e2e-coverage.html"

# Frontend E2E tests
e2e-frontend:
	@echo "Running frontend e2e tests..."
	@cd web && npm run test:e2e

e2e-frontend-ui:
	@echo "Running frontend e2e tests in UI mode..."
	@cd web && npm run test:e2e:ui

e2e-frontend-debug:
	@echo "Running frontend e2e tests in debug mode..."
	@cd web && npm run test:e2e:debug

# JDBC E2E tests
e2e-jdbc:
	@echo "Running JDBC e2e tests..."
	@cd jdbc && ./gradlew test --tests "io.github.mystisql.jdbc.e2e.*"

# Run all E2E tests
e2e-all: e2e-test e2e-frontend e2e-jdbc
	@echo "All E2E tests completed!"

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
	@echo "  build            - Build the binary for current OS"
	@echo "  release          - Build release binaries for all platforms (Linux, macOS, Windows)"
	@echo "  build-linux      - Build for Linux AMD64"
	@echo "  build-darwin     - Build for macOS (AMD64 + ARM64)"
	@echo "  build-windows    - Build for Windows AMD64"
	@echo "  test             - Run unit tests"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  lint             - Run linter"
	@echo "  fmt              - Format code"
	@echo "  vet              - Run go vet"
	@echo "  clean            - Clean build artifacts"
	@echo ""
	@echo "Unified E2E Testing (Recommended):"
	@echo "  e2e-run          - Run all E2E tests (frontend + backend + jdbc)"
	@echo "  e2e-run-frontend - Run frontend E2E tests only"
	@echo "  e2e-run-backend  - Run backend E2E tests only"
	@echo "  e2e-run-jdbc     - Run JDBC E2E tests only"
	@echo "  e2e-report       - Extract and view test reports"
	@echo ""
	@echo "Legacy E2E Testing (Backward Compatibility):"
	@echo "  e2e-check        - Check e2e test environment"
	@echo "  e2e-setup        - Setup e2e test environment"
	@echo "  e2e-test         - Run backend e2e tests"
	@echo "  e2e-test-coverage- Run backend e2e tests with coverage"
	@echo "  e2e-frontend     - Run frontend e2e tests"
	@echo "  e2e-frontend-ui  - Run frontend e2e tests in UI mode"
	@echo "  e2e-frontend-debug- Run frontend e2e tests in debug mode"
	@echo "  e2e-jdbc         - Run JDBC e2e tests"
	@echo "  e2e-all          - Run all e2e tests (backend + frontend + jdbc)"
	@echo "  e2e-teardown     - Teardown e2e test environment"
	@echo "  e2e-reset        - Reset all test databases"
	@echo "  e2e-reset-mysql  - Reset MySQL test database"
	@echo "  e2e-reset-postgres- Reset PostgreSQL test database"
	@echo ""
	@echo "Development:"
	@echo "  dev              - Build and start development server"
	@echo "  docker-build     - Build Docker image"
	@echo "  docker-run       - Run Docker container"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION          - Version to embed in binary (default: v0.3.0)"
