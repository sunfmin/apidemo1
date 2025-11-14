-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for products table
CREATE TRIGGER update_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create trigger for variants table
CREATE TRIGGER update_variants_updated_at
    BEFORE UPDATE ON variants
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create trigger for media_assets table
CREATE TRIGGER update_media_assets_updated_at
    BEFORE UPDATE ON media_assets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create trigger for product_types table
CREATE TRIGGER update_product_types_updated_at
    BEFORE UPDATE ON product_types
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comment explaining the function
COMMENT ON FUNCTION update_updated_at_column() IS 'Automatically updates the updated_at column to the current timestamp on row update';

