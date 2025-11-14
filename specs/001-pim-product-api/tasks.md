# Tasks: PIM Product API

**Input**: Design documents from `/specs/001-pim-product-api/`
**Prerequisites**: plan.md (✓), spec.md (✓), research.md (✓), data-model.md (✓), contracts/ (✓)

**Tests**: Integration tests are MANDATORY per constitution. All tests use real PostgreSQL database (no mocking), follow table-driven patterns, and cover comprehensive edge cases.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Go project**: Root level for `main.go`, packages in subdirectories, `*_test.go` files alongside source
- **Test organization**: Integration tests in `*_test.go` files (no separate `tests/` directory per Go convention)
- **Test database**: Use real PostgreSQL, configure via environment variables
- Paths shown below use absolute paths from repository root

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Initialize Go module with `go mod init apidemo1`
- [x] T002 [P] Install Chi router dependency: `go get github.com/go-chi/chi/v5`
- [x] T003 [P] Install pgx driver dependency: `go get github.com/jackc/pgx/v5`
- [x] T004 [P] Install sqlx dependency: `go get github.com/jmoiron/sqlx`
- [x] T005 [P] Install golang-migrate dependencies: `go get github.com/golang-migrate/migrate/v4`
- [x] T006 Create project directory structure (cmd/api, internal/{models,handlers,repository,middleware,validator,testutil}, migrations)
- [x] T007 Create `docker-compose.yml` for PostgreSQL test database at repository root
- [x] T008 Create `Makefile` with test, run, migrate-up, migrate-down, migrate-create commands at repository root
- [x] T009 [P] Create `.gitignore` for Go projects at repository root
- [x] T010 [P] Create `.env.example` with DATABASE_URL and TEST_DATABASE_URL templates at repository root

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T011 Create database migration `001_create_products.up.sql` in migrations/ (products table with JSONB attributes)
- [x] T012 Create database migration `001_create_products.down.sql` in migrations/ (drop products table)
- [x] T013 Create database migration `002_create_variants.up.sql` in migrations/ (variants table with product_id FK)
- [x] T014 Create database migration `002_create_variants.down.sql` in migrations/ (drop variants table)
- [x] T015 Create database migration `003_create_media_assets.up.sql` in migrations/ (media_assets table with CHECK constraint)
- [x] T016 Create database migration `003_create_media_assets.down.sql` in migrations/ (drop media_assets table)
- [x] T017 Create database migration `004_create_product_types.up.sql` in migrations/ (product_types table)
- [x] T018 Create database migration `004_create_product_types.down.sql` in migrations/ (drop product_types table)
- [x] T019 Create database migration `005_create_attribute_definitions.up.sql` in migrations/ (attribute_definitions table)
- [x] T020 Create database migration `005_create_attribute_definitions.down.sql` in migrations/ (drop attribute_definitions table)
- [x] T021 Create database migration `006_create_triggers.up.sql` in migrations/ (updated_at trigger for all tables)
- [x] T022 Create database migration `006_create_triggers.down.sql` in migrations/ (drop triggers)
- [x] T023 [P] Create test database helper in internal/testutil/database.go (SetupTestDB, RunMigrations, BeginTestTransaction)
- [x] T024 [P] Create fixture helper utilities in internal/testutil/fixtures.go (CreateTestProduct, CreateTestVariant, CreateTestMediaAsset, CreateTestProductType)
- [x] T025 [P] Create HTTP test helpers in internal/testutil/http.go (MakeRequest, ParseJSONResponse, AssertStatus, AssertJSONEqual)
- [x] T026 [P] Implement logging middleware in internal/middleware/logger.go
- [x] T027 [P] Implement panic recovery middleware in internal/middleware/recovery.go
- [x] T028 [P] Implement CORS middleware in internal/middleware/cors.go
- [x] T029 Create error response types in internal/models/errors.go (ErrorResponse struct, error code constants)
- [x] T030 Create base HTTP router setup in cmd/api/main.go (Chi router, middleware stack, database connection pool)

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Create Basic Products with Flexible Attributes (Priority: P1) 🎯 MVP

**Goal**: Enable creating, reading, updating, and deleting products with flexible JSONB attributes

