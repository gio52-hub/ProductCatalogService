.PHONY: all build test run clean proto migrate emulator-up emulator-down setup-db setup-emulator test-unit

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
BINARY_NAME=product-catalog-service

# Spanner emulator settings
SPANNER_EMULATOR_HOST=localhost:9010
SPANNER_PROJECT=test-project
SPANNER_INSTANCE=test-instance
SPANNER_DATABASE=test-database

all: build

build:
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/server

test:
	SPANNER_EMULATOR_HOST=$(SPANNER_EMULATOR_HOST) $(GOTEST) -v ./...

test-e2e:
	SPANNER_EMULATOR_HOST=$(SPANNER_EMULATOR_HOST) $(GOTEST) -v ./tests/e2e/...

run:
	SPANNER_EMULATOR_HOST=$(SPANNER_EMULATOR_HOST) $(GOCMD) run ./cmd/server

clean:
	rm -f $(BINARY_NAME)

# Generate protobuf code
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/product/v1/product_service.proto

# Docker compose commands
emulator-up:
	docker-compose up -d

emulator-down:
	docker-compose down

# Setup Spanner database using gcloud CLI with emulator
setup-db:
	@echo "Setting up Spanner emulator database..."
	gcloud config configurations create emulator --no-activate 2>/dev/null || true
	gcloud config configurations activate emulator
	gcloud config set auth/disable_credentials true
	gcloud config set project $(SPANNER_PROJECT)
	gcloud config set api_endpoint_overrides/spanner http://$(SPANNER_EMULATOR_HOST)/
	gcloud spanner instances create $(SPANNER_INSTANCE) \
		--config=emulator-config \
		--description="Test Instance" \
		--nodes=1
	gcloud spanner databases create $(SPANNER_DATABASE) \
		--instance=$(SPANNER_INSTANCE) \
		--ddl-file=migrations/001_initial_schema.sql

# Setup Spanner emulator database using Go script
setup-emulator:
	@echo "Setting up Spanner emulator database..."
	SPANNER_EMULATOR_HOST=$(SPANNER_EMULATOR_HOST) $(GOCMD) run scripts/setup_emulator.go

# Run unit tests only (no Spanner required)
test-unit:
	$(GOTEST) -v ./internal/...

# Run all unit tests with coverage
test-coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out ./internal/...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
fmt:
	$(GOCMD) fmt ./...

# Vet code
vet:
	$(GOCMD) vet ./...

# Run linter
lint:
	golangci-lint run ./...

# Docker build
docker-build:
	docker build -t product-catalog-service:latest .

# Docker run with emulator
docker-run:
	docker-compose up -d

# Docker stop
docker-stop:
	docker-compose down

# Full CI pipeline locally
ci: deps fmt vet lint test-unit build
	@echo "CI pipeline completed successfully!"

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application binary"
	@echo "  test          - Run all tests (requires Spanner emulator)"
	@echo "  test-unit     - Run unit tests only (no Spanner required)"
	@echo "  test-e2e      - Run E2E tests (requires Spanner emulator)"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  run           - Run the application"
	@echo "  clean         - Remove build artifacts"
	@echo "  proto         - Generate protobuf code"
	@echo "  emulator-up   - Start Spanner emulator"
	@echo "  emulator-down - Stop Spanner emulator"
	@echo "  setup-emulator- Setup database schema"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo "  lint          - Run golangci-lint"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run with Docker Compose"
	@echo "  docker-stop   - Stop Docker Compose"
	@echo "  ci            - Run full CI pipeline locally"
