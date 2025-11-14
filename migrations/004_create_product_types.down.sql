-- Remove foreign key constraint from products table first
ALTER TABLE products
DROP CONSTRAINT IF EXISTS fk_products_product_type;

-- Drop index
DROP INDEX IF EXISTS idx_product_types_name;

-- Drop product_types table
DROP TABLE IF EXISTS product_types;

