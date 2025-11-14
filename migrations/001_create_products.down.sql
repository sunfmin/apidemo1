-- Drop indexes first
DROP INDEX IF EXISTS idx_products_created_at;
DROP INDEX IF EXISTS idx_products_attributes;
DROP INDEX IF EXISTS idx_products_product_type_id;
DROP INDEX IF EXISTS idx_products_sku;

-- Drop products table
DROP TABLE IF EXISTS products;

