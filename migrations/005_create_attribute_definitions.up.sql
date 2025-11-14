-- Create attribute_definitions table for template schemas
CREATE TABLE attribute_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_type_id UUID NOT NULL REFERENCES product_types(id) ON DELETE CASCADE,
    attribute_name TEXT NOT NULL CHECK (char_length(attribute_name) >= 1 AND char_length(attribute_name) <= 100),
    data_type TEXT NOT NULL CHECK (data_type IN ('string', 'number', 'boolean', 'date')),
    is_required BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(product_type_id, attribute_name)
);

-- Create index for attribute_definitions table
CREATE INDEX idx_attribute_definitions_product_type_id ON attribute_definitions(product_type_id);

-- Add comment explaining the table
COMMENT ON TABLE attribute_definitions IS 'Attribute schema definitions for product type templates';
COMMENT ON COLUMN attribute_definitions.is_required IS 'Whether this attribute is required for products of this type';