**Independent Test**: Create product via POST with custom attributes, retrieve via GET to verify all attributes stored, update attributes via PUT, delete via DELETE. System works without variants, media, or templates.

### Integration Tests for User Story 1 (MANDATORY) ⚠️

> **CRITICAL: Write these tests FIRST, ensure they FAIL before implementation**
> **All tests MUST use real PostgreSQL, table-driven pattern, and cover edge cases**

- [ ] T031 [US1] Integration test for POST /api/v1/products in internal/handlers/products_test.go
  - Happy path: Create product with name, SKU, description, and various attribute types (string, number, boolean, date)
  - Edge case - Input validation: Empty name, empty SKU, nil attributes, invalid JSON
  - Edge case - Input validation: SQL injection in name/description, XSS payloads in text fields
  - Edge case - Boundary conditions: Name at min (1 char) and max (500 chars), description at max (10000 chars)
  - Edge case - Boundary conditions: Zero attributes, 100+ attributes, oversized attribute values
  - Edge case - Database errors: Duplicate SKU (409 Conflict), invalid UUID for product_type_id
  - Edge case - HTTP specifics: Missing Content-Type header, invalid JSON syntax, wrong HTTP method (GET when POST expected)
  - Use httptest.ResponseRecorder and real database fixtures
  - Table-driven test structure with test case structs

- [ ] T032 [US1] Integration test for GET /api/v1/products/{id} in internal/handlers/products_test.go
  - Happy path: Retrieve existing product with all attributes
  - Edge case - Data state: Non-existent product ID (404 Not Found)
  - Edge case - Input validation: Invalid UUID format (400 Bad Request)
  - Edge case - HTTP specifics: Wrong HTTP method (POST when GET expected)
  - Table-driven test with multiple product fixtures

- [ ] T033 [US1] Integration test for GET /api/v1/products (list with pagination) in internal/handlers/products_test.go
  - Happy path: List products with default pagination (page=1, page_size=50)
  - Edge case - Boundary conditions: Empty product list, single product, 100+ products
  - Edge case - Input validation: Invalid page number (0, negative), invalid page_size (0, negative, >100)
  - Edge case - Query parameters: Filter by product_type_id (valid, invalid, non-existent)
  - Table-driven test with various pagination scenarios

- [ ] T034 [US1] Integration test for PUT /api/v1/products/{id} in internal/handlers/products_test.go
  - Happy path: Update product name, description, and attributes (add new, modify existing, remove)
  - Edge case - Data state: Update non-existent product (404 Not Found)
  - Edge case - Input validation: Empty name after update, invalid attribute types
  - Edge case - Database errors: Update SKU to duplicate value (409 Conflict) if SKU update is allowed
  - Edge case - Concurrent updates: Optimistic locking or last-write-wins behavior
  - Table-driven test with multiple update scenarios

- [ ] T035 [US1] Integration test for DELETE /api/v1/products/{id} in internal/handlers/products_test.go
  - Happy path: Delete product and verify it's removed
  - Edge case - Data state: Delete non-existent product (404 Not Found)
  - Edge case - Input validation: Invalid UUID format (400 Bad Request)
  - Edge case - HTTP specifics: Wrong HTTP method (GET when DELETE expected)
  - Table-driven test with multiple delete scenarios

### Implementation for User Story 1

- [ ] T036 [P] [US1] Create Product model struct in internal/models/product.go (id, name, sku, description, product_type_id, attributes, created_at, updated_at)
- [ ] T037 [P] [US1] Create Attribute value types in internal/models/attribute.go (AttributeValue struct with type and value, helper methods for type validation)
- [ ] T038 [US1] Create product repository in internal/repository/products.go (Create, GetByID, List, Update, Delete methods using sqlx)
- [ ] T039 [US1] Create product validation logic in internal/validator/product.go (ValidateName, ValidateSKU, ValidateAttributes, ValidateAttributeTypes)
- [ ] T040 [US1] Implement POST /api/v1/products handler in internal/handlers/products.go (create product with attribute validation)
- [ ] T041 [US1] Implement GET /api/v1/products/{id} handler in internal/handlers/products.go (retrieve product by ID)
- [ ] T042 [US1] Implement GET /api/v1/products handler in internal/handlers/products.go (list products with pagination and filtering)
- [ ] T043 [US1] Implement PUT /api/v1/products/{id} handler in internal/handlers/products.go (update product with attribute validation)
- [ ] T044 [US1] Implement DELETE /api/v1/products/{id} handler in internal/handlers/products.go (delete product)
- [ ] T045 [US1] Wire product routes to Chi router in cmd/api/main.go (mount /api/v1/products routes)

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Add Product Variants (Priority: P2)

