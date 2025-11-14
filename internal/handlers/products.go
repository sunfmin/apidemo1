package handlers

import (
	"net/http"

	"github.com/jmoiron/sqlx"
)

// ProductHandler handles product-related HTTP requests using protobuf
type ProductHandler struct {
	db *sqlx.DB
}

// NewProductHandler creates a new product handler
func NewProductHandler(db *sqlx.DB) *ProductHandler {
	return &ProductHandler{db: db}
}

// CreateProduct handles POST /api/v1/products using protobuf
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	// For write operations, use DB directly (auto-commit per operation)
	// In production, consider proper transaction management
	tx, err := h.db.Beginx()
	if err != nil {
		writeProtoError(w, http.StatusInternalServerError, "internal_error", "Failed to start transaction")
		return
	}

	// Create custom writer to capture status
	rw := &statusRecorder{ResponseWriter: w, status: 200}
	
	CreateProductHandler_Proto(tx).ServeHTTP(rw, r)

	// Commit on success, rollback on error
	if rw.status < 400 {
		tx.Commit()
	} else {
		tx.Rollback()
	}
}

// GetProduct handles GET /api/v1/products/{id} using protobuf
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	// Read-only, use transaction (will be rolled back)
	tx, _ := h.db.Beginx()
	defer tx.Rollback()
	
	GetProductHandler_Proto(tx).ServeHTTP(w, r)
}

// ListProducts handles GET /api/v1/products using protobuf
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	tx, _ := h.db.Beginx()
	defer tx.Rollback()
	
	ListProductsHandler_Proto(tx).ServeHTTP(w, r)
}

// UpdateProduct handles PUT /api/v1/products/{id} using protobuf
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	tx, err := h.db.Beginx()
	if err != nil {
		writeProtoError(w, http.StatusInternalServerError, "internal_error", "Failed to start transaction")
		return
	}

	rw := &statusRecorder{ResponseWriter: w, status: 200}
	UpdateProductHandler_Proto(tx).ServeHTTP(rw, r)

	if rw.status < 400 {
		tx.Commit()
	} else {
		tx.Rollback()
	}
}

// DeleteProduct handles DELETE /api/v1/products/{id} using protobuf
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	tx, err := h.db.Beginx()
	if err != nil {
		writeProtoError(w, http.StatusInternalServerError, "internal_error", "Failed to start transaction")
		return
	}

	rw := &statusRecorder{ResponseWriter: w, status: 200}
	DeleteProductHandler_Proto(tx).ServeHTTP(rw, r)

	if rw.status < 400 {
		tx.Commit()
	} else {
		tx.Rollback()
	}
}

// statusRecorder records HTTP status codes
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

