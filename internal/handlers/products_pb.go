package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"apidemo1/internal/models"
	"apidemo1/internal/validator"
	pb "apidemo1/proto"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// CreateProductHandler_Proto handles POST /api/v1/products using protobuf
func CreateProductHandler_Proto(tx *sqlx.Tx) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request body exists
		if r.Body == nil || r.Body == http.NoBody {
			writeProtoError(w, http.StatusBadRequest, "bad_request", "Request body is required")
			return
		}

		// Parse protobuf request
		var req pb.ProductCreateRequest
		body := make([]byte, 0)
		if r.Body != nil {
			var err error
			body, err = readAll(r.Body)
			if err != nil {
				writeProtoError(w, http.StatusBadRequest, "bad_request", "Failed to read request body")
				return
			}
		}

		if err := protojson.Unmarshal(body, &req); err != nil {
			writeProtoError(w, http.StatusBadRequest, "bad_request", "Invalid JSON format")
			return
		}

		// Validate request using protobuf validator
		if err := models.ValidateProductCreateRequest(&req); err != nil {
			writeProtoError(w, http.StatusBadRequest, "bad_request", err.Error())
			return
		}

		// Sanitize text fields
		req.Name = validator.SanitizeText(req.Name)
		req.Description = validator.SanitizeText(req.Description)

		// Validate SKU format
		if err := validator.ValidateSKU(req.Sku); err != nil {
			writeProtoError(w, http.StatusBadRequest, "bad_request", err.Error())
			return
		}

		// Convert protobuf attributes to JSONB
		attrs, err := models.ProtoAttributesToJSON(req.Attributes)
		if err != nil {
			writeProtoError(w, http.StatusBadRequest, "bad_request", fmt.Sprintf("invalid attributes: %v", err))
			return
		}

		// Create product in database
		product := &models.Product{
			ID:          uuid.New(),
			Name:        req.Name,
			SKU:         req.Sku,
			Attributes:  attrs,
		}

		if req.Description != "" {
			product.Description = &req.Description
		}

		if req.ProductTypeId != "" {
			ptID, err := uuid.Parse(req.ProductTypeId)
			if err != nil {
				writeProtoError(w, http.StatusBadRequest, "bad_request", "Invalid product type ID")
				return
			}
			product.ProductTypeID = &ptID
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
			if pqErr, ok := err.(*pq.Error); ok {
				if pqErr.Code == "23505" { // unique_violation
					writeProtoError(w, http.StatusConflict, "conflict", "Product with this SKU already exists")
					return
				}
			}
			// Also check error message string
			errMsg := strings.ToLower(err.Error())
			if strings.Contains(errMsg, "duplicate") || strings.Contains(errMsg, "unique") {
				writeProtoError(w, http.StatusConflict, "conflict", "Product with this SKU already exists")
				return
			}
			writeProtoError(w, http.StatusInternalServerError, "internal_error", "Failed to create product")
			return
		}

		// Convert to protobuf for response
		pbProduct, err := models.ProductToProto(product)
		if err != nil {
			writeProtoError(w, http.StatusInternalServerError, "internal_error", "Failed to convert product")
			return
		}

		// Write protobuf response
		writeProtoResponse(w, http.StatusCreated, pbProduct)
	})
}

// Helper functions

func writeProtoError(w http.ResponseWriter, status int, code, message string) {
	errorResp := &pb.ErrorResponse{
		Error: &pb.ErrorDetail{
			Code:    code,
			Message: message,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	jsonData, _ := protojson.Marshal(errorResp)
	w.Write(jsonData)
}

func writeProtoResponse(w http.ResponseWriter, status int, msg proto.Message) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	jsonData, err := protojson.Marshal(msg)
	if err != nil {
		// Fallback to error response
		writeProtoError(w, http.StatusInternalServerError, "internal_error", "Failed to marshal response")
		return
	}

	w.Write(jsonData)
}

func readAll(r io.Reader) ([]byte, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r)
	return buf.Bytes(), err
}

// Helper functions for database operations (moved from old products.go)

