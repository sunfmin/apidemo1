-- Drop triggers
DROP TRIGGER IF EXISTS update_product_types_updated_at ON product_types;
DROP TRIGGER IF EXISTS update_media_assets_updated_at ON media_assets;
DROP TRIGGER IF EXISTS update_variants_updated_at ON variants;
DROP TRIGGER IF EXISTS update_products_updated_at ON products;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