**Goal**: Enable creating and managing product variants with their own SKUs and attributes

**Independent Test**: Create parent product (US1), add variants via POST to /products/{id}/variants, retrieve variants, update variant attributes, delete individual variant, verify cascade delete when parent product is deleted.

### Integration Tests for User Story 2 (MANDATORY) ⚠️

- [ ] T046 [US2] Integration test for POST /api/v1/products/{productId}/variants in internal/handlers/variants_test.go
  - Happy path: Create variant with SKU and variant-specific attributes (size, color, price)
  - Edge case - Data state: Create variant for non-existent parent product (404 Not Found)
  - Edge case - Input validation: Empty SKU, duplicate SKU (across products and variants), invalid attribute types
  - Edge case - Boundary conditions: Product with zero variants, product with 1000+ variants
  - Edge case - Database errors: Duplicate variant SKU (409 Conflict), foreign key violation
  - Edge case - HTTP specifics: Missing Content-Type, invalid JSON, wrong method
  - Table-driven test with comprehensive edge cases per constitution

- [ ] T047 [US2] Integration test for GET /api/v1/products/{productId}/variants in internal/handlers/variants_test.go
  - Happy path: List all variants for a product
  - Edge case - Data state: List variants for non-existent product (404 Not Found)
  - Edge case - Boundary conditions: Product with zero variants, product with many variants
  - Table-driven test with various scenarios

- [ ] T048 [US2] Integration test for GET /api/v1/variants/{id} in internal/handlers/variants_test.go
  - Happy path: Retrieve single variant with all attributes and parent product reference
  - Edge case - Data state: Non-existent variant ID (404 Not Found)
  - Edge case - Input validation: Invalid UUID format (400 Bad Request)
  - Table-driven test with multiple variant fixtures

- [ ] T049 [US2] Integration test for PUT /api/v1/variants/{id} in internal/handlers/variants_test.go
  - Happy path: Update variant attributes without affecting parent product or other variants
  - Edge case - Data state: Update non-existent variant (404 Not Found)
  - Edge case - Database errors: Update SKU to duplicate (409 Conflict) if SKU update allowed
  - Table-driven test with update scenarios

- [ ] T050 [US2] Integration test for DELETE /api/v1/variants/{id} in internal/handlers/variants_test.go
  - Happy path: Delete variant, verify parent product and other variants unaffected
  - Edge case - Data state: Delete non-existent variant (404 Not Found)
  - Table-driven test with delete scenarios

- [ ] T051 [US2] Integration test for cascade delete behavior in internal/handlers/products_test.go
  - Happy path: Delete parent product with variants, verify all variants are cascade deleted
  - Edge case - Database errors: Verify foreign key cascade works correctly
  - Table-driven test with cascade scenarios

### Implementation for User Story 2

- [ ] T052 [P] [US2] Create Variant model struct in internal/models/variant.go (id, product_id, sku, attributes, created_at, updated_at)
- [ ] T053 [US2] Create variant repository in internal/repository/variants.go (Create, GetByID, ListByProductID, Update, Delete methods using sqlx)
- [ ] T054 [US2] Create variant validation logic in internal/validator/product.go (ValidateVariantSKU, ValidateVariantAttributes, CheckSKUUniqueness)
- [ ] T055 [US2] Implement POST /api/v1/products/{productId}/variants handler in internal/handlers/variants.go
- [ ] T056 [US2] Implement GET /api/v1/products/{productId}/variants handler in internal/handlers/variants.go
- [ ] T057 [US2] Implement GET /api/v1/variants/{id} handler in internal/handlers/variants.go
- [ ] T058 [US2] Implement PUT /api/v1/variants/{id} handler in internal/handlers/variants.go
- [ ] T059 [US2] Implement DELETE /api/v1/variants/{id} handler in internal/handlers/variants.go
- [ ] T060 [US2] Wire variant routes to Chi router in cmd/api/main.go (mount /api/v1/variants and /api/v1/products/{id}/variants routes)
- [ ] T061 [US2] Update GET /api/v1/products/{id} handler to include variants in response (modify internal/handlers/products.go)

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Attach Media Assets (Priority: P3)

