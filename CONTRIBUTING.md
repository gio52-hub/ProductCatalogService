# Contributing to Product Catalog Service

Thank you for your interest in contributing to this project!

## Development Setup

1. **Prerequisites**
   - Go 1.22+
   - Docker & Docker Compose
   - Protocol Buffer compiler (protoc)

2. **Clone and Setup**
   ```bash
   git clone <repository-url>
   cd product-catalog-service
   make deps
   ```

3. **Start Spanner Emulator**
   ```bash
   make emulator-up
   make setup-emulator
   ```

## Development Workflow

### Running Tests

```bash
# Unit tests only (fast, no dependencies)
make test-unit

# All tests with coverage
make test-coverage

# E2E tests (requires Spanner emulator)
make test-e2e
```

### Code Quality

```bash
# Format code
make fmt

# Run vet
make vet

# Run linter
make lint

# Full CI locally
make ci
```

### Building

```bash
# Build binary
make build

# Build Docker image
make docker-build
```

## Code Style Guidelines

1. **Follow Go conventions**
   - Use `gofmt` and `goimports`
   - Follow [Effective Go](https://golang.org/doc/effective_go)
   - Use meaningful variable and function names

2. **Domain-Driven Design**
   - Keep domain layer pure (no external dependencies)
   - Use value objects for primitive obsession
   - Emit domain events for state changes

3. **Testing**
   - Write table-driven tests
   - Use testify for assertions
   - Aim for high test coverage on domain logic

4. **Error Handling**
   - Use sentinel errors for domain violations
   - Wrap errors with context
   - Don't return internal details in API responses

## Pull Request Process

1. Create a feature branch from `main`
2. Write tests for new functionality
3. Ensure all tests pass locally
4. Run the full CI pipeline: `make ci`
5. Submit PR with clear description
6. Address review feedback

## Architecture Guidelines

- Follow Clean Architecture principles
- Maintain strict layer separation
- Use dependency injection
- Implement repository pattern for data access
