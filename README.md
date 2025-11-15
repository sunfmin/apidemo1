# apidemo1

A Go backend API service with PostgreSQL database following integration-first testing principles.

## Project Constitution

This project follows strict development principles documented in [`.specify/memory/constitution.md`](.specify/memory/constitution.md).

### Core Principles

1. **Integration Testing First (No Mocking)** - All tests use real PostgreSQL database
2. **Table-Driven Test Design** - All tests follow table-driven patterns
3. **Edge Case Coverage (NON-NEGOTIABLE)** - Comprehensive edge case testing required
4. **Real Database Fixtures** - Test data prepared via actual database operations
5. **ServeHTTP Endpoint Testing** - API endpoints tested through full HTTP stack
6. **Protobuf Data Structures** - All public API types defined in .proto files, no maps in tests

## Technology Stack

- **Language**: Go 1.21+
- **Database**: PostgreSQL 15+
- **HTTP**: Standard library `net/http` or compatible framework
- **Protocol Buffers**: protoc compiler with Go plugins for type-safe API contracts
- **Testing**: Go `testing` package with `httptest`, protobuf-generated structs, and `protocmp` for assertions
- **Database Driver**: TBD (pgx, database/sql, or ORM)

## Development Workflow

### Test-First Development (TDD)

1. Design API contract (endpoints, request/response schemas)
2. Write table-driven integration tests covering happy path + edge cases
3. Verify tests fail (red phase)
4. Review tests with team before implementation
5. Implement code to make tests pass (green phase)
6. Refactor while keeping tests green

### Running Tests

```bash
# Run all integration tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package tests
go test -v ./[package-name]
```

### Test Database Setup

Tests require a PostgreSQL database. Configure via environment variables:

```bash
# Example test database configuration
export TEST_DATABASE_URL="postgres://user:password@localhost:5432/apidemo1_test?sslmode=disable"
```

## Project Structure

TBD - Will be defined in feature specification and implementation plan.

## Getting Started

TBD - Will be documented as features are implemented.

## Contributing

All contributions must comply with the project constitution. Key requirements:

- ✅ Integration tests required (no mocking)
- ✅ Table-driven test structure
- ✅ Comprehensive edge case coverage
- ✅ Real database fixtures
- ✅ Tests through ServeHTTP interface
- ✅ Protobuf structs for all API types (no maps)

See [`.specify/memory/constitution.md`](.specify/memory/constitution.md) for complete principles.

## License

TBD

