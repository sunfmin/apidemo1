# Data Model: PIM Product API

**Feature**: 001-pim-product-api  
**Date**: 2025-11-14  
**Database**: PostgreSQL 15+

## Overview

The PIM system data model supports flexible product catalogs with varying attribute schemas. The core design uses JSONB columns for flexible attributes, enabling different product types without schema migrations. All entities use UUID primary keys for scalability and security.

## Entity Relationship Diagram

```
┌─────────────────────┐
│  product_types      │
│  (templates)        │
└──────────┬──────────┘
           │
           │ 0..1 (optional)
           ↓
┌─────────────────────┐         ┌─────────────────────┐
│  products           │←────────│  variants           │
│  (parent products)  │  1..N   │  (SKU level)        │
└──────────┬──────────┘         └──────────┬──────────┘
           │                               │
           │ 1..N                          │ 1..N
           ↓                               ↓
┌─────────────────────┐         ┌─────────────────────┐
│  media_assets       │         │  media_assets       │
│  (product level)    │         │  (variant level)    │
└─────────────────────┘         └─────────────────────┘

┌─────────────────────┐
│  product_types      │
└──────────┬──────────┘
           │ 1..N
           ↓
┌─────────────────────┐
│  attribute_defs     │
│  (template schema)  │
└─────────────────────┘
```

## Entities

### 1. Products

**Purpose**: Represents parent products in the catalog. Each product has core identifying information and a flexible collection of custom attributes stored in JSONB format.

**Table**: `products`

**Schema**:
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

CREATE INDEX idx_products_sku ON products(sku);
CREATE INDEX idx_products_product_type_id ON products(product_type_id);
CREATE INDEX idx_products_attributes ON products USING GIN (attributes);
CREATE INDEX idx_products_created_at ON products(created_at DESC);
```

**Fields**:
- `id`: Unique identifier (UUID, auto-generated)
- `name`: Product display name (required, max 500 chars)
- `sku`: Stock Keeping Unit, globally unique across products and variants (required, max 100 chars)
- `description`: Long-form product description (optional, max 10000 chars)
- `product_type_id`: Optional reference to product type template (null if untyped)
- `attributes`: Flexible attributes as JSONB (format: `{"attr_name": {"type": "string|number|boolean|date", "value": ...}}`)
- `created_at`: Timestamp when product was created (auto-set)
- `updated_at`: Timestamp when product was last modified (auto-updated via trigger)

**Constraints**:
- SKU must be unique across all products and variants
- Name cannot be empty
- If `product_type_id` is set, must reference valid product type
- Attributes JSONB must be valid JSON object

**Validation Rules** (application layer):
- SKU format: alphanumeric, hyphens, underscores only
- Name: 1-500 characters
- Description: 0-10000 characters
- If product type is set, validate required attributes exist
- Validate attribute values match declared types

**Relationships**:
- 0..1 relationship with `product_types` (optional template)
- 1..N relationship with `variants` (cascade delete variants when product deleted)
- 1..N relationship with `media_assets` (cascade delete assets when product deleted)

---

### 2. Variants

**Purpose**: Represents specific configurations of a parent product (e.g., size/color combinations). Each variant has its own SKU and can be independently priced and inventoried.

**Table**: `variants`

**Schema**:
```sql
CREATE TABLE variants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sku TEXT UNIQUE NOT NULL,
    attributes JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_variants_product_id ON variants(product_id);
CREATE INDEX idx_variants_sku ON variants(sku);
CREATE INDEX idx_variants_attributes ON variants USING GIN (attributes);
```

**Fields**:
- `id`: Unique identifier (UUID, auto-generated)
- `product_id`: Reference to parent product (required, foreign key)
- `sku`: Variant-specific SKU, globally unique (required, max 100 chars)
- `attributes`: Variant-specific attributes as JSONB (format: same as products)
- `created_at`: Timestamp when variant was created (auto-set)
- `updated_at`: Timestamp when variant was last modified (auto-updated via trigger)

**Constraints**:
- Must reference valid product
- SKU must be unique across all products and variants
- Cascade delete when parent product is deleted

**Validation Rules** (application layer):
- SKU format: alphanumeric, hyphens, underscores only
- Parent product must exist
- SKU must not conflict with any product or variant SKU
- Validate attribute values match declared types

**Relationships**:
- N..1 relationship with `products` (required parent)
- 1..N relationship with `media_assets` (cascade delete assets when variant deleted)

---

### 3. Media Assets

**Purpose**: Represents images and videos attached to products or variants. Stores metadata and display order for media presentation.

**Table**: `media_assets`

**Schema**:
```sql
CREATE TABLE media_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id) ON DELETE CASCADE,
    variant_id UUID REFERENCES variants(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('image', 'video')),
    url TEXT NOT NULL,
    alt_text TEXT,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT media_asset_owner CHECK (
        (product_id IS NOT NULL AND variant_id IS NULL) OR
        (product_id IS NULL AND variant_id IS NOT NULL)
    )
);

