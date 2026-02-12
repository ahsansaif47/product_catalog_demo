# Product Catalog Service

A Go-based Product Catalog Service implementing Domain-Driven Design (DDD) and Clean Architecture principles with Google Cloud Spanner as the database.

## Architecture

This service follows a clean architecture pattern with clear separation of concerns:

- **Domain Layer**: Pure business logic with no external dependencies
- **Application Layer**: Use cases for commands and queries (CQRS)
- **Infrastructure Layer**: Database repositories, gRPC transport, and external integrations
- **Domain-Driven Design**: Aggregates, value objects, domain events, and repositories

## Technology Stack

- **Go 1.21+**
- **Google Cloud Spanner** (use emulator for local development)
- **gRPC** with Protocol Buffers
- **CommitPlan** for atomic transactions
- **Docker** for Spanner emulator

## Project Structure

```
product-catalog-service/
├── cmd/server/              # Application entry point
├── internal/
│   ├── app/product/         # Application layer
│   │   ├── domain/          # Domain entities and business logic
│   │   ├── usecases/        # Command handlers
│   │   ├── queries/         # Query handlers
│   │   ├── contracts/       # Repository interfaces
│   │   └── repo/            # Repository implementations
│   ├── models/              # Database models
│   ├── transport/grpc/      # gRPC handlers
│   ├── services/            # Dependency injection
│   └── pkg/                 # Shared packages
├── proto/                   # Protocol buffer definitions
├── migrations/              # Database schema
├── tests/e2e/              # End-to-end tests
└── docker-compose.yml      # Spanner emulator setup
```

## Getting Started

### Prerequisites

- Go 1.21+
- Docker and Docker Compose
- `protoc` compiler with Go plugins

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd product-catalog-service
```

2. Install dependencies:
```bash
make deps
```

3. Generate protobuf code:
```bash
make proto
```

### Running the Service

1. Start the Spanner emulator:
```bash
make spanner-up
```

2. Apply database migrations:
```bash
export SPANNER_EMULATOR_HOST=localhost:9010
gcloud config set auth/disable_credentials true
gcloud config set project test-project
gcloud spanner instances create test-instance --config=emulator-config --description="Test Instance"
gcloud spanner databases create product-catalog --instance=test-instance
gcloud spanner databases ddl update product-catalog --instance=test-instance --ddl="$(cat migrations/001_initial_schema.sql)"
```

3. Run the service:
```bash
make run
```

The service will start on port 50051 by default.

## Running Tests

### Run all tests:
```bash
make test
```

### Run E2E tests only:
```bash
make test-e2e
```

### Run tests with coverage:
```bash
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## API Endpoints

### Commands

| RPC | Description |
|-----|-------------|
| `CreateProduct` | Create a new product |
| `UpdateProduct` | Update product details |
| `ActivateProduct` | Activate a product |
| `DeactivateProduct` | Deactivate a product |
| `ApplyDiscount` | Apply a discount to a product |
| `RemoveDiscount` | Remove a discount from a product |
| `ArchiveProduct` | Archive a product (soft delete) |

### Queries

| RPC | Description |
|-----|-------------|
| `GetProduct` | Get a product by ID with effective price |
| `ListProducts` | List products with pagination and filtering |

## Key Features

### Domain-Driven Design
- Rich domain models with behavior
- Value objects (Money, Discount)
- Domain events for state changes
- Aggregate boundaries

### CQRS Pattern
- Separate command and query handlers
- Optimized read models
- Clear separation of concerns

### Transactional Outbox
- Reliable event publishing
- Atomic writes with events
- Decoupled event processing

### Precise Money Handling
- Uses `big.Rat` for decimal precision
- No floating-point arithmetic
- Exact discount calculations

## Development

### Build the binary:
```bash
make build
```

### Format code:
```bash
make fmt
```

### Run linter:
```bash
make lint
```

### Clean build artifacts:
```bash
make clean
```

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `50051` | gRPC server port |
| `SPANNER_DATABASE` | `projects/test-project/instances/test-instance/databases/product-catalog` | Spanner database path |
| `SPANNER_EMULATOR_HOST` | `localhost:9010` | Spanner emulator host |

## Design Decisions

### Why Domain-Driven Design?
- Captures business complexity in the domain layer
- Enables ubiquitous language with business stakeholders
- Isolates business logic from infrastructure concerns

### Why CQRS?
- Optimizes read and write operations separately
- Enables complex queries without compromising write logic
- Clear separation of commands (mutations) and queries (reads)

### Why Spanner?
- Horizontal scalability
- Strong consistency
- Global distribution
- Managed service with high availability

### Why CommitPlan?
- Atomic transactions across multiple tables
- Outbox pattern for reliable events
- Clean separation between mutation generation and application

## Testing

The test suite includes:
- Unit tests for domain logic
- Integration tests for repositories
- E2E tests for complete user flows

E2E tests verify:
- Product creation with outbox events
- Product updates with change tracking
- Discount application with price calculations
- State transitions (activate/deactivate/archive)
- Business rule validation errors
- Pagination and filtering

## License

This is a test project for demonstration purposes.

## Contributing

This is a test task repository. Please refer to the task description for implementation requirements.
