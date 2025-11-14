-- Create product_types table for product templates
CREATE TABLE product_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL CHECK (char_length(name) >= 1 AND char_length(name) <= 100),
    description TEXT CHECK (char_length(description) <= 1000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create index for product_types table
CREATE INDEX idx_product_types_name ON product_types(name);

-- Add comment explaining the table
COMMENT ON TABLE product_types IS 'Product type templates defining attribute schemas for different product categories';

-- Now add the foreign key constraint to products table that we couldn't add in migration 001
ALTER TABLE products
ADD CONSTRAINT fk_products_product_type
FOREIGN KEY (product_type_id)
REFERENCES product_types(id)
ON DELETE SET NULL;

