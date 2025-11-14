# Implementation Plan: PIM Product API

**Branch**: `001-pim-product-api` | **Date**: 2025-11-14 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-pim-product-api/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Build a RESTful API for a Product Information Management (PIM) system that supports flexible product attributes, variants, media assets, and product type templates. The system enables merchandising managers to manage diverse product catalogs with varying attribute schemas without requiring database schema changes. Core capabilities include creating products with custom attributes (string, number, boolean, date types), adding variants for product configurations, attaching media assets (images/videos), defining reusable product type templates with validation, and managing multiple product types within a single system. All endpoints will be tested through integration tests using real PostgreSQL database and comprehensive edge case coverage per constitution requirements.

## Technical Context

**Language/Version**: Go 1.21+ (latest stable recommended)  
**Primary Dependencies**: Chi router (go-chi/chi/v5), pgx/v5 driver, sqlx query builder, golang-migrate for migrations  
**Storage**: PostgreSQL 15+ (with JSONB columns for flexible attributes, GIN indexes for attribute queries)  
**Testing**: Go testing package with httptest, table-driven integration tests, Docker PostgreSQL test database with transaction rollback  
**Target Platform**: Linux server (containerized deployment recommended)  
**Project Type**: Single API project (RESTful backend only, standard Go structure with cmd/ and internal/)  
**Performance Goals**: Product creation <5s for 10 attributes (SC-001), product retrieval <2s for 20 variants + 50 assets (SC-003), 1000 products created without degradation (SC-005), cascade delete <5s for 100 variants (SC-007)  
**Constraints**: Single-tenant MVP, media stored externally (URLs only), no full-text search, pagination default 50 items, UTC timezone  
**Scale/Scope**: Support 50+ product types (SC-009), 1000+ variants per product, 100+ media assets per product, concurrent updates with 99.9% success rate (SC-010)

**Research**: See [research.md](./research.md) for technology selection rationale and best practices.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- ✅ **Integration Testing First**: All tests use real PostgreSQL database (no mocking)
- ✅ **Table-Driven Tests**: All tests follow table-driven pattern with test case structs
- ✅ **Edge Case Coverage**: Tests include input validation, boundary conditions, auth errors, data state, database errors, HTTP specifics
- ✅ **Real Database Fixtures**: Test data prepared via real database operations
- ✅ **ServeHTTP Testing**: Endpoints tested through httptest.ResponseRecorder and actual HTTP handlers

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
/Users/sunfmin/Developments/apidemo1/
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
├── internal/
│   ├── models/                     # Domain entities
│   │   ├── product.go              # Product entity
│   │   ├── variant.go              # Variant entity
│   │   ├── attribute.go            # Attribute entity
│   │   ├── media_asset.go          # MediaAsset entity
│   │   └── product_type.go         # ProductTypeTemplate entity
│   ├── handlers/                   # HTTP handlers (ServeHTTP implementations)
│   │   ├── products.go             # Product CRUD handlers
│   │   ├── products_test.go        # Integration tests for products
│   │   ├── variants.go             # Variant CRUD handlers
│   │   ├── variants_test.go        # Integration tests for variants
│   │   ├── media_assets.go         # Media asset handlers
│   │   ├── media_assets_test.go    # Integration tests for media
│   │   ├── product_types.go        # Product type template handlers
│   │   └── product_types_test.go   # Integration tests for templates
│   ├── repository/                 # Database operations
│   │   ├── products.go             # Product repository
│   │   ├── variants.go             # Variant repository
│   │   ├── media_assets.go         # Media asset repository
│   │   └── product_types.go        # Product type repository
│   ├── middleware/                 # HTTP middleware
│   │   ├── logger.go               # Request logging
│   │   ├── recovery.go             # Panic recovery
│   │   └── cors.go                 # CORS headers
│   ├── validator/                  # Validation logic
│   │   ├── product.go              # Product validation
│   │   ├── attribute.go            # Attribute type validation
│   │   └── template.go             # Template validation
│   └── testutil/                   # Test helpers
│       ├── database.go             # Test DB setup/teardown
│       ├── fixtures.go             # Fixture creation helpers
│       └── http.go                 # HTTP test helpers
├── migrations/                     # Database migrations
│   ├── 001_create_products.up.sql
│   ├── 001_create_products.down.sql
│   ├── 002_create_variants.up.sql
│   ├── 002_create_variants.down.sql
│   ├── 003_create_attributes.up.sql
│   ├── 003_create_attributes.down.sql
│   ├── 004_create_media_assets.up.sql
│   ├── 004_create_media_assets.down.sql
│   ├── 005_create_product_types.up.sql
│   └── 005_create_product_types.down.sql
├── go.mod                          # Go module definition
├── go.sum                          # Go dependency checksums
├── Makefile                        # Build and test commands
├── docker-compose.yml              # PostgreSQL test database
└── README.md                       # Project documentation
```

**Structure Decision**: Standard Go project structure with `cmd/` for entry points and `internal/` for application code. Tests are co-located with implementation files using `*_test.go` suffix per Go conventions. All integration tests use real PostgreSQL database via `testutil` helpers. No separate `tests/` directory needed - Go's testing conventions place tests alongside source code.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations. All constitution principles are satisfied:
- ✅ Integration tests with real PostgreSQL (no mocking)
- ✅ Table-driven test structure
- ✅ Comprehensive edge case coverage planned
- ✅ Real database fixtures approach defined
- ✅ ServeHTTP testing via httptest

---

## Planning Phase Status

### Phase 0: Research (✅ Complete)

**Output**: [research.md](./research.md)

**Decisions Made**:
- HTTP Router: Chi (go-chi/chi/v5)
- Database Driver: pgx/v5 with sqlx
- Migration Tool: golang-migrate
- Flexible Attributes: JSONB columns with GIN indexes
- Test Database: Docker Compose with transaction rollback

**Rationale**: See research.md for detailed analysis and alternatives considered.

---

### Phase 1: Design & Contracts (✅ Complete)

**Outputs**:
- [data-model.md](./data-model.md) - Entity schemas, relationships, constraints
- [contracts/openapi.yaml](./contracts/openapi.yaml) - Complete REST API specification
- [quickstart.md](./quickstart.md) - Local development setup guide

**Entities Defined** (5):
1. Products - Parent products with flexible JSONB attributes
2. Variants - Product configurations with own SKUs
3. Media Assets - Images/videos for products and variants
4. Product Types - Reusable templates with validation
5. Attribute Definitions - Schema definitions for product types

**API Endpoints Defined** (20):
- Products: POST, GET (list), GET (by ID), PUT, DELETE
- Variants: POST, GET (list), GET (by ID), PUT, DELETE
- Media Assets: POST (product), GET (product), POST (variant), GET (variant), PUT, DELETE
- Product Types: POST, GET (list), GET (by ID), PUT, DELETE

**Agent Context**: ✅ Updated (CLAUDE.md)

---

### Next Phase: Implementation (Ready)

Run `/speckit.implement` to generate task breakdown and begin TDD implementation.

**Test-First Approach** (per constitution):
1. Write table-driven integration tests with comprehensive edge cases
2. Verify tests fail (red phase)
3. Review tests with team
4. Implement code to pass tests (green phase)
5. Refactor while keeping tests green

**Expected Test Coverage**:
- Happy path scenarios from acceptance criteria
- Input validation (empty, null, SQL injection, XSS)
- Boundary conditions (zero/max values, limits)
- Auth/authorization errors (if implemented)
- Data state errors (404, conflicts, concurrent updates)
- Database errors (constraints, foreign keys, transactions)
- HTTP specifics (methods, headers, content-types)
