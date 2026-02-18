# Product Catalog Service

A simplified Product Catalog Service built with **Go 1.22+**, implementing **Domain-Driven Design (DDD)** and **Clean Architecture** principles.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Transport Layer                                 │
│                         (gRPC Handlers, Validators)                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                            Application Layer                                 │
│                    (Use Cases / Queries / DTOs)                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                              Domain Layer                                    │
│          (Aggregates, Value Objects, Domain Events, Domain Services)        │
├─────────────────────────────────────────────────────────────────────────────┤
│                           Infrastructure Layer                               │
│              (Repositories, Database Models, External Services)             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Key Design Patterns

- **CQRS (Command Query Responsibility Segregation)**: Commands mutate state via use cases; queries read via read models
- **Golden Mutation Pattern**: Atomic writes using CommitPlan for transaction management
- **Repository Pattern**: Repositories return Spanner mutations, not applying them directly
- **Transactional Outbox**: Reliable event publishing via database-stored events
- **Change Tracking**: Optimized updates by tracking modified fields only

## Features

- **Product Management**: Create, Update, Activate, Deactivate, Archive products
- **Pricing Rules**: Percentage-based discounts with date ranges, precise decimal arithmetic using `math/big`
- **Product Queries**: Get by ID, List with pagination and filters
- **Event Publishing**: Domain events stored in transactional outbox

## Technology Stack

| Technology | Purpose |
|------------|---------|
| Go 1.22+ | Programming language |
| Google Cloud Spanner | Database (with emulator for local dev) |
| gRPC | Transport layer (Protocol Buffers) |
| math/big.Rat | Precise decimal arithmetic |
| testify | Testing framework |
| Docker & Docker Compose | Containerization |
| GitHub Actions | CI/CD |
| golangci-lint | Code quality |

## Project Structure

```
├── .github/workflows/             # GitHub Actions CI/CD pipeline
├── cmd/server/                    # Application entry point
├── internal/
│   ├── clock/                     # Time abstraction for testing
│   ├── committer/                 # Transaction commit plan
│   ├── contract/                  # Repository & read model interfaces
│   ├── domain/                    # Domain layer (pure Go, no dependencies)
│   ├── handler/                   # gRPC handlers, validators, mappers
│   ├── query/                     # Query handlers (CQRS read side)
│   ├── repository/                # Spanner implementations + DB models
│   └── usecase/                   # Command handlers (CQRS write side)
├── migrations/
│   └── 001_initial_schema.sql     # Database schema
├── proto/
│   └── product/
│       └── v1/                    # Protocol Buffer definitions
├── scripts/
│   └── setup_emulator.go          # Spanner emulator setup script
├── test/                          # End-to-end tests
├── .dockerignore                  # Docker build exclusions
├── .editorconfig                  # Editor configuration
├── .golangci.yml                  # Linter configuration
├── buf.yaml                       # Protobuf linting configuration
├── docker-compose.yml             # Local development setup
├── Dockerfile                     # Container build configuration
├── go.mod                         # Go module definition
├── go.sum                         # Dependency checksums
└── Makefile                       # Build automation
```

## Getting Started

### Prerequisites

- Go 1.22 or later
- Docker and Docker Compose
- Protocol Buffer compiler (protoc) - for regenerating proto files
- Make (optional, for using Makefile commands)

### Quick Start with Docker

The easiest way to run the service is using Docker Compose:

```bash
# Start both Spanner emulator and the service
docker-compose up -d

# Check logs
docker-compose logs -f product-catalog-service

# Stop everything
docker-compose down
```

The service will be available at `localhost:50051`.

### Local Development Setup

1. **Start Spanner Emulator**
```bash
docker-compose up -d spanner-emulator
```

2. **Setup Database**
```bash
# Set emulator host
$env:SPANNER_EMULATOR_HOST = "localhost:9010"  # PowerShell
export SPANNER_EMULATOR_HOST=localhost:9010     # Bash

# Create instance and database
go run scripts/setup_emulator.go
```

3. **Run the Service**
```bash
go run cmd/server/main.go
```

4. **Run Tests**
```bash
# Unit tests
go test -v ./internal/...

# E2E tests (requires Spanner emulator)
go test -v ./tests/e2e/...

# All tests with coverage
go test -v -race -coverprofile=coverage.out ./...
```

### Using Makefile

```bash
make build           # Build the binary
make test-unit       # Run unit tests
make test-e2e        # Run E2E tests
make emulator-up     # Start Spanner emulator
make emulator-down   # Stop Spanner emulator
make setup-emulator  # Setup database schema
make proto           # Regenerate protobuf files
```

