package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"apidemo1/internal/models"
	"apidemo1/internal/validator"
	pb "apidemo1/proto"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"google.golang.org/protobuf/encoding/protojson"
)

// CreateVariantHandler_Proto handles POST /api/v1/products/{productId}/variants using protobuf
func CreateVariantHandler_Proto(tx *sqlx.Tx, productIDStr string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse product ID
		productID, err := uuid.Parse(productIDStr)
		if err != nil {
			writeProtoError(w, http.StatusBadRequest, "bad_request", "Invalid product ID format")
			return
		}

		// Check if request body exists
		if r.Body == nil || r.Body == http.NoBody {
			writeProtoError(w, http.StatusBadRequest, "bad_request", "Request body is required")
			return
		}

		// Parse protobuf request
		body, _ := readAll(r.Body)
		var req pb.VariantCreateRequest
		if err := protojson.Unmarshal(body, &req); err != nil {
			writeProtoError(w, http.StatusBadRequest, "bad_request", "Invalid JSON format")
			return
		}

		// Validate request
		if err := models.ValidateVariantCreateRequest(&req); err != nil {
			writeProtoError(w, http.StatusBadRequest, "bad_request", err.Error())
			return
		}

		// Validate SKU format
		if err := validator.ValidateSKU(req.Sku); err != nil {
			writeProtoError(w, http.StatusBadRequest, "bad_request", err.Error())
			return
		}

		// Check if parent product exists
		var exists bool
		err = tx.Get(&exists, "SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)", productID)
		if err != nil || !exists {
			writeProtoError(w, http.StatusNotFound, "not_found", "Parent product not found")
			return
		}

		// Convert attributes to JSONB
		attrs, err := models.ProtoAttributesToJSON(req.Attributes)
		if err != nil {
			writeProtoError(w, http.StatusBadRequest, "bad_request", fmt.Sprintf("invalid attributes: %v", err))
			return
		}

		// Create variant
		variant := &models.Variant{
			ID:         uuid.New(),
			ProductID:  productID,
			SKU:        req.Sku,
			Attributes: attrs,
		}

		query := `
			INSERT INTO variants (id, product_id, sku, attributes)
			VALUES ($1, $2, $3, $4)
			RETURNING id, product_id, sku, attributes, created_at, updated_at
		`

		err = tx.QueryRowx(query, variant.ID, variant.ProductID, variant.SKU, variant.Attributes).StructScan(variant)

		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				writeProtoError(w, http.StatusConflict, "conflict", "Variant with this SKU already exists")
				return
			}
			errMsg := strings.ToLower(err.Error())
			if strings.Contains(errMsg, "duplicate") || strings.Contains(errMsg, "unique") {
				writeProtoError(w, http.StatusConflict, "conflict", "Variant with this SKU already exists")
				return
			}
			writeProtoError(w, http.StatusInternalServerError, "internal_error", "Failed to create variant")
			return
		}

		// Convert to protobuf
		pbVariant, err := models.VariantToProto(variant)
		if err != nil {
			writeProtoError(w, http.StatusInternalServerError, "internal_error", "Failed to convert variant")
			return
		}

		writeProtoResponse(w, http.StatusCreated, pbVariant)
	})
}

// ListVariantsHandler_Proto handles GET /api/v1/products/{productId}/variants using protobuf
func ListVariantsHandler_Proto(tx *sqlx.Tx, productIDStr string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse product ID
		productID, err := uuid.Parse(productIDStr)
		if err != nil {
			writeProtoError(w, http.StatusBadRequest, "bad_request", "Invalid product ID format")
			return
		}

		// List variants
		variants, err := listVariantsByProductIDInTx(tx, productID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				writeProtoError(w, http.StatusNotFound, "not_found", "Parent product not found")
				return
			}
			writeProtoError(w, http.StatusInternalServerError, "internal_error", "Failed to list variants")
			return
		}

		// Convert to protobuf
		pbVariants := make([]*pb.Variant, len(variants))
		for i, v := range variants {
			pbVar, err := models.VariantToProto(&v)
			if err != nil {
				writeProtoError(w, http.StatusInternalServerError, "internal_error", "Failed to convert variant")
				return
			}
			pbVariants[i] = pbVar
		}

		response := &pb.VariantListResponse{
			Variants: pbVariants,
		}

		writeProtoResponse(w, http.StatusOK, response)
	})
}

