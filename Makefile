.PHONY: help test test-verbose run migrate-up migrate-down migrate-create clean db-up db-down

# Default target
help:
	@echo "Available targets:"
	@echo "  make test          - Run all tests"
	@echo "  make test-verbose  - Run tests with verbose output"
	@echo "  make run           - Run the API server"
	@echo "  make migrate-up    - Run database migrations up"
	@echo "  make migrate-down  - Rollback last migration"
	@echo "  make migrate-create NAME=<name> - Create new migration"
	@echo "  make db-up         - Start PostgreSQL database"
	@echo "  make db-down       - Stop PostgreSQL database"
	@echo "  make clean         - Clean build artifacts"

# Database URL (override with environment variable if needed)
DATABASE_URL ?= postgres://test:test@localhost:5433/apidemo1_test?sslmode=disable
TEST_DATABASE_URL ?= postgres://test:test@localhost:5433/apidemo1_test?sslmode=disable

# Start database
db-up:
	@echo "Starting PostgreSQL database..."
	docker-compose up -d

# Stop database
db-down:
	@echo "Stopping PostgreSQL database..."
	docker-compose down

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
	migrate -path migrations -database "$(DATABASE_URL)" up

# Rollback last migration
migrate-down:
	@echo "Rolling back last migration..."
	migrate -path migrations -database "$(DATABASE_URL)" down 1

# Create new migration
migrate-create:
	@echo "Creating migration: $(NAME)"
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=my_migration"; \
		exit 1; \
	fi
	migrate create -ext sql -dir migrations -seq $(NAME)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	go clean
	rm -rf bin/