**Goal**: Enable attaching images and videos to products and variants with display order

**Independent Test**: Create product (US1), attach multiple media assets via POST, retrieve assets in display order, update asset metadata, delete asset, verify cascade delete. Also test variant-level assets.

### Integration Tests for User Story 3 (MANDATORY) ⚠️

- [ ] T062 [US3] Integration test for POST /api/v1/products/{productId}/media in internal/handlers/media_assets_test.go
  - Happy path: Attach image and video to product with URL, alt_text, display_order
  - Edge case - Data state: Attach media to non-existent product (404 Not Found)
  - Edge case - Input validation: Invalid URL format, missing required fields, invalid media type (not 'image' or 'video')
  - Edge case - Boundary conditions: Product with zero assets, product with 100+ assets
  - Edge case - HTTP specifics: Missing Content-Type, invalid JSON
  - Table-driven test with comprehensive edge cases

- [ ] T063 [US3] Integration test for GET /api/v1/products/{productId}/media in internal/handlers/media_assets_test.go
  - Happy path: List media assets ordered by display_order
  - Edge case - Data state: List media for non-existent product (404 Not Found)
  - Edge case - Boundary conditions: Product with zero assets, verify sort order with negative/positive display_order values
  - Table-driven test with ordering scenarios

- [ ] T064 [US3] Integration test for POST /api/v1/variants/{variantId}/media in internal/handlers/media_assets_test.go
  - Happy path: Attach media to variant independently of parent product's assets
  - Edge case - Data state: Attach media to non-existent variant (404 Not Found)
  - Edge case - Database errors: Verify CHECK constraint (exactly one of product_id or variant_id)
  - Table-driven test

- [ ] T065 [US3] Integration test for GET /api/v1/variants/{variantId}/media in internal/handlers/media_assets_test.go
  - Happy path: List variant media assets ordered by display_order
  - Edge case - Boundary conditions: Variant with zero assets
  - Table-driven test

- [ ] T066 [US3] Integration test for PUT /api/v1/media/{id} in internal/handlers/media_assets_test.go
  - Happy path: Update alt_text and display_order, verify subsequent GET returns new order
  - Edge case - Data state: Update non-existent media asset (404 Not Found)
  - Edge case - Input validation: Invalid display_order (non-integer if applicable)
  - Table-driven test with update scenarios

- [ ] T067 [US3] Integration test for DELETE /api/v1/media/{id} in internal/handlers/media_assets_test.go
  - Happy path: Delete media asset without affecting product/variant or other assets
  - Edge case - Data state: Delete non-existent asset (404 Not Found)
  - Table-driven test

- [ ] T068 [US3] Integration test for cascade delete of media assets in internal/handlers/products_test.go and internal/handlers/variants_test.go
  - Happy path: Delete product with media assets, verify all assets cascade deleted
  - Happy path: Delete variant with media assets, verify all assets cascade deleted
  - Table-driven test with cascade scenarios

### Implementation for User Story 3