CREATE INDEX idx_media_assets_product_id ON media_assets(product_id, display_order);
CREATE INDEX idx_media_assets_variant_id ON media_assets(variant_id, display_order);
```

**Fields**:
- `id`: Unique identifier (UUID, auto-generated)
- `product_id`: Reference to product (if product-level asset, mutually exclusive with variant_id)
- `variant_id`: Reference to variant (if variant-level asset, mutually exclusive with product_id)
- `type`: Media type (`image` or `video`, required)
- `url`: URL where media is stored (required, max 2000 chars)
- `alt_text`: Accessibility description (optional, max 500 chars)
- `display_order`: Integer for sorting (default 0, can be negative or positive)
- `created_at`: Timestamp when asset was created (auto-set)
- `updated_at`: Timestamp when asset was last modified (auto-updated via trigger)

**Constraints**:
- Exactly one of `product_id` or `variant_id` must be set (enforced by CHECK constraint)
- Type must be 'image' or 'video'
- Cascade delete when parent product or variant is deleted

**Validation Rules** (application layer):
- URL format validation (valid HTTP/HTTPS URL)
- Type must be 'image' or 'video'
- Alt text: 0-500 characters
- display_order: any integer

**Relationships**:
- N..1 relationship with `products` OR `variants` (exactly one parent)

**Query Pattern** (ordered assets for a product):
```sql
SELECT * FROM media_assets 
WHERE product_id = $1 
ORDER BY display_order ASC, created_at ASC;
```

---

### 4. Product Types (Templates)

**Purpose**: Defines reusable schemas for product categories. Specifies which attributes are required or optional for a product type.

**Table**: `product_types`

**Schema**:
```sql
CREATE TABLE product_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_product_types_name ON product_types(name);
```

**Fields**:
- `id`: Unique identifier (UUID, auto-generated)
- `name`: Type name (required, unique, max 100 chars, e.g., "Clothing", "Electronics")
- `description`: Description of what this type represents (optional, max 1000 chars)
- `created_at`: Timestamp when type was created (auto-set)
- `updated_at`: Timestamp when type was last modified (auto-updated via trigger)

**Constraints**:
- Name must be unique
- Name cannot be empty

**Validation Rules** (application layer):
- Name: 1-100 characters
- Description: 0-1000 characters

**Relationships**:
- 1..N relationship with `products` (products can reference this type)
- 1..N relationship with `attribute_definitions` (defines schema for this type)

---

### 5. Attribute Definitions

**Purpose**: Part of a product type template. Specifies an attribute's name, data type, and whether it's required for products of that type.

**Table**: `attribute_definitions`

**Schema**:
```sql
CREATE TABLE attribute_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_type_id UUID NOT NULL REFERENCES product_types(id) ON DELETE CASCADE,
    attribute_name TEXT NOT NULL,
    data_type TEXT NOT NULL CHECK (data_type IN ('string', 'number', 'boolean', 'date')),
    is_required BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(product_type_id, attribute_name)
);

