package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Variant represents a product variant in the PIM system
type Variant struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	ProductID  uuid.UUID       `json:"product_id" db:"product_id"`
	SKU        string          `json:"sku" db:"sku"`
	Attributes json.RawMessage `json:"attributes" db:"attributes"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at" db:"updated_at"`
}

// VariantCreateRequest represents the request body for creating a variant
type VariantCreateRequest struct {
	SKU        string                 `json:"sku"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// VariantUpdateRequest represents the request body for updating a variant
type VariantUpdateRequest struct {
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// VariantListResponse represents a list of variants
type VariantListResponse struct {
	Variants []Variant `json:"variants"`
}

// Validate validates the variant create request
func (r *VariantCreateRequest) Validate() error {
	if r.SKU == "" {
		return NewValidationError("sku is required")
	}
	if len(r.SKU) > 100 {
		return NewValidationError("sku must be at most 100 characters")
	}
	return nil
}

