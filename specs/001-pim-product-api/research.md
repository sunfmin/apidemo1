# Research: PIM Product API

**Feature**: 001-pim-product-api  
**Date**: 2025-11-14  
**Status**: Complete

## Overview

This document resolves technical uncertainties identified in the Technical Context section of the implementation plan. Primary research areas include HTTP routing framework selection, database library selection, and migration tooling.

## Research Tasks

### 1. HTTP Router Selection

**Question**: Should we use standard library `net/http` ServeMux, or a third-party router (Chi, Echo, Gin)?

**Decision**: **Chi router** (github.com/go-chi/chi/v5)

**Rationale**:
- **Idiomatic Go**: Chi is built on standard library `net/http`, uses `http.HandlerFunc` interface
- **Constitution compliant**: Easy to test via `httptest.ResponseRecorder` (no special test helpers needed)
- **Middleware composability**: Native middleware stack compatible with standard library
- **Routing features**: URL parameters, route groups, sub-routers for clean organization
- **Lightweight**: Minimal dependencies, fast routing, low memory footprint
- **Active maintenance**: Well-maintained, widely adopted in Go community

**Alternatives Considered**:
- **net/http ServeMux**: Too basic for REST APIs, lacks URL parameter extraction, verbose route definitions
- **Echo**: Good features but uses custom context (`echo.Context`) that complicates testing, less idiomatic
- **Gin**: Fast but uses reflection-heavy binding, custom context, more magic than explicit code

**Best Practices**:
- Use Chi's `chi.URLParam(r, "id")` for extracting path parameters
- Organize routes in route groups: `/api/v1/products`, `/api/v1/variants`
- Use Chi's middleware stack: `chi.Logger`, `chi.Recoverer`, custom CORS
- Mount sub-routers for logical separation: `r.Mount("/products", productsRouter())`

---

### 2. Database Library Selection

**Question**: Should we use `database/sql` with `pgx` driver, `sqlx`, or an ORM like `GORM`?

**Decision**: **pgx/v5** with **sqlx** for query building

**Rationale**:
- **pgx/v5**: Native PostgreSQL driver, better performance than `lib/pq`, excellent connection pooling
- **sqlx**: Lightweight layer over `database/sql`, adds scanning to structs, named queries, no ORM overhead
- **Constitution compliant**: Direct SQL queries in tests (real database operations, no mocking)
- **Flexibility**: Full control over SQL for performance tuning, explicit queries for clarity
- **JSONB support**: pgx has excellent support for PostgreSQL JSONB (critical for flexible attributes)
- **Transaction handling**: Explicit transaction control for testing (rollback after each test)

**Alternatives Considered**:
- **GORM**: Full ORM with migrations, associations, hooks. Too much magic, hides SQL, harder to debug performance issues, migration conflicts with separate migration tools
- **database/sql only**: Too verbose, requires manual row scanning, no struct mapping
- **sqlc**: Code generation from SQL. Good option but adds build step complexity, less dynamic for flexible attributes

**Best Practices**:
- Use `pgx.Pool` for connection pooling (configure in `main.go`)
- Use `sqlx.DB.Get()` and `sqlx.DB.Select()` for single row and multiple row queries
- Use `sqlx.Named()` for named parameter queries (clearer than positional `$1, $2`)
- Store flexible attributes in JSONB columns: `attributes JSONB NOT NULL DEFAULT '{}'`
- Use PostgreSQL functions for JSONB queries: `jsonb_build_object`, `jsonb_agg`
- Explicit transaction handling in tests: `tx, _ := db.Beginx(); defer tx.Rollback()`

---

### 3. Database Migration Tool

**Question**: Should we use `golang-migrate`, `goose`, or embed migrations in Go code?

**Decision**: **golang-migrate/migrate** (github.com/golang-migrate/migrate/v4)

**Rationale**:
- **Industry standard**: Most widely used migration tool in Go ecosystem
- **Multiple sources**: SQL files, embedded Go files, remote sources
- **CLI tool**: Easy to run migrations from command line: `migrate -path migrations -database postgres://... up`
- **Go library**: Can also run migrations programmatically in `main.go` or tests
- **Rollback support**: Down migrations for reverting changes
- **Constitution compliant**: Test migrations run before each test suite to ensure schema matches production

**Alternatives Considered**:
- **goose**: Good alternative, but less community adoption, less documentation
- **Embedded migrations in Go**: Hard to version, mixing concerns, no standalone CLI tool

**Best Practices**:
- Store migrations in `migrations/` directory at repo root
- Name files: `001_create_products.up.sql`, `001_create_products.down.sql`
- Each migration file handles one logical schema change
- Run migrations in test setup: `testutil.RunMigrations(db)`
- Use PostgreSQL-specific features: JSONB, UUID, indexes, constraints