// GetVariantHandler_Proto handles GET /api/v1/variants/{id} using protobuf
func GetVariantHandler_Proto(tx *sqlx.Tx) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from URL
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		idStr := parts[len(parts)-1]

		id, err := uuid.Parse(idStr)
		if err != nil {
			writeProtoError(w, http.StatusBadRequest, "bad_request", "Invalid variant ID format")
			return
		}

		// Get variant
		variant, err := getVariantByIDInTx(tx, id)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				writeProtoError(w, http.StatusNotFound, "not_found", "Variant not found")
				return
			}
			writeProtoError(w, http.StatusInternalServerError, "internal_error", "Failed to get variant")
			return
		}

		// Convert to protobuf
		pbVariant, err := models.VariantToProto(variant)
		if err != nil {
			writeProtoError(w, http.StatusInternalServerError, "internal_error", "Failed to convert variant")
			return
		}

		writeProtoResponse(w, http.StatusOK, pbVariant)
	})
}

// UpdateVariantHandler_Proto handles PUT /api/v1/variants/{id} using protobuf
func UpdateVariantHandler_Proto(tx *sqlx.Tx) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from URL
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		idStr := parts[len(parts)-1]

		id, err := uuid.Parse(idStr)
		if err != nil {
			writeProtoError(w, http.StatusBadRequest, "bad_request", "Invalid variant ID format")
			return
		}

		// Parse request
		body, _ := readAll(r.Body)
		var req pb.VariantUpdateRequest
		if err := protojson.Unmarshal(body, &req); err != nil {
			writeProtoError(w, http.StatusBadRequest, "bad_request", "Invalid JSON format")
			return
		}

		// Validate
		if err := models.ValidateVariantUpdateRequest(&req); err != nil {
			writeProtoError(w, http.StatusBadRequest, "bad_request", err.Error())
			return
		}

		// Convert to internal format
		updateReq := &models.VariantUpdateRequest{}
		if len(req.Attributes) > 0 {
			attrs, _ := models.ProtoAttributesToJSON(req.Attributes)
			jsonMap := make(map[string]interface{})
			_ = json.Unmarshal(attrs, &jsonMap)
			updateReq.Attributes = jsonMap
		}

		// Update variant
		variant, err := updateVariantInTx(tx, id, updateReq)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				writeProtoError(w, http.StatusNotFound, "not_found", "Variant not found")
				return
			}
			writeProtoError(w, http.StatusInternalServerError, "internal_error", "Failed to update variant")
			return
		}

		// Convert to protobuf
		pbVariant, _ := models.VariantToProto(variant)
		writeProtoResponse(w, http.StatusOK, pbVariant)
	})
}

// DeleteVariantHandler_Proto handles DELETE /api/v1/variants/{id} using protobuf
func DeleteVariantHandler_Proto(tx *sqlx.Tx) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		idStr := parts[len(parts)-1]

		id, err := uuid.Parse(idStr)
		if err != nil {
			writeProtoError(w, http.StatusBadRequest, "bad_request", "Invalid variant ID format")
			return
		}

		err = deleteVariantInTx(tx, id)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				writeProtoError(w, http.StatusNotFound, "not_found", "Variant not found")
				return
			}
			writeProtoError(w, http.StatusInternalServerError, "internal_error", "Failed to delete variant")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

// Helper functions for database operations

func getVariantByIDInTx(tx *sqlx.Tx, id uuid.UUID) (*models.Variant, error) {
	variant := &models.Variant{}
	query := `
		SELECT id, product_id, sku, attributes, created_at, updated_at
		FROM variants
		WHERE id = $1
	`

	err := tx.Get(variant, query, id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, fmt.Errorf("variant not found")
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
		return nil, fmt.Errorf("parent product not found")
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
		return fmt.Errorf("variant not found")
	}

	return nil
}

