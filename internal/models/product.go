package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Product represents a product in the PIM system
type Product struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	Name          string          `json:"name" db:"name"`
	SKU           string          `json:"sku" db:"sku"`
	Description   *string         `json:"description,omitempty" db:"description"`
	ProductTypeID *uuid.UUID      `json:"product_type_id,omitempty" db:"product_type_id"`
	Attributes    json.RawMessage `json:"attributes" db:"attributes"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

// ProductCreateRequest represents the request body for creating a product
type ProductCreateRequest struct {
	Name          string                 `json:"name"`
	SKU           string                 `json:"sku"`
	Description   *string                `json:"description,omitempty"`
	ProductTypeID *uuid.UUID             `json:"product_type_id,omitempty"`
	Attributes    map[string]interface{} `json:"attributes,omitempty"`
}

// ProductUpdateRequest represents the request body for updating a product
type ProductUpdateRequest struct {
	Name          *string                `json:"name,omitempty"`
	Description   *string                `json:"description,omitempty"`
	ProductTypeID *uuid.UUID             `json:"product_type_id,omitempty"`
	Attributes    map[string]interface{} `json:"attributes,omitempty"`
}

// ProductListResponse represents a paginated list of products
type ProductListResponse struct {
	Products   []Product  `json:"products"`
	Pagination Pagination `json:"pagination"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

// Validate validates the product create request
func (r *ProductCreateRequest) Validate() error {
	if r.Name == "" {
		return NewValidationError("name is required")
	}
	if len(r.Name) > 500 {
		return NewValidationError("name must be at most 500 characters")
	}
	if r.SKU == "" {
		return NewValidationError("sku is required")
	}
	if len(r.SKU) > 100 {
		return NewValidationError("sku must be at most 100 characters")
	}
	if r.Description != nil && len(*r.Description) > 10000 {
		return NewValidationError("description must be at most 10000 characters")
	}
	return nil
}

// Validate validates the product update request
func (r *ProductUpdateRequest) Validate() error {
	if r.Name != nil && *r.Name == "" {
		return NewValidationError("name cannot be empty")
	}
	if r.Name != nil && len(*r.Name) > 500 {
		return NewValidationError("name must be at most 500 characters")
	}
	if r.Description != nil && len(*r.Description) > 10000 {
		return NewValidationError("description must be at most 10000 characters")
	}
	return nil
}

// ValidationErr is a simple error type for validation errors
type ValidationErr struct {
	msg string
}

func (e ValidationErr) Error() string {
	return e.msg
}

// NewValidationError creates a new validation error
func NewValidationError(message string) error {
	return ValidationErr{msg: message}
}

