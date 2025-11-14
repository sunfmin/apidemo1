-- Create media_assets table for images and videos
CREATE TABLE media_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id) ON DELETE CASCADE,
    variant_id UUID REFERENCES variants(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('image', 'video')),
    url TEXT NOT NULL CHECK (char_length(url) <= 2000),
    alt_text TEXT CHECK (char_length(alt_text) <= 500),
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT media_asset_owner CHECK (
        (product_id IS NOT NULL AND variant_id IS NULL) OR
        (product_id IS NULL AND variant_id IS NOT NULL)
    )
);

-- Create indexes for media_assets table
CREATE INDEX idx_media_assets_product_id ON media_assets(product_id, display_order);
CREATE INDEX idx_media_assets_variant_id ON media_assets(variant_id, display_order);

-- Add comment explaining the table
COMMENT ON TABLE media_assets IS 'Images and videos attached to products or variants';
COMMENT ON CONSTRAINT media_asset_owner ON media_assets IS 'Ensures exactly one of product_id or variant_id is set';

