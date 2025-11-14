package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"apidemo1/internal/models"
	"apidemo1/internal/repository"
	"apidemo1/internal/validator"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// ProductHandler handles product-related HTTP requests
type ProductHandler struct {
	repo *repository.ProductRepository
}

// NewProductHandler creates a new product handler
func NewProductHandler(db *sqlx.DB) *ProductHandler {
	return &ProductHandler{
		repo: repository.NewProductRepository(db),
	}
}

// CreateProduct handles POST /api/v1/products
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req models.ProductCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.BadRequest("Invalid JSON format").WriteJSON(w, http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validator.ValidateProductCreate(&req); err != nil {
		if validationErr, ok := err.(models.ValidationErr); ok {
			models.BadRequest(validationErr.Error()).WriteJSON(w, http.StatusBadRequest)
			return
		}
		models.BadRequest(err.Error()).WriteJSON(w, http.StatusBadRequest)
		return
	}

	// Create product
	product, err := h.repo.Create(&req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			models.Conflict("Product with this SKU already exists").WriteJSON(w, http.StatusConflict)
			return
		}
		models.InternalError("Failed to create product").WriteJSON(w, http.StatusInternalServerError)
		return
	}

	// Return created product
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

// GetProduct handles GET /api/v1/products/{id}
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	// Get ID from URL
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		models.BadRequest("Product ID is required").WriteJSON(w, http.StatusBadRequest)
		return
	}

	// Parse UUID
	id, err := uuid.Parse(idStr)
	if err != nil {
		models.BadRequest("Invalid product ID format").WriteJSON(w, http.StatusBadRequest)
		return
	}

	// Get product
	product, err := h.repo.GetByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			models.NotFound("Product not found").WriteJSON(w, http.StatusNotFound)
			return
		}
		models.InternalError("Failed to get product").WriteJSON(w, http.StatusInternalServerError)
		return
	}

	// Return product
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(product)
}

// ListProducts handles GET /api/v1/products
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")
	productTypeIDStr := r.URL.Query().Get("product_type_id")

	page := 1
	pageSize := 50

	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil {
			models.BadRequest("Invalid page parameter").WriteJSON(w, http.StatusBadRequest)
			return
		}
		page = p
	}

	if pageSizeStr != "" {
		ps, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			models.BadRequest("Invalid page_size parameter").WriteJSON(w, http.StatusBadRequest)
			return
		}
		pageSize = ps
	}

	// Validate pagination
	page, pageSize, err := validator.ValidatePagination(page, pageSize)
	if err != nil {
		if validationErr, ok := err.(models.ValidationErr); ok {
			models.BadRequest(validationErr.Error()).WriteJSON(w, http.StatusBadRequest)
			return
		}
		models.BadRequest(err.Error()).WriteJSON(w, http.StatusBadRequest)
		return
	}

	// Parse product type filter
	var productTypeID *uuid.UUID
	if productTypeIDStr != "" {
		id, err := uuid.Parse(productTypeIDStr)
		if err != nil {
			models.BadRequest("Invalid product_type_id format").WriteJSON(w, http.StatusBadRequest)
			return
		}
		productTypeID = &id
	}

	// Get products
	products, total, err := h.repo.List(page, pageSize, productTypeID)
	if err != nil {
		models.InternalError("Failed to list products").WriteJSON(w, http.StatusInternalServerError)
		return
	}

	// Calculate pagination metadata
	totalPages := (total + pageSize - 1) / pageSize

	response := models.ProductListResponse{
		Products: products,
		Pagination: models.Pagination{
			Page:       page,
			PageSize:   pageSize,
			TotalItems: total,
			TotalPages: totalPages,
		},
	}

	// Return products
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateProduct handles PUT /api/v1/products/{id}
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	// Get ID from URL
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		models.BadRequest("Product ID is required").WriteJSON(w, http.StatusBadRequest)
		return
	}

	// Parse UUID
	id, err := uuid.Parse(idStr)
	if err != nil {
		models.BadRequest("Invalid product ID format").WriteJSON(w, http.StatusBadRequest)
		return
	}

	// Parse request body
	var req models.ProductUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.BadRequest("Invalid JSON format").WriteJSON(w, http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validator.ValidateProductUpdate(&req); err != nil {
		if validationErr, ok := err.(models.ValidationErr); ok {
			models.BadRequest(validationErr.Error()).WriteJSON(w, http.StatusBadRequest)
			return
		}
		models.BadRequest(err.Error()).WriteJSON(w, http.StatusBadRequest)
		return
	}

	// Update product
	product, err := h.repo.Update(id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			models.NotFound("Product not found").WriteJSON(w, http.StatusNotFound)
			return
		}
		models.InternalError("Failed to update product").WriteJSON(w, http.StatusInternalServerError)
		return
	}

	// Return updated product
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(product)
}