## API Reference

### gRPC Endpoints

| Method | Description |
|--------|-------------|
| `CreateProduct` | Create a new product |
| `UpdateProduct` | Update product details |
| `ActivateProduct` | Activate a product |
| `DeactivateProduct` | Deactivate a product |
| `ArchiveProduct` | Archive (soft delete) a product |
| `ApplyDiscount` | Apply percentage discount |
| `RemoveDiscount` | Remove active discount |
| `GetProduct` | Get product by ID |
| `ListProducts` | List products with filters |

### Example gRPC Calls (using grpcurl)

```bash
# Create a product
grpcurl -plaintext -d '{
  "name": "Premium Widget",
  "description": "High-quality widget",
  "category": "Electronics",
  "base_price": {"numerator": 9999, "denominator": 100}
}' localhost:50051 product.v1.ProductService/CreateProduct

# Get product by ID
grpcurl -plaintext -d '{"product_id": "<UUID>"}' \
  localhost:50051 product.v1.ProductService/GetProduct

# List products with filter
grpcurl -plaintext -d '{"category": "Electronics", "page_size": 10}' \
  localhost:50051 product.v1.ProductService/ListProducts

# Apply discount
grpcurl -plaintext -d '{
  "product_id": "<UUID>",
  "discount_percentage": 15.5,
  "start_date": "2025-01-01T00:00:00Z",
  "end_date": "2025-12-31T23:59:59Z"
}' localhost:50051 product.v1.ProductService/ApplyDiscount
```

## Domain Model

### Product Aggregate

The `Product` is the core aggregate root with the following states:

```
┌─────────┐    Activate    ┌────────┐
│ DRAFT   │ ─────────────► │ ACTIVE │
└─────────┘                └────────┘
     │                          │
     │                          │ Deactivate
     │                          ▼
     │                    ┌──────────┐
     │                    │ INACTIVE │
     │                    └──────────┘
     │                          │
     │      Archive             │ Archive
     ├──────────────────────────┤
     │                          │
     ▼                          ▼
┌──────────┐              ┌──────────┐
│ ARCHIVED │              │ ARCHIVED │
└──────────┘              └──────────┘
```

### Value Objects

- **Money**: Precise decimal representation using `math/big.Rat`
- **Discount**: Percentage discount with validity dates

### Domain Events

| Event | Trigger |
|-------|---------|
| `ProductCreated` | Product creation |
| `ProductUpdated` | Product details update |
| `ProductActivated` | Product activation |
| `ProductDeactivated` | Product deactivation |
| `ProductArchived` | Product archival |
| `DiscountApplied` | Discount application |
| `DiscountRemoved` | Discount removal |

## Database Schema

```sql
CREATE TABLE products (
    product_id STRING(36) NOT NULL,
    name STRING(255) NOT NULL,
    description STRING(MAX),
    category STRING(100) NOT NULL,
    base_price_numerator INT64 NOT NULL,
    base_price_denominator INT64 NOT NULL,
    discount_percent NUMERIC,
    discount_start_date TIMESTAMP,
    discount_end_date TIMESTAMP,
    status STRING(20) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    archived_at TIMESTAMP
) PRIMARY KEY (product_id);

CREATE TABLE outbox_events (
    event_id STRING(36) NOT NULL,
    event_type STRING(100) NOT NULL,
    aggregate_id STRING(36) NOT NULL,
    payload JSON,
    status STRING(20) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    processed_at TIMESTAMP
) PRIMARY KEY (event_id);
```

## Testing Strategy

| Test Type | Location | Purpose |
|-----------|----------|---------|
| Unit Tests | `internal/**/*_test.go` | Domain logic, validators, mappers |
| E2E Tests | `tests/e2e/` | Full flow with Spanner emulator |

Tests follow table-driven patterns for comprehensive coverage.

## CI/CD Pipeline

GitHub Actions workflow includes:

1. **Lint**: Code quality check with golangci-lint
2. **Test**: Unit tests with race detection and coverage
3. **Build**: Binary compilation
4. **Docker**: Container image build
5. **E2E Tests**: Integration tests with Spanner emulator

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `GRPC_PORT` | `50051` | gRPC server port |
| `SPANNER_PROJECT_ID` | `test-project` | GCP project ID |
| `SPANNER_INSTANCE_ID` | `test-instance` | Spanner instance ID |
| `SPANNER_DATABASE_ID` | `test-database` | Spanner database ID |
| `SPANNER_EMULATOR_HOST` | - | Emulator host (for local dev) |

## License

MIT License