---

### 4. Flexible Attribute Storage Strategy

**Question**: How should we store flexible attributes that vary by product type?

**Decision**: **JSONB column** in `products` and `variants` tables

**Rationale**:
- **Schema flexibility**: No ALTER TABLE needed when adding new attribute types
- **Type safety**: JSONB validates JSON structure, supports indexing
- **Query capability**: PostgreSQL JSONB operators for filtering: `@>`, `?`, `->`
- **Performance**: GIN indexes on JSONB columns for fast lookups
- **Simplicity**: Single column vs. EAV (Entity-Attribute-Value) pattern with separate tables

**Schema Example**:
```sql
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    sku TEXT UNIQUE NOT NULL,
    description TEXT,
    product_type_id UUID REFERENCES product_types(id) ON DELETE SET NULL,
    attributes JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_attributes ON products USING GIN (attributes);
```

**Attribute Storage Format**:
```json
{
  "color": {"type": "string", "value": "blue"},
  "size": {"type": "string", "value": "XL"},
  "weight": {"type": "number", "value": 2.5},
  "in_stock": {"type": "boolean", "value": true},
  "release_date": {"type": "date", "value": "2025-01-15"}
}
```

**Alternatives Considered**:
- **EAV pattern** (separate `attributes` table): Requires JOINs, complex queries, slower performance, harder to maintain
- **Fixed columns**: Requires migrations for new attributes, not flexible
- **NoSQL database**: Overkill for this use case, loses PostgreSQL transaction guarantees

**Best Practices**:
- Store type metadata with each attribute for validation
- Use GIN indexes for attribute queries: `WHERE attributes @> '{"color": {"type": "string", "value": "blue"}}'`
- Validate attribute types in application layer before persisting
- Use PostgreSQL's `jsonb_set()` for updating individual attributes without full replacement

---

### 5. Test Database Strategy

**Question**: How should integration tests manage the test database?

**Decision**: **Docker Compose PostgreSQL + Transaction Rollback per test**

**Rationale**:
- **Isolation**: Each test runs in a transaction that's rolled back, ensuring clean state
- **Speed**: Transaction rollback is faster than recreating database or truncating tables
- **Consistency**: Tests run against same schema as production (via migrations)
- **Local development**: Docker Compose makes setup easy for developers
- **CI/CD**: Same approach works in CI with ephemeral containers

**Implementation**:
```go
// testutil/database.go
func SetupTestDB(t *testing.T) *sqlx.DB {
    db, _ := sqlx.Connect("pgx", os.Getenv("TEST_DATABASE_URL"))
    // Run migrations
    RunMigrations(db)
    return db
}

func BeginTestTransaction(t *testing.T, db *sqlx.DB) *sqlx.Tx {
    tx, _ := db.Beginx()
    t.Cleanup(func() { tx.Rollback() })
    return tx
}
```

**Docker Compose Example**:
```yaml
version: '3.8'
services:
  postgres-test:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: apidemo1_test
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
    ports:
      - "5433:5432"
```

**Alternatives Considered**:
- **In-memory SQLite**: Doesn't support PostgreSQL-specific features (JSONB, UUID)
- **Truncate tables**: Slower than rollback, risks leaving data if test panics
- **Recreate database**: Very slow, not practical for hundreds of tests

**Best Practices**:
- Use environment variable for test database URL: `TEST_DATABASE_URL`
- Run migrations once per test suite startup, not per test
- Use transactions for test isolation, rollback in cleanup
- Provide fixture helpers that work within transactions: `CreateTestProduct(tx *sqlx.Tx)`

---

## Summary of Decisions

| Component | Decision | Rationale |
|-----------|----------|-----------|
| HTTP Router | Chi (go-chi/chi/v5) | Idiomatic Go, easy testing, middleware composability |
| Database Driver | pgx/v5 | Native PostgreSQL, excellent performance, JSONB support |
| Query Builder | sqlx | Lightweight, struct scanning, no ORM overhead |
| Migrations | golang-migrate | Industry standard, CLI tool, programmatic usage |
| Flexible Attributes | JSONB columns | Schema flexibility, query capability, performance |
| Test Database | Docker + Transaction Rollback | Fast isolation, consistent schema, easy local dev |

## Dependencies to Add

```
go get github.com/go-chi/chi/v5
go get github.com/jackc/pgx/v5
go get github.com/jmoiron/sqlx
go get github.com/golang-migrate/migrate/v4
go get github.com/golang-migrate/migrate/v4/database/postgres
go get github.com/golang-migrate/migrate/v4/source/file
```

## Next Steps

With research complete, proceed to Phase 1:
1. Create `data-model.md` with detailed entity schemas
2. Generate API contracts in `contracts/` directory (OpenAPI spec)
3. Create `quickstart.md` for local development setup

