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

// ProductRepository handles database operations for products
type ProductRepository struct {
	db *sqlx.DB
}

// NewProductRepository creates a new product repository
func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create inserts a new product into the database
func (r *ProductRepository) Create(req *models.ProductCreateRequest) (*models.Product, error) {
	// Sanitize and validate attributes
	attrs, err := models.SanitizeAttributes(req.Attributes)
	if err != nil {
		return nil, fmt.Errorf("invalid attributes: %w", err)
	}

	product := &models.Product{
		ID:            uuid.New(),
		Name:          req.Name,
		SKU:           req.SKU,
		Description:   req.Description,
		ProductTypeID: req.ProductTypeID,
		Attributes:    attrs,
	}

	query := `
		INSERT INTO products (id, name, sku, description, product_type_id, attributes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, sku, description, product_type_id, attributes, created_at, updated_at
	`

	err = r.db.QueryRowx(
		query,
		product.ID,
		product.Name,
		product.SKU,
		product.Description,
		product.ProductTypeID,
		product.Attributes,
	).StructScan(product)

	if err != nil {
		// Check for unique constraint violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, errors.New("product with this SKU already exists")
		}
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

// GetByID retrieves a product by its ID
func (r *ProductRepository) GetByID(id uuid.UUID) (*models.Product, error) {
	product := &models.Product{}
	query := `
		SELECT id, name, sku, description, product_type_id, attributes, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	err := r.db.Get(product, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("product not found")
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}

// List retrieves a paginated list of products
func (r *ProductRepository) List(page, pageSize int, productTypeID *uuid.UUID) ([]models.Product, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	offset := (page - 1) * pageSize

	// Build query based on filters
	query := `
		SELECT id, name, sku, description, product_type_id, attributes, created_at, updated_at
		FROM products
	`
	countQuery := `SELECT COUNT(*) FROM products`
	
	args := []interface{}{}
	whereClause := ""

	if productTypeID != nil {
		whereClause = " WHERE product_type_id = $1"
		args = append(args, productTypeID)
	}

	query += whereClause + ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, pageSize, offset)

	// Get total count
	var total int
	countArgs := []interface{}{}
	if productTypeID != nil {
		countArgs = append(countArgs, productTypeID)
	}
	err := r.db.Get(&total, countQuery+whereClause, countArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Get products
	products := []models.Product{}
	err = r.db.Select(&products, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}

	return products, total, nil
}

// Update updates a product
func (r *ProductRepository) Update(id uuid.UUID, req *models.ProductUpdateRequest) (*models.Product, error) {
	// Check if product exists
	existing, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Build dynamic update query
	updates := []string{}
	args := []interface{}{}
	argCount := 1

	if req.Name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", argCount))
		args = append(args, *req.Name)
		argCount++
	}

	if req.Description != nil {
		updates = append(updates, fmt.Sprintf("description = $%d", argCount))
		args = append(args, *req.Description)
		argCount++
	}

	if req.ProductTypeID != nil {
		updates = append(updates, fmt.Sprintf("product_type_id = $%d", argCount))
		args = append(args, *req.ProductTypeID)
		argCount++
	}

	if req.Attributes != nil {
		attrs, err := models.SanitizeAttributes(req.Attributes)
		if err != nil {
			return nil, fmt.Errorf("invalid attributes: %w", err)
		}
		updates = append(updates, fmt.Sprintf("attributes = $%d", argCount))
		args = append(args, attrs)
		argCount++
	}

	if len(updates) == 0 {
		// No updates, return existing product
		return existing, nil
	}

	// Add ID to args
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE products
		SET %s, updated_at = NOW()
		WHERE id = $%d
		RETURNING id, name, sku, description, product_type_id, attributes, created_at, updated_at
	`, join(updates, ", "), argCount)

	product := &models.Product{}
	err = r.db.QueryRowx(query, args...).StructScan(product)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

// Delete deletes a product
func (r *ProductRepository) Delete(id uuid.UUID) error {
	result, err := r.db.Exec("DELETE FROM products WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return errors.New("product not found")
	}

	return nil
}

// Helper function to join strings
func join(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// CreateWithTx creates a product within a transaction
func (r *ProductRepository) CreateWithTx(tx *sqlx.Tx, req *models.ProductCreateRequest) (*models.Product, error) {
	// Sanitize and validate attributes
	attrs, err := models.SanitizeAttributes(req.Attributes)
	if err != nil {
		return nil, fmt.Errorf("invalid attributes: %w", err)
	}

	product := &models.Product{
		ID:            uuid.New(),
		Name:          req.Name,
		SKU:           req.SKU,
		Description:   req.Description,
		ProductTypeID: req.ProductTypeID,
		Attributes:    attrs,
	}

	query := `
		INSERT INTO products (id, name, sku, description, product_type_id, attributes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, sku, description, product_type_id, attributes, created_at, updated_at
	`

	err = tx.QueryRowx(
		query,
		product.ID,
		product.Name,
		product.SKU,
		product.Description,
		product.ProductTypeID,
		product.Attributes,
	).StructScan(product)

	if err != nil {
		// Check for unique constraint violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, errors.New("product with this SKU already exists")
		}
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

// GetByIDWithTx retrieves a product within a transaction
func (r *ProductRepository) GetByIDWithTx(tx *sqlx.Tx, id uuid.UUID) (*models.Product, error) {
	product := &models.Product{}
	query := `
		SELECT id, name, sku, description, product_type_id, attributes, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	err := tx.Get(product, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("product not found")
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}

// ListWithTx retrieves products within a transaction
func (r *ProductRepository) ListWithTx(tx *sqlx.Tx, page, pageSize int, productTypeID *uuid.UUID) ([]models.Product, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	offset := (page - 1) * pageSize

	query := `
		SELECT id, name, sku, description, product_type_id, attributes, created_at, updated_at
		FROM products
	`
	countQuery := `SELECT COUNT(*) FROM products`
	
	args := []interface{}{}
	whereClause := ""

	if productTypeID != nil {
		whereClause = " WHERE product_type_id = $1"
		args = append(args, productTypeID)
	}

	query += whereClause + ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, pageSize, offset)

	var total int
	countArgs := []interface{}{}
	if productTypeID != nil {
		countArgs = append(countArgs, productTypeID)
	}
	err := tx.Get(&total, countQuery+whereClause, countArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	products := []models.Product{}
	err = tx.Select(&products, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}

	return products, total, nil
}

// UpdateWithTx updates a product within a transaction
func (r *ProductRepository) UpdateWithTx(tx *sqlx.Tx, id uuid.UUID, req *models.ProductUpdateRequest) (*models.Product, error) {
	existing, err := r.GetByIDWithTx(tx, id)
	if err != nil {
		return nil, err
	}

	updates := []string{}
	args := []interface{}{}
	argCount := 1

	if req.Name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", argCount))
		args = append(args, *req.Name)
		argCount++
	}

	if req.Description != nil {
		updates = append(updates, fmt.Sprintf("description = $%d", argCount))
		args = append(args, *req.Description)
		argCount++
	}

	if req.ProductTypeID != nil {
		updates = append(updates, fmt.Sprintf("product_type_id = $%d", argCount))
		args = append(args, *req.ProductTypeID)
		argCount++
	}

	if req.Attributes != nil {
		attrs, err := models.SanitizeAttributes(req.Attributes)
		if err != nil {
			return nil, fmt.Errorf("invalid attributes: %w", err)
		}
		updates = append(updates, fmt.Sprintf("attributes = $%d", argCount))
		args = append(args, attrs)
		argCount++
	}

	if len(updates) == 0 {
		return existing, nil
	}

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE products
		SET %s, updated_at = NOW()
		WHERE id = $%d
		RETURNING id, name, sku, description, product_type_id, attributes, created_at, updated_at
	`, join(updates, ", "), argCount)

	product := &models.Product{}
	err = tx.QueryRowx(query, args...).StructScan(product)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

// DeleteWithTx deletes a product within a transaction
func (r *ProductRepository) DeleteWithTx(tx *sqlx.Tx, id uuid.UUID) error {
	result, err := tx.Exec("DELETE FROM products WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return errors.New("product not found")
	}

	return nil
}