// DeleteProduct handles DELETE /api/v1/products/{id}
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	// Get ID from URL
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		models.BadRequest("Product ID is required").WriteJSON(w, http.StatusBadRequest)
		return
	}

	// Parse UUID
	id, err := uuid.Parse(idStr)
	if err != nil {
		models.BadRequest("Invalid product ID format").WriteJSON(w, http.StatusBadRequest)
		return
	}

	// Delete product
	err = h.repo.Delete(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			models.NotFound("Product not found").WriteJSON(w, http.StatusNotFound)
			return
		}
		models.InternalError("Failed to delete product").WriteJSON(w, http.StatusInternalServerError)
		return
	}

	// Return 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// Helper functions for test compatibility

// CreateProductHandler returns a handler for tests
func CreateProductHandler(db interface{}) http.Handler {
	tx, ok := db.(*sqlx.Tx)
	if !ok {
		// Not a transaction, return error handler
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			models.InternalError("Invalid database connection").WriteJSON(w, http.StatusInternalServerError)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request body exists
		if r.Body == nil || r.Body == http.NoBody {
			models.BadRequest("Request body is required").WriteJSON(w, http.StatusBadRequest)
			return
		}
		
		var req models.ProductCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			models.BadRequest("Invalid JSON format").WriteJSON(w, http.StatusBadRequest)
			return
		}

		if err := validator.ValidateProductCreate(&req); err != nil {
			if validationErr, ok := err.(models.ValidationErr); ok {
				models.BadRequest(validationErr.Error()).WriteJSON(w, http.StatusBadRequest)
				return
			}
			models.BadRequest(err.Error()).WriteJSON(w, http.StatusBadRequest)
			return
		}

		// Create product directly in transaction
		product, err := createProductInTx(tx, &req)
		if err != nil {
			errMsg := strings.ToLower(err.Error())
			if strings.Contains(errMsg, "already exists") || strings.Contains(errMsg, "duplicate") {
				models.Conflict("Product with this SKU already exists").WriteJSON(w, http.StatusConflict)
				return
			}
			models.InternalError("Failed to create product").WriteJSON(w, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(product)
	})
}

// GetProductHandler returns a handler for tests
func GetProductHandler(db interface{}) http.Handler {
	tx, ok := db.(*sqlx.Tx)
	if !ok {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			models.InternalError("Invalid database connection").WriteJSON(w, http.StatusInternalServerError)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from URL path
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		idStr := parts[len(parts)-1]

		id, err := uuid.Parse(idStr)
		if err != nil {
			models.BadRequest("Invalid product ID format").WriteJSON(w, http.StatusBadRequest)
			return
		}

		product, err := getProductByIDInTx(tx, id)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				models.NotFound("Product not found").WriteJSON(w, http.StatusNotFound)
				return
			}
			models.InternalError("Failed to get product").WriteJSON(w, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(product)
	})
}