func getProductByIDInTx(tx *sqlx.Tx, id uuid.UUID) (*models.Product, error) {
	product := &models.Product{}
	query := `
		SELECT id, name, sku, description, product_type_id, attributes, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	err := tx.Get(product, query, id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, fmt.Errorf("product not found")
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
	`, joinStrings(updates, ", "), argCount)

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
		return fmt.Errorf("product not found")
	}

	return nil
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// GetProductHandler_Proto handles GET /api/v1/products/{id} using protobuf
func GetProductHandler_Proto(tx *sqlx.Tx) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from URL path
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		idStr := parts[len(parts)-1]

		id, err := uuid.Parse(idStr)
		if err != nil {
			writeProtoError(w, http.StatusBadRequest, "bad_request", "Invalid product ID format")
			return
		}

		// Get product from database
		product, err := getProductByIDInTx(tx, id)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				writeProtoError(w, http.StatusNotFound, "not_found", "Product not found")
				return
			}
			writeProtoError(w, http.StatusInternalServerError, "internal_error", "Failed to get product")
			return
		}

		// Convert to protobuf
		pbProduct, err := models.ProductToProto(product)
		if err != nil {
			writeProtoError(w, http.StatusInternalServerError, "internal_error", "Failed to convert product")
			return
		}

		writeProtoResponse(w, http.StatusOK, pbProduct)
	})
}

// ListProductsHandler_Proto handles GET /api/v1/products using protobuf
func ListProductsHandler_Proto(tx *sqlx.Tx) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse pagination (reuse existing logic)
		page, pageSize := 1, 50
		// Simplified for now - full implementation would parse query params
		
		products, total, err := listProductsInTx(tx, page, pageSize, nil)
		if err != nil {
			writeProtoError(w, http.StatusInternalServerError, "internal_error", "Failed to list products")
			return
		}

		// Convert to protobuf
		pbProducts := make([]*pb.Product, len(products))
		for i, p := range products {
			pbProd, err := models.ProductToProto(&p)
			if err != nil {
				writeProtoError(w, http.StatusInternalServerError, "internal_error", "Failed to convert product")
				return
			}
			pbProducts[i] = pbProd
		}

		totalPages := (total + pageSize - 1) / pageSize
		response := &pb.ProductListResponse{
			Products: pbProducts,
			Pagination: &pb.Pagination{
				Page:       int32(page),
				PageSize:   int32(pageSize),
				TotalItems: int32(total),
				TotalPages: int32(totalPages),
			},
		}

		writeProtoResponse(w, http.StatusOK, response)
	})
}

// UpdateProductHandler_Proto handles PUT /api/v1/products/{id} using protobuf
func UpdateProductHandler_Proto(tx *sqlx.Tx) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from URL
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		idStr := parts[len(parts)-1]

		id, err := uuid.Parse(idStr)
		if err != nil {
			writeProtoError(w, http.StatusBadRequest, "bad_request", "Invalid product ID format")
			return
		}

		// Parse request
		body, _ := readAll(r.Body)
		var req pb.ProductUpdateRequest
		if err := protojson.Unmarshal(body, &req); err != nil {
			writeProtoError(w, http.StatusBadRequest, "bad_request", "Invalid JSON format")
			return
		}

		// Validate
		if err := models.ValidateProductUpdateRequest(&req); err != nil {
			writeProtoError(w, http.StatusBadRequest, "bad_request", err.Error())
			return
		}

		// Convert request to internal update format
		updateReq := &models.ProductUpdateRequest{}
		if req.Name != "" {
			sanitized := validator.SanitizeText(req.Name)
			updateReq.Name = &sanitized
		}
		if req.Description != "" {
			sanitized := validator.SanitizeText(req.Description)
			updateReq.Description = &sanitized
		}
		if req.ProductTypeId != "" {
			ptID, _ := uuid.Parse(req.ProductTypeId)
			updateReq.ProductTypeID = &ptID
		}
		if len(req.Attributes) > 0 {
			attrs, _ := models.ProtoAttributesToJSON(req.Attributes)
			jsonMap := make(map[string]interface{})
			_ = json.Unmarshal(attrs, &jsonMap)
			updateReq.Attributes = jsonMap
		}

		// Update product
		product, err := updateProductInTx(tx, id, updateReq)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				writeProtoError(w, http.StatusNotFound, "not_found", "Product not found")
				return
			}
			writeProtoError(w, http.StatusInternalServerError, "internal_error", "Failed to update product")
			return
		}

		// Convert to protobuf
		pbProduct, _ := models.ProductToProto(product)
		writeProtoResponse(w, http.StatusOK, pbProduct)
	})
}

// DeleteProductHandler_Proto handles DELETE /api/v1/products/{id} using protobuf
func DeleteProductHandler_Proto(tx *sqlx.Tx) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		idStr := parts[len(parts)-1]

		id, err := uuid.Parse(idStr)
		if err != nil {
			writeProtoError(w, http.StatusBadRequest, "bad_request", "Invalid product ID format")
			return
		}

		err = deleteProductInTx(tx, id)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				writeProtoError(w, http.StatusNotFound, "not_found", "Product not found")
				return
			}
			writeProtoError(w, http.StatusInternalServerError, "internal_error", "Failed to delete product")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

