package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"apidemo1/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// VariantRepository handles database operations for variants
type VariantRepository struct {
	db *sqlx.DB
}

// NewVariantRepository creates a new variant repository
func NewVariantRepository(db *sqlx.DB) *VariantRepository {
	return &VariantRepository{db: db}
}

// CreateWithTx creates a variant within a transaction
func (r *VariantRepository) CreateWithTx(tx *sqlx.Tx, productID uuid.UUID, req *models.VariantCreateRequest) (*models.Variant, error) {
	// Check if parent product exists
	var exists bool
	err := tx.Get(&exists, "SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)", productID)
	if err != nil {
		return nil, fmt.Errorf("failed to check product existence: %w", err)
	}
	if !exists {
		return nil, errors.New("parent product not found")
	}

	// Sanitize and validate attributes
	attrs, err := models.SanitizeAttributes(req.Attributes)
	if err != nil {
		return nil, fmt.Errorf("invalid attributes: %w", err)
	}

	variant := &models.Variant{
		ID:         uuid.New(),
		ProductID:  productID,
		SKU:        req.SKU,
		Attributes: attrs,
	}

	query := `
		INSERT INTO variants (id, product_id, sku, attributes)
		VALUES ($1, $2, $3, $4)
		RETURNING id, product_id, sku, attributes, created_at, updated_at
	`

	err = tx.QueryRowx(
		query,
		variant.ID,
		variant.ProductID,
		variant.SKU,
		variant.Attributes,
	).StructScan(variant)

	if err != nil {
		// Check for unique constraint violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, errors.New("already exists")
		}
		// Check for foreign key violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
			return nil, errors.New("parent product not found")
		}
		return nil, fmt.Errorf("failed to create variant: %w", err)
	}

	return variant, nil
}

// GetByIDWithTx retrieves a variant within a transaction
func (r *VariantRepository) GetByIDWithTx(tx *sqlx.Tx, id uuid.UUID) (*models.Variant, error) {
	variant := &models.Variant{}
	query := `
		SELECT id, product_id, sku, attributes, created_at, updated_at
		FROM variants
		WHERE id = $1
	`

	err := tx.Get(variant, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("variant not found")
		}
		return nil, fmt.Errorf("failed to get variant: %w", err)
	}

	return variant, nil
}

// ListByProductIDWithTx retrieves all variants for a product within a transaction
func (r *VariantRepository) ListByProductIDWithTx(tx *sqlx.Tx, productID uuid.UUID) ([]models.Variant, error) {
	// Check if parent product exists
	var exists bool
	err := tx.Get(&exists, "SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)", productID)
	if err != nil {
		return nil, fmt.Errorf("failed to check product existence: %w", err)
	}
	if !exists {
		return nil, errors.New("parent product not found")
	}

	variants := []models.Variant{}
	query := `
		SELECT id, product_id, sku, attributes, created_at, updated_at
		FROM variants
		WHERE product_id = $1
		ORDER BY created_at ASC
	`

	err = tx.Select(&variants, query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to list variants: %w", err)
	}

	return variants, nil
}

// UpdateWithTx updates a variant within a transaction
func (r *VariantRepository) UpdateWithTx(tx *sqlx.Tx, id uuid.UUID, req *models.VariantUpdateRequest) (*models.Variant, error) {
	// Check if variant exists
	existing, err := r.GetByIDWithTx(tx, id)
	if err != nil {
		return nil, err
	}

	// If no attributes to update, return existing
	if req.Attributes == nil {
		return existing, nil
	}

	// Sanitize and validate attributes
	attrs, err := models.SanitizeAttributes(req.Attributes)
	if err != nil {
		return nil, fmt.Errorf("invalid attributes: %w", err)
	}

	query := `
		UPDATE variants
		SET attributes = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, product_id, sku, attributes, created_at, updated_at
	`

	variant := &models.Variant{}
	err = tx.QueryRowx(query, attrs, id).StructScan(variant)
	if err != nil {
		return nil, fmt.Errorf("failed to update variant: %w", err)
	}

	return variant, nil
}

// DeleteWithTx deletes a variant within a transaction
func (r *VariantRepository) DeleteWithTx(tx *sqlx.Tx, id uuid.UUID) error {
	result, err := tx.Exec("DELETE FROM variants WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete variant: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return errors.New("variant not found")
	}

	return nil
}