// ListProductsHandler returns a handler for tests
func ListProductsHandler(db interface{}) http.Handler {
	tx, ok := db.(*sqlx.Tx)
	if !ok {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			models.InternalError("Invalid database connection").WriteJSON(w, http.StatusInternalServerError)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := 1
		pageSize := 50

		pageStr := r.URL.Query().Get("page")
		if pageStr != "" {
			p, err := strconv.Atoi(pageStr)
			if err != nil {
				models.BadRequest("Invalid page parameter").WriteJSON(w, http.StatusBadRequest)
				return
			}
			page = p
		}

		pageSizeStr := r.URL.Query().Get("page_size")
		if pageSizeStr != "" {
			ps, err := strconv.Atoi(pageSizeStr)
			if err != nil {
				models.BadRequest("Invalid page_size parameter").WriteJSON(w, http.StatusBadRequest)
				return
			}
			pageSize = ps
		}

		page, pageSize, err := validator.ValidatePagination(page, pageSize)
		if err != nil {
			if validationErr, ok := err.(models.ValidationErr); ok {
				models.BadRequest(validationErr.Error()).WriteJSON(w, http.StatusBadRequest)
				return
			}
			models.BadRequest(err.Error()).WriteJSON(w, http.StatusBadRequest)
			return
		}

		products, total, err := listProductsInTx(tx, page, pageSize, nil)
		if err != nil {
			models.InternalError("Failed to list products").WriteJSON(w, http.StatusInternalServerError)
			return
		}

		totalPages := (total + pageSize - 1) / pageSize
		response := models.ProductListResponse{
			Products: products,
			Pagination: models.Pagination{
				Page:       page,
				PageSize:   pageSize,
				TotalItems: total,
				TotalPages: totalPages,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})
}

// UpdateProductHandler returns a handler for tests
func UpdateProductHandler(db interface{}) http.Handler {
	tx, ok := db.(*sqlx.Tx)
	if !ok {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			models.InternalError("Invalid database connection").WriteJSON(w, http.StatusInternalServerError)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		idStr := parts[len(parts)-1]

		id, err := uuid.Parse(idStr)
		if err != nil {
			models.BadRequest("Invalid product ID format").WriteJSON(w, http.StatusBadRequest)
			return
		}

		var req models.ProductUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			models.BadRequest("Invalid JSON format").WriteJSON(w, http.StatusBadRequest)
			return
		}

		if err := validator.ValidateProductUpdate(&req); err != nil {
			if validationErr, ok := err.(models.ValidationErr); ok {
				models.BadRequest(validationErr.Error()).WriteJSON(w, http.StatusBadRequest)
				return
			}
			models.BadRequest(err.Error()).WriteJSON(w, http.StatusBadRequest)
			return
		}

		product, err := updateProductInTx(tx, id, &req)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				models.NotFound("Product not found").WriteJSON(w, http.StatusNotFound)
				return
			}
			models.InternalError("Failed to update product").WriteJSON(w, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(product)
	})
}

// DeleteProductHandler returns a handler for tests
func DeleteProductHandler(db interface{}) http.Handler {
	tx, ok := db.(*sqlx.Tx)
	if !ok {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			models.InternalError("Invalid database connection").WriteJSON(w, http.StatusInternalServerError)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		idStr := parts[len(parts)-1]

		id, err := uuid.Parse(idStr)
		if err != nil {
			models.BadRequest("Invalid product ID format").WriteJSON(w, http.StatusBadRequest)
			return
		}

		err = deleteProductInTx(tx, id)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				models.NotFound("Product not found").WriteJSON(w, http.StatusNotFound)
				return
			}
			models.InternalError("Failed to delete product").WriteJSON(w, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

// Helper functions for transaction-based operations

func createProductInTx(tx *sqlx.Tx, req *models.ProductCreateRequest) (*models.Product, error) {
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
		// Check for unique constraint violation (duplicate SKU)
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				return nil, errors.New("already exists")
			}
		}
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

func getProductByIDInTx(tx *sqlx.Tx, id uuid.UUID) (*models.Product, error) {
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

func listProductsInTx(tx *sqlx.Tx, page, pageSize int, productTypeID *uuid.UUID) ([]models.Product, int, error) {
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

func updateProductInTx(tx *sqlx.Tx, id uuid.UUID, req *models.ProductUpdateRequest) (*models.Product, error) {
	existing, err := getProductByIDInTx(tx, id)
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
	`, strings.Join(updates, ", "), argCount)

	product := &models.Product{}
	err = tx.QueryRowx(query, args...).StructScan(product)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

func deleteProductInTx(tx *sqlx.Tx, id uuid.UUID) error {
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