- [ ] T069 [P] [US3] Create MediaAsset model struct in internal/models/media_asset.go (id, product_id, variant_id, type, url, alt_text, display_order, created_at, updated_at)
- [ ] T070 [US3] Create media asset repository in internal/repository/media_assets.go (Create, GetByID, ListByProductID, ListByVariantID, Update, Delete methods using sqlx)
- [ ] T071 [US3] Create media asset validation logic in internal/validator/media.go (ValidateURL, ValidateMediaType, ValidateDisplayOrder)
- [ ] T072 [US3] Implement POST /api/v1/products/{productId}/media handler in internal/handlers/media_assets.go
- [ ] T073 [US3] Implement GET /api/v1/products/{productId}/media handler in internal/handlers/media_assets.go (with ORDER BY display_order)
- [ ] T074 [US3] Implement POST /api/v1/variants/{variantId}/media handler in internal/handlers/media_assets.go
- [ ] T075 [US3] Implement GET /api/v1/variants/{variantId}/media handler in internal/handlers/media_assets.go (with ORDER BY display_order)
- [ ] T076 [US3] Implement PUT /api/v1/media/{id} handler in internal/handlers/media_assets.go
- [ ] T077 [US3] Implement DELETE /api/v1/media/{id} handler in internal/handlers/media_assets.go
- [ ] T078 [US3] Wire media asset routes to Chi router in cmd/api/main.go
- [ ] T079 [US3] Update GET /api/v1/products/{id} handler to include media_assets in response (modify internal/handlers/products.go)
- [ ] T080 [US3] Update GET /api/v1/variants/{id} handler to include media_assets in response (modify internal/handlers/variants.go)

**Checkpoint**: At this point, User Stories 1, 2, AND 3 should all work independently

---

## Phase 6: User Story 4 - Define Product Type Templates (Priority: P4)

**Goal**: Enable creating product type templates with attribute definitions and validation

**Independent Test**: Create product type template with attribute definitions, create product referencing template, verify required attributes are validated, update template, delete template and verify products become untyped.

### Integration Tests for User Story 4 (MANDATORY) ⚠️

- [ ] T081 [US4] Integration test for POST /api/v1/product-types in internal/handlers/product_types_test.go
  - Happy path: Create product type with name, description, and attribute definitions (string, number, boolean, date, required/optional)
  - Edge case - Input validation: Empty name, duplicate type name (409 Conflict), invalid data_type in attribute definition
  - Edge case - Boundary conditions: Template with zero attribute definitions, template with 50+ definitions
  - Edge case - Database errors: Duplicate template name (409 Conflict), invalid data types
  - Edge case - HTTP specifics: Missing Content-Type, invalid JSON
  - Table-driven test with comprehensive edge cases

- [ ] T082 [US4] Integration test for GET /api/v1/product-types/{id} in internal/handlers/product_types_test.go
  - Happy path: Retrieve template with all attribute definitions
  - Edge case - Data state: Non-existent template ID (404 Not Found)
  - Table-driven test

- [ ] T083 [US4] Integration test for GET /api/v1/product-types (list all) in internal/handlers/product_types_test.go
  - Happy path: List all product type templates
  - Edge case - Boundary conditions: Zero templates, 50+ templates
  - Table-driven test

- [ ] T084 [US4] Integration test for template validation in internal/handlers/products_test.go
  - Happy path: Create product with product_type_id, verify required attributes are validated
  - Edge case - Input validation: Create product with missing required attribute (400 Bad Request with descriptive error)
  - Edge case - Input validation: Create product with attribute type mismatch (string when number required)
  - Edge case - Data state: Create product with non-existent product_type_id (404 or 400)
  - Table-driven test with validation scenarios

- [ ] T085 [US4] Integration test for PUT /api/v1/product-types/{id} in internal/handlers/product_types_test.go
  - Happy path: Update template name, description, add new attribute definitions, modify is_required flags
  - Edge case - Data state: Update non-existent template (404 Not Found)
  - Edge case - Database errors: Update name to duplicate (409 Conflict)
  - Table-driven test with update scenarios

- [ ] T086 [US4] Integration test for DELETE /api/v1/product-types/{id} in internal/handlers/product_types_test.go
  - Happy path: Delete template, verify products that referenced it become untyped (product_type_id set to null), verify products retain all attributes
  - Edge case - Data state: Delete non-existent template (404 Not Found)
  - Table-driven test with delete and SET NULL behavior

### Implementation for User Story 4

