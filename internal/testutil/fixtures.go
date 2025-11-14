package testutil

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Product represents a test product fixture
type Product struct {
	ID            uuid.UUID              `db:"id"`
	Name          string                 `db:"name"`
	SKU           string                 `db:"sku"`
	Description   *string                `db:"description"`
	ProductTypeID *uuid.UUID             `db:"product_type_id"`
	Attributes    json.RawMessage        `db:"attributes"`
	CreatedAt     time.Time              `db:"created_at"`
	UpdatedAt     time.Time              `db:"updated_at"`
}

// Variant represents a test variant fixture
type Variant struct {
	ID         uuid.UUID       `db:"id"`
	ProductID  uuid.UUID       `db:"product_id"`
	SKU        string          `db:"sku"`
	Attributes json.RawMessage `db:"attributes"`
	CreatedAt  time.Time       `db:"created_at"`
	UpdatedAt  time.Time       `db:"updated_at"`
}

// MediaAsset represents a test media asset fixture
type MediaAsset struct {
	ID           uuid.UUID  `db:"id"`
	ProductID    *uuid.UUID `db:"product_id"`
	VariantID    *uuid.UUID `db:"variant_id"`
	Type         string     `db:"type"`
	URL          string     `db:"url"`
	AltText      *string    `db:"alt_text"`
	DisplayOrder int        `db:"display_order"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}

// ProductType represents a test product type fixture
type ProductType struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// AttributeDefinition represents a test attribute definition fixture
type AttributeDefinition struct {
	ID            uuid.UUID `db:"id"`
	ProductTypeID uuid.UUID `db:"product_type_id"`
	AttributeName string    `db:"attribute_name"`
	DataType      string    `db:"data_type"`
	IsRequired    bool      `db:"is_required"`
	CreatedAt     time.Time `db:"created_at"`
}

// CreateTestProduct creates a product in the test database
func CreateTestProduct(t *testing.T, tx *sqlx.Tx, name, sku string, attributes json.RawMessage) *Product {
	t.Helper()

	if attributes == nil {
		attributes = json.RawMessage("{}")
	}

	product := &Product{
		ID:         uuid.New(),
		Name:       name,
		SKU:        sku,
		Attributes: attributes,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	query := `
		INSERT INTO products (id, name, sku, description, product_type_id, attributes, created_at, updated_at)
		VALUES (:id, :name, :sku, :description, :product_type_id, :attributes, :created_at, :updated_at)
	`

	_, err := tx.NamedExec(query, product)
	if err != nil {
		t.Fatalf("failed to create test product: %v", err)
	}

	return product
}

// CreateTestVariant creates a variant in the test database
func CreateTestVariant(t *testing.T, tx *sqlx.Tx, productID uuid.UUID, sku string, attributes json.RawMessage) *Variant {
	t.Helper()

	if attributes == nil {
		attributes = json.RawMessage("{}")
	}

	variant := &Variant{
		ID:         uuid.New(),
		ProductID:  productID,
		SKU:        sku,
		Attributes: attributes,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	query := `
		INSERT INTO variants (id, product_id, sku, attributes, created_at, updated_at)
		VALUES (:id, :product_id, :sku, :attributes, :created_at, :updated_at)
	`

	_, err := tx.NamedExec(query, variant)
	if err != nil {
		t.Fatalf("failed to create test variant: %v", err)
	}

	return variant
}

// CreateTestMediaAsset creates a media asset in the test database
func CreateTestMediaAsset(t *testing.T, tx *sqlx.Tx, productID, variantID *uuid.UUID, assetType, url string, displayOrder int) *MediaAsset {
	t.Helper()

	asset := &MediaAsset{
		ID:           uuid.New(),
		ProductID:    productID,
		VariantID:    variantID,
		Type:         assetType,
		URL:          url,
		DisplayOrder: displayOrder,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	query := `
		INSERT INTO media_assets (id, product_id, variant_id, type, url, alt_text, display_order, created_at, updated_at)
		VALUES (:id, :product_id, :variant_id, :type, :url, :alt_text, :display_order, :created_at, :updated_at)
	`

	_, err := tx.NamedExec(query, asset)
	if err != nil {
		t.Fatalf("failed to create test media asset: %v", err)
	}

	return asset
}

// CreateTestProductType creates a product type in the test database
func CreateTestProductType(t *testing.T, tx *sqlx.Tx, name string, attributeDefs []AttributeDefinition) *ProductType {
	t.Helper()

	productType := &ProductType{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	query := `
		INSERT INTO product_types (id, name, description, created_at, updated_at)
		VALUES (:id, :name, :description, :created_at, :updated_at)
	`

	_, err := tx.NamedExec(query, productType)
	if err != nil {
		t.Fatalf("failed to create test product type: %v", err)
	}

	// Create attribute definitions if provided
	for i := range attributeDefs {
		def := &attributeDefs[i]
		def.ID = uuid.New()
		def.ProductTypeID = productType.ID
		def.CreatedAt = time.Now()

		defQuery := `
			INSERT INTO attribute_definitions (id, product_type_id, attribute_name, data_type, is_required, created_at)
			VALUES (:id, :product_type_id, :attribute_name, :data_type, :is_required, :created_at)
		`

		_, err := tx.NamedExec(defQuery, def)
		if err != nil {
			t.Fatalf("failed to create test attribute definition: %v", err)
		}
	}

	return productType
}

// StringPtr returns a pointer to a string value
func StringPtr(s string) *string {
	return &s
}

// UUIDPtr returns a pointer to a UUID value
func UUIDPtr(id uuid.UUID) *uuid.UUID {
	return &id
}

