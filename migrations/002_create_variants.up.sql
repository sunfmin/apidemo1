-- Create variants table for product configurations
CREATE TABLE variants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sku TEXT UNIQUE NOT NULL CHECK (char_length(sku) >= 1 AND char_length(sku) <= 100),
    attributes JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for variants table
CREATE INDEX idx_variants_product_id ON variants(product_id);
CREATE INDEX idx_variants_sku ON variants(sku);
CREATE INDEX idx_variants_attributes ON variants USING GIN (attributes);

-- Add comment explaining the table
COMMENT ON TABLE variants IS 'Product variants (e.g., size/color combinations) with their own SKUs';
COMMENT ON COLUMN variants.attributes IS 'Variant-specific attributes stored as JSONB with same format as products.attributes';