- [ ] T087 [P] [US4] Create ProductType model struct in internal/models/product_type.go (id, name, description, created_at, updated_at)
- [ ] T088 [P] [US4] Create AttributeDefinition model struct in internal/models/attribute_definition.go (id, product_type_id, attribute_name, data_type, is_required, created_at)
- [ ] T089 [US4] Create product type repository in internal/repository/product_types.go (Create, GetByID, List, Update, Delete methods, include attribute_definitions in queries)
- [ ] T090 [US4] Create template validation logic in internal/validator/template.go (ValidateProductType, ValidateAttributeDefinitions, ValidateProductAgainstTemplate)
- [ ] T091 [US4] Implement POST /api/v1/product-types handler in internal/handlers/product_types.go
- [ ] T092 [US4] Implement GET /api/v1/product-types/{id} handler in internal/handlers/product_types.go
- [ ] T093 [US4] Implement GET /api/v1/product-types handler in internal/handlers/product_types.go (list all)
- [ ] T094 [US4] Implement PUT /api/v1/product-types/{id} handler in internal/handlers/product_types.go
- [ ] T095 [US4] Implement DELETE /api/v1/product-types/{id} handler in internal/handlers/product_types.go
- [ ] T096 [US4] Wire product type routes to Chi router in cmd/api/main.go
- [ ] T097 [US4] Update POST /api/v1/products handler to validate against product type template if product_type_id is provided (modify internal/handlers/products.go)

**Checkpoint**: At this point, User Stories 1-4 should all work independently

---

## Phase 7: User Story 5 - Manage Multiple Product Types (Priority: P5)

**Goal**: Enable filtering and managing products by product type across multiple types

**Independent Test**: Create multiple product types, create products of different types, filter products by type, verify each product includes its type in response, verify error when creating product with non-existent type.

### Integration Tests for User Story 5 (MANDATORY) ⚠️

- [ ] T098 [US5] Integration test for product type filtering in internal/handlers/products_test.go
  - Happy path: Create multiple product types and products of each type, GET /api/v1/products?product_type_id={id} returns only products of that type
  - Edge case - Data state: Filter by non-existent product_type_id returns empty list or 404
  - Edge case - Boundary conditions: Filter when no products of that type exist
  - Table-driven test with filtering scenarios

- [ ] T099 [US5] Integration test for product type inclusion in responses in internal/handlers/products_test.go
  - Happy path: GET /api/v1/products (list all), verify each product includes its product_type information (id, name) in response
  - Edge case - Data state: Untyped products (product_type_id is null) have null product_type in response
  - Table-driven test

- [ ] T100 [US5] Integration test for invalid product type reference in internal/handlers/products_test.go
  - Edge case - Data state: POST product with non-existent product_type_id returns error (400 or 404) with descriptive message
  - Table-driven test with invalid type scenarios

- [ ] T101 [US5] Integration test for attribute search across product types in internal/handlers/products_test.go
  - Happy path: Search products by attributes specific to one type (e.g., JSONB query for "size" attribute), only products with that attribute are returned
  - Edge case - Boundary conditions: Search for attribute that no products have, returns empty list
  - Table-driven test with attribute search scenarios (stretch goal, may require additional endpoint)

### Implementation for User Story 5

- [ ] T102 [US5] Update product repository List method in internal/repository/products.go to support product_type_id filtering
- [ ] T103 [US5] Update GET /api/v1/products handler to support product_type_id query parameter (already implemented in T042, verify and enhance if needed)
- [ ] T104 [US5] Update product repository to JOIN with product_types table when retrieving products to include type information in response
- [ ] T105 [US5] Update POST /api/v1/products handler to validate product_type_id exists before creating product (modify internal/handlers/products.go)
- [ ] T106 [US5] Enhance product response model to include product_type information (name, description) in internal/models/product.go

