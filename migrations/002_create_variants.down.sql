-- Drop indexes first
DROP INDEX IF EXISTS idx_variants_attributes;
DROP INDEX IF EXISTS idx_variants_sku;
DROP INDEX IF EXISTS idx_variants_product_id;

-- Drop variants table
DROP TABLE IF EXISTS variants;

