# Quickstart Guide: PIM Product API

**Feature**: 001-pim-product-api  
**Last Updated**: 2025-11-14

## Overview

This guide walks you through setting up the local development environment, running the database, executing migrations, running tests, and starting the API server.

## Prerequisites

- **Go**: Version 1.21 or later ([install](https://go.dev/doc/install))
- **Docker & Docker Compose**: For PostgreSQL test database ([install](https://docs.docker.com/get-docker/))
- **Make**: For build automation (usually pre-installed on macOS/Linux)
- **Git**: For version control

Verify installations:
```bash
go version       # Should show go1.21 or later
docker --version # Should show Docker version
make --version   # Should show GNU Make
```

## Step 1: Clone and Setup Project

```bash
# Navigate to project directory
cd /Users/sunfmin/Developments/apidemo1

# Initialize Go module (if not already done)
go mod init apidemo1

# Install dependencies
go get github.com/go-chi/chi/v5
go get github.com/jackc/pgx/v5
go get github.com/jmoiron/sqlx
go get github.com/golang-migrate/migrate/v4
go get github.com/golang-migrate/migrate/v4/database/postgres
go get github.com/golang-migrate/migrate/v4/source/file

# Tidy up dependencies
go mod tidy
```

## Step 2: Start PostgreSQL Test Database

Create `docker-compose.yml` in project root:

```yaml
version: '3.8'

services:
  postgres-test:
    image: postgres:15-alpine
    container_name: apidemo1-postgres-test
    environment:
      POSTGRES_DB: apidemo1_test
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
    ports:
      - "5433:5432"
    volumes:
      - postgres-test-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U test"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  postgres-test-data:
```

Start the database:
```bash
# Start PostgreSQL in detached mode
docker-compose up -d

# Verify database is running
docker-compose ps

# Check logs (optional)
docker-compose logs postgres-test
```

## Step 3: Configure Environment Variables

Create `.env` file in project root:

```bash
# Database URLs
DATABASE_URL=postgres://test:test@localhost:5433/apidemo1_test?sslmode=disable
TEST_DATABASE_URL=postgres://test:test@localhost:5433/apidemo1_test?sslmode=disable

# Server configuration
SERVER_PORT=8080
SERVER_HOST=localhost

# Log level (debug, info, warn, error)
LOG_LEVEL=debug
```

Load environment variables:
```bash
# Option 1: Export manually
export TEST_DATABASE_URL="postgres://test:test@localhost:5433/apidemo1_test?sslmode=disable"

# Option 2: Use direnv (recommended for automatic loading)
# Install direnv: https://direnv.net/
echo 'export TEST_DATABASE_URL="postgres://test:test@localhost:5433/apidemo1_test?sslmode=disable"' > .envrc
direnv allow
```

## Step 4: Create Database Migrations

Create migrations directory structure:
```bash
mkdir -p migrations
```

The implementation phase will create migration files like:
- `migrations/001_create_products.up.sql`
- `migrations/001_create_products.down.sql`
- etc.

Install migration CLI tool (optional, for manual migration runs):
```bash
# macOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# Verify installation
migrate -version
```

Run migrations manually (after implementation):
```bash
migrate -path migrations -database "$TEST_DATABASE_URL" up
```

## Step 5: Project Structure

Your project will have this structure after implementation:

```
/Users/sunfmin/Developments/apidemo1/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── models/                  # Domain entities
│   ├── handlers/                # HTTP handlers with tests
│   ├── repository/              # Database operations
│   ├── middleware/              # HTTP middleware
│   ├── validator/               # Validation logic
│   └── testutil/                # Test helpers
├── migrations/                  # Database migrations
├── specs/                       # Feature specifications
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── Makefile                     # Build commands
├── docker-compose.yml           # PostgreSQL setup
├── .env                         # Environment variables (git-ignored)
├── .gitignore                   # Git ignore rules
└── README.md                    # Project documentation
```

## Step 6: Create Makefile

Create `Makefile` in project root for common commands:

```makefile
.PHONY: help test test-verbose run migrate-up migrate-down migrate-create clean

# Default target
help:
	@echo "Available targets:"
	@echo "  make test          - Run all tests"
	@echo "  make test-verbose  - Run tests with verbose output"
	@echo "  make run           - Run the API server"
	@echo "  make migrate-up    - Run database migrations up"
	@echo "  make migrate-down  - Rollback last migration"
	@echo "  make migrate-create NAME=<name> - Create new migration"
	@echo "  make clean         - Clean build artifacts"

# Run all tests
test:
	@echo "Running integration tests..."
	TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test ./... -race

# Run tests with verbose output
test-verbose:
	@echo "Running integration tests (verbose)..."
	TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test ./... -v -race

# Run API server
run:
	@echo "Starting API server..."
	DATABASE_URL="$(DATABASE_URL)" go run cmd/api/main.go

# Run migrations up
migrate-up:
	@echo "Running migrations up..."
	migrate -path migrations -database "$(TEST_DATABASE_URL)" up

# Rollback last migration
migrate-down:
	@echo "Rolling back last migration..."
	migrate -path migrations -database "$(TEST_DATABASE_URL)" down 1

# Create new migration
migrate-create:
	@echo "Creating migration: $(NAME)"
	migrate create -ext sql -dir migrations -seq $(NAME)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	go clean
	rm -rf bin/
```

## Step 7: Running Tests

Once implementation is complete:

```bash
# Run all tests
make test

# Run with verbose output
make test-verbose

# Run specific package tests
go test ./internal/handlers -v

# Run specific test function
go test ./internal/handlers -v -run TestCreateProduct

# Run with coverage
go test ./... -cover

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Step 8: Running the API Server

Once implementation is complete:

```bash
# Start the server
make run

# Or run directly
go run cmd/api/main.go

# Server will start on http://localhost:8080
```

Test the API:
```bash
# Health check (example, will be implemented)
curl http://localhost:8080/health

# Create a product (example)
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Blue T-Shirt",
    "sku": "TSHIRT-BLUE-001",
    "description": "Comfortable cotton t-shirt",
    "attributes": {
      "color": {"type": "string", "value": "blue"},
      "size": {"type": "string", "value": "XL"}
    }
  }'
```

## Step 9: Development Workflow

### TDD Workflow (per Constitution)

1. **Write Tests First**:
   ```bash
   # Create test file (example)
   touch internal/handlers/products_test.go
   
   # Write table-driven integration tests covering:
   # - Happy path
   # - Input validation edge cases
   # - Boundary conditions
   # - Auth errors (if applicable)
   # - Data state errors (404, conflicts)
   # - Database errors
   # - HTTP specifics
   ```

2. **Run Tests (Red Phase)**:
   ```bash
   make test-verbose
   # Tests should FAIL (red phase)
   ```

3. **Implement Code (Green Phase)**:
   ```bash
   # Create implementation file
   touch internal/handlers/products.go
   
   # Implement minimal code to pass tests
   ```

4. **Run Tests Again (Green Phase)**:
   ```bash
   make test-verbose
   # Tests should PASS (green phase)
   ```

5. **Refactor**:
   ```bash
   # Improve code quality while keeping tests green
   make test
   ```

### Database Migrations

```bash
# Create new migration
make migrate-create NAME=add_new_field

# This creates:
# - migrations/00X_add_new_field.up.sql
# - migrations/00X_add_new_field.down.sql

# Run migrations
make migrate-up

# Rollback if needed
make migrate-down
```

### Running Specific Tests

```bash
# Test a single handler
go test ./internal/handlers -v -run TestCreateProduct

# Test with real database (default)
TEST_DATABASE_URL="postgres://test:test@localhost:5433/apidemo1_test?sslmode=disable" \
  go test ./internal/handlers -v

# Run with race detector (recommended)
go test ./... -race
```

## Step 10: Common Issues & Troubleshooting

### Issue: Database Connection Failed

```bash
# Check if PostgreSQL is running
docker-compose ps

# Check logs
docker-compose logs postgres-test

# Restart database
docker-compose restart postgres-test

# Or stop and start fresh
docker-compose down
docker-compose up -d
```

### Issue: Port Already in Use

```bash
# Check what's using port 5433
lsof -i :5433

# Kill the process or change port in docker-compose.yml
```

### Issue: Migrations Failed

```bash
# Check migration version
migrate -path migrations -database "$TEST_DATABASE_URL" version

# Force version (use with caution)
migrate -path migrations -database "$TEST_DATABASE_URL" force VERSION

# Rollback and retry
make migrate-down
make migrate-up
```

### Issue: Tests Failing

```bash
# Ensure database is clean
docker-compose down -v  # Remove volumes
docker-compose up -d    # Start fresh

# Run migrations
make migrate-up

# Run tests
make test-verbose
```

## Step 11: Development Tools (Recommended)

### VS Code Extensions
- Go (official Go extension)
- PostgreSQL (for database management)
- REST Client (for API testing)
- Docker (for container management)

### Database GUI Tools
- [pgAdmin](https://www.pgadmin.org/) - Full-featured PostgreSQL admin tool
- [DBeaver](https://dbeaver.io/) - Universal database tool
- [Postico](https://eggerapps.at/postico/) - macOS PostgreSQL client

### API Testing Tools
- [Postman](https://www.postman.com/) - Import OpenAPI spec for API testing
- [Insomnia](https://insomnia.rest/) - Alternative API client
- `curl` - Command-line HTTP client (built-in)

## Step 12: CI/CD Setup (Future)

For CI/CD pipelines, use GitHub Actions or similar:

```yaml
# .github/workflows/test.yml (example)
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_DB: apidemo1_test
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run tests
        env:
          TEST_DATABASE_URL: postgres://test:test@localhost:5432/apidemo1_test?sslmode=disable
        run: make test
```

## Next Steps

After completing this quickstart setup:

1. Wait for implementation phase (`/speckit.implement`)
2. Tests will be written first (per TDD constitution)
3. Review tests before implementation
4. Implement code to pass tests
5. Run full test suite
6. Deploy to staging/production

## Resources

- [Go Documentation](https://go.dev/doc/)
- [Chi Router](https://go-chi.io/)
- [pgx Driver](https://github.com/jackc/pgx)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Feature Spec](./spec.md)
- [Data Model](./data-model.md)
- [API Contracts](./contracts/openapi.yaml)

