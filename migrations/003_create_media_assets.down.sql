-- Drop indexes first
DROP INDEX IF EXISTS idx_media_assets_variant_id;
DROP INDEX IF EXISTS idx_media_assets_product_id;

-- Drop media_assets table
DROP TABLE IF EXISTS media_assets;