**Checkpoint**: All user stories should now be independently functional

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T107 [P] Add structured logging for all CRUD operations (audit logging per FR-028) across all handlers
- [ ] T108 [P] Add input sanitization for XSS prevention (per FR-025) in validation layer
- [ ] T109 [P] Review and enhance error messages for validation failures (per SC-008 - 90% first-retry success rate)
- [ ] T110 [P] Add API versioning headers and documentation to all endpoints
- [ ] T111 Performance testing: Verify SC-001 (product creation <5s for 10 attributes)
- [ ] T112 Performance testing: Verify SC-003 (product retrieval <2s for 20 variants + 50 assets)
- [ ] T113 Performance testing: Verify SC-005 (1000 products without degradation)
- [ ] T114 Performance testing: Verify SC-007 (cascade delete <5s for 100 variants)
- [ ] T115 [P] Update README.md with API documentation, setup instructions, and quickstart guide
- [ ] T116 [P] Create API documentation from OpenAPI spec in docs/ directory (e.g., using Redoc or Swagger UI)
- [ ] T117 Verify all integration tests pass with real database: `make test`
- [ ] T118 Run test coverage report and verify >80% coverage: `go test ./... -cover`
- [ ] T119 Run quickstart.md validation (manual walkthrough of setup steps)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed) or sequentially in priority order (P1 → P2 → P3 → P4 → P5)
- **Polish (Phase 8)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Integrates with US1 (adds variants to products) but independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Integrates with US1 and US2 (adds media to products/variants) but independently testable
- **User Story 4 (P4)**: Can start after Foundational (Phase 2) - Integrates with US1 (validates products against templates) but independently testable
- **User Story 5 (P5)**: Can start after Foundational (Phase 2) - Builds on US1 and US4 (filtering and management) but independently testable

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Models before repositories
- Repositories before handlers
- Handlers before routing
- Core implementation before integration with other stories
- Story complete before moving to next priority

### Parallel Opportunities

- **Setup (Phase 1)**: T002-T005 (dependencies), T007-T010 (config files) can run in parallel
- **Foundational (Phase 2)**: T023-T028 (testutil and middleware) can run in parallel after migrations complete
- **User Story Tests**: Within each story, test files for different endpoints can be written in parallel
- **User Story Models**: Within each story, model structs can be created in parallel
- **Different User Stories**: Once Foundational phase completes, different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# After Foundational phase completes, launch these in parallel:

# Developer A: Write integration tests
Task T031: Integration test for POST /api/v1/products
Task T032: Integration test for GET /api/v1/products/{id}
Task T033: Integration test for GET /api/v1/products (list)
Task T034: Integration test for PUT /api/v1/products/{id}
Task T035: Integration test for DELETE /api/v1/products/{id}

# Developer B: Create models and validation (can start in parallel with tests)
Task T036: Create Product model struct
Task T037: Create Attribute value types

# After models complete:
Task T038: Create product repository
Task T039: Create product validation logic

# After repository and validation complete:
Task T040-T044: Implement all handlers
Task T045: Wire routes
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
4. Add User Story 3 → Test independently → Deploy/Demo
5. Add User Story 4 → Test independently → Deploy/Demo
6. Add User Story 5 → Test independently → Deploy/Demo
7. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (write tests, implement)
   - Developer B: User Story 2 (write tests, implement)
   - Developer C: User Story 3 (write tests, implement)
3. Stories complete and integrate independently
4. Continue with User Stories 4 and 5

---

## Task Summary

**Total Tasks**: 119

**Tasks by Phase**:
- Phase 1 (Setup): 10 tasks
- Phase 2 (Foundational): 20 tasks
- Phase 3 (User Story 1): 15 tasks (5 test tasks, 10 implementation tasks)
- Phase 4 (User Story 2): 16 tasks (6 test tasks, 10 implementation tasks)
- Phase 5 (User Story 3): 19 tasks (7 test tasks, 12 implementation tasks)
- Phase 6 (User Story 4): 17 tasks (6 test tasks, 11 implementation tasks)
- Phase 7 (User Story 5): 9 tasks (4 test tasks, 5 implementation tasks)
- Phase 8 (Polish): 13 tasks

**Parallel Opportunities**: 27 tasks marked with [P] can run in parallel within their phase

**Independent Test Criteria**:
- US1: Product CRUD works without variants, media, or templates
- US2: Variant management works with products from US1
- US3: Media asset management works with products/variants from US1/US2
- US4: Template validation works with products from US1
- US5: Multi-type management works with templates and products from US1/US4

**Suggested MVP Scope**: Phases 1, 2, and 3 (User Story 1 only) = 45 tasks

---

## Notes

- All tasks follow checklist format: `- [ ] [TaskID] [P?] [Story?] Description with file path`
- [P] indicates parallelizable tasks (different files, no dependencies)
- [Story] label maps task to specific user story for traceability
- Integration tests are MANDATORY per constitution (no mocking, real PostgreSQL)
- Tests MUST be written FIRST (TDD) and verified to fail before implementation
- Each user story is independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- All tests use table-driven patterns with comprehensive edge case coverage

