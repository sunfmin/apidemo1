-- Create products table with JSONB attributes for flexible product data
-- Note: Foreign key constraint for product_type_id will be added in a later migration after product_types table is created
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL CHECK (char_length(name) >= 1 AND char_length(name) <= 500),
    sku TEXT UNIQUE NOT NULL CHECK (char_length(sku) >= 1 AND char_length(sku) <= 100),
    description TEXT CHECK (char_length(description) <= 10000),
    product_type_id UUID,
    attributes JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for products table
CREATE INDEX idx_products_sku ON products(sku);
CREATE INDEX idx_products_product_type_id ON products(product_type_id);
CREATE INDEX idx_products_attributes ON products USING GIN (attributes);
CREATE INDEX idx_products_created_at ON products(created_at DESC);

-- Add comment explaining the table
COMMENT ON TABLE products IS 'Parent products with flexible JSONB attributes';
COMMENT ON COLUMN products.attributes IS 'Flexible attributes stored as JSONB with format: {"attr_name": {"type": "string|number|boolean|date", "value": <actual value>}}';

