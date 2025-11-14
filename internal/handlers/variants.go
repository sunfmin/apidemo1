package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"apidemo1/internal/models"
	"apidemo1/internal/validator"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// CreateVariantHandler handles POST /api/v1/products/{productId}/variants
func CreateVariantHandler(db interface{}, productIDStr string) http.Handler {
	tx, ok := db.(*sqlx.Tx)
	if !ok {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			models.InternalError("Invalid database connection").WriteJSON(w, http.StatusInternalServerError)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse product ID
		productID, err := uuid.Parse(productIDStr)
		if err != nil {
			models.BadRequest("Invalid product ID format").WriteJSON(w, http.StatusBadRequest)
			return
		}

		// Check if request body exists
		if r.Body == nil || r.Body == http.NoBody {
			models.BadRequest("Request body is required").WriteJSON(w, http.StatusBadRequest)
			return
		}

		// Parse request body
		var req models.VariantCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			models.BadRequest("Invalid JSON format").WriteJSON(w, http.StatusBadRequest)
			return
		}

		// Validate request
		if err := validator.ValidateVariantCreate(&req); err != nil {
			if validationErr, ok := err.(models.ValidationErr); ok {
				models.BadRequest(validationErr.Error()).WriteJSON(w, http.StatusBadRequest)
				return
			}
			models.BadRequest(err.Error()).WriteJSON(w, http.StatusBadRequest)
			return
		}

		// Create variant
		variant, err := createVariantInTx(tx, productID, &req)
		if err != nil {
			errMsg := strings.ToLower(err.Error())
			if strings.Contains(errMsg, "already exists") || strings.Contains(errMsg, "duplicate") {
				models.Conflict("Variant with this SKU already exists").WriteJSON(w, http.StatusConflict)
				return
			}
			if strings.Contains(errMsg, "not found") {
				models.NotFound("Parent product not found").WriteJSON(w, http.StatusNotFound)
				return
			}
			models.InternalError("Failed to create variant").WriteJSON(w, http.StatusInternalServerError)
			return
		}

		// Return created variant
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(variant)
	})
}

// ListVariantsHandler handles GET /api/v1/products/{productId}/variants
func ListVariantsHandler(db interface{}, productIDStr string) http.Handler {
	tx, ok := db.(*sqlx.Tx)
	if !ok {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			models.InternalError("Invalid database connection").WriteJSON(w, http.StatusInternalServerError)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse product ID
		productID, err := uuid.Parse(productIDStr)
		if err != nil {
			models.BadRequest("Invalid product ID format").WriteJSON(w, http.StatusBadRequest)
			return
		}

		// List variants
		variants, err := listVariantsByProductIDInTx(tx, productID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				models.NotFound("Parent product not found").WriteJSON(w, http.StatusNotFound)
				return
			}
			models.InternalError("Failed to list variants").WriteJSON(w, http.StatusInternalServerError)
			return
		}

		// Return variants
		response := models.VariantListResponse{
			Variants: variants,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})
}

// GetVariantHandler handles GET /api/v1/variants/{id}
func GetVariantHandler(db interface{}) http.Handler {
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
			models.BadRequest("Invalid variant ID format").WriteJSON(w, http.StatusBadRequest)
			return
		}

		// Get variant
		variant, err := getVariantByIDInTx(tx, id)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				models.NotFound("Variant not found").WriteJSON(w, http.StatusNotFound)
				return
			}
			models.InternalError("Failed to get variant").WriteJSON(w, http.StatusInternalServerError)
			return
		}

		// Return variant
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(variant)
	})
}

// UpdateVariantHandler handles PUT /api/v1/variants/{id}
func UpdateVariantHandler(db interface{}) http.Handler {
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
			models.BadRequest("Invalid variant ID format").WriteJSON(w, http.StatusBadRequest)
			return
		}

		// Parse request body
		var req models.VariantUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			models.BadRequest("Invalid JSON format").WriteJSON(w, http.StatusBadRequest)
			return
		}

		// Validate request
		if err := validator.ValidateVariantUpdate(&req); err != nil {
			if validationErr, ok := err.(models.ValidationErr); ok {
				models.BadRequest(validationErr.Error()).WriteJSON(w, http.StatusBadRequest)
				return
			}
			models.BadRequest(err.Error()).WriteJSON(w, http.StatusBadRequest)
			return
		}

		// Update variant
		variant, err := updateVariantInTx(tx, id, &req)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				models.NotFound("Variant not found").WriteJSON(w, http.StatusNotFound)
				return
			}
			models.InternalError("Failed to update variant").WriteJSON(w, http.StatusInternalServerError)
			return
		}

		// Return updated variant
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(variant)
	})
}

// DeleteVariantHandler handles DELETE /api/v1/variants/{id}
func DeleteVariantHandler(db interface{}) http.Handler {
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
			models.BadRequest("Invalid variant ID format").WriteJSON(w, http.StatusBadRequest)
			return
		}

		// Delete variant
		err = deleteVariantInTx(tx, id)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				models.NotFound("Variant not found").WriteJSON(w, http.StatusNotFound)
				return
			}
			models.InternalError("Failed to delete variant").WriteJSON(w, http.StatusInternalServerError)
			return
		}

		// Return 204 No Content
		w.WriteHeader(http.StatusNoContent)
	})
}

// Helper functions for transaction-based operations

func createVariantInTx(tx *sqlx.Tx, productID uuid.UUID, req *models.VariantCreateRequest) (*models.Variant, error) {
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
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, errors.New("already exists")
		}
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
			return nil, errors.New("parent product not found")
		}
		return nil, fmt.Errorf("failed to create variant: %w", err)
	}

	return variant, nil
}

func getVariantByIDInTx(tx *sqlx.Tx, id uuid.UUID) (*models.Variant, error) {
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

func listVariantsByProductIDInTx(tx *sqlx.Tx, productID uuid.UUID) ([]models.Variant, error) {
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

func updateVariantInTx(tx *sqlx.Tx, id uuid.UUID, req *models.VariantUpdateRequest) (*models.Variant, error) {
	// Check if variant exists
	existing, err := getVariantByIDInTx(tx, id)
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

func deleteVariantInTx(tx *sqlx.Tx, id uuid.UUID) error {
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