CREATE INDEX idx_attribute_definitions_product_type_id ON attribute_definitions(product_type_id);
```

**Fields**:
- `id`: Unique identifier (UUID, auto-generated)
- `product_type_id`: Reference to product type (required, foreign key)
- `attribute_name`: Name of the attribute (required, max 100 chars)
- `data_type`: Expected data type (`string`, `number`, `boolean`, `date`, required)
- `is_required`: Whether this attribute is required for products of this type (default false)
- `created_at`: Timestamp when definition was created (auto-set)

**Constraints**:
- Must reference valid product type
- Combination of (product_type_id, attribute_name) must be unique
- Data type must be one of: 'string', 'number', 'boolean', 'date'
- Cascade delete when parent product type is deleted

**Validation Rules** (application layer):
- Attribute name: 1-100 characters, alphanumeric and underscores
- Data type: must be one of the allowed types

**Relationships**:
- N..1 relationship with `product_types` (required parent)

**Usage Pattern** (validate product attributes):
```sql
-- Get required attributes for a product type
SELECT attribute_name, data_type 
FROM attribute_definitions 
WHERE product_type_id = $1 AND is_required = true;
```

---

## Flexible Attribute Storage Format

Attributes are stored in JSONB columns with the following structure:

```json
{
  "attribute_name": {
    "type": "string|number|boolean|date",
    "value": <actual value>
  }
}
```

**Examples**:

```json
{
  "color": {
    "type": "string",
    "value": "blue"
  },
  "size": {
    "type": "string",
    "value": "XL"
  },
  "weight": {
    "type": "number",
    "value": 2.5
  },
  "in_stock": {
    "type": "boolean",
    "value": true
  },
  "release_date": {
    "type": "date",
    "value": "2025-01-15"
  }
}
```

**Rationale**:
- Type metadata enables runtime validation
- Structured format enables consistent querying
- JSONB indexing enables fast attribute lookups

**Query Patterns**:

```sql
-- Find products with specific attribute value
SELECT * FROM products 
WHERE attributes @> '{"color": {"type": "string", "value": "blue"}}';

-- Check if attribute exists
SELECT * FROM products 
WHERE attributes ? 'color';

-- Extract specific attribute
SELECT id, name, attributes->'color'->>'value' as color 
FROM products 
WHERE attributes ? 'color';
```

---

## Database Migrations

Migrations are located in `/migrations/` directory using golang-migrate format:

1. `001_create_products.up.sql` / `001_create_products.down.sql`
2. `002_create_variants.up.sql` / `002_create_variants.down.sql`
3. `003_create_media_assets.up.sql` / `003_create_media_assets.down.sql`
4. `004_create_product_types.up.sql` / `004_create_product_types.down.sql`
5. `005_create_attribute_definitions.up.sql` / `005_create_attribute_definitions.down.sql`
6. `006_create_triggers.up.sql` / `006_create_triggers.down.sql` (for updated_at auto-update)

---

## Indexes Strategy

**Performance Considerations**:

- **UUID Primary Keys**: GIN indexes on UUID for fast lookups
- **SKU Lookups**: B-tree indexes on SKU columns (products and variants)
- **JSONB Attributes**: GIN indexes on attributes columns for attribute queries
- **Foreign Keys**: B-tree indexes on all foreign key columns
- **Time Ordering**: B-tree index on created_at for chronological queries
- **Media Assets**: Composite index on (product_id/variant_id, display_order) for sorted retrieval

**Index Maintenance**:
- Monitor index usage with PostgreSQL's `pg_stat_user_indexes`
- Consider partial indexes if certain attribute queries are common
- Use `EXPLAIN ANALYZE` to validate query plans in tests

---

## Constraints Summary

| Constraint Type | Tables | Enforcement |
|----------------|--------|-------------|
| Unique SKU | products, variants | Database UNIQUE constraint |
| Unique Template Name | product_types | Database UNIQUE constraint |
| Unique Attribute Name per Type | attribute_definitions | Database UNIQUE (product_type_id, attribute_name) |
| Media Owner | media_assets | Database CHECK (exactly one of product_id or variant_id) |
| Valid Media Type | media_assets | Database CHECK (type IN ('image', 'video')) |
| Valid Data Type | attribute_definitions | Database CHECK (data_type IN (...)) |
| Cascade Delete | variants, media_assets, attribute_definitions | Database ON DELETE CASCADE |
| Set Null on Type Delete | products.product_type_id | Database ON DELETE SET NULL |
| Required Template Attributes | products | Application layer validation |
| Attribute Type Matching | products, variants | Application layer validation |

---

## Transaction Boundaries

**CRUD Operations** (single entity):
- Each create/update/delete operation wraps in a transaction
- Example: Creating product with attributes is atomic

**Multi-Entity Operations** (require explicit transactions):
- Creating product with multiple variants: wrap in transaction
- Creating product with media assets: wrap in transaction
- Deleting product (cascade to variants and assets): single transaction (handled by database)

**Test Isolation**:
- Each integration test runs in a transaction that's rolled back
- Ensures clean database state between tests
- No need for manual cleanup or truncation

