package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"apidemo1/internal/models"
	"apidemo1/internal/testutil"
	pb "apidemo1/proto"

	"github.com/google/uuid"
)

// TestCreateVariant tests POST /api/v1/products/{productId}/variants using protobuf
func TestCreateVariant(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	tests := []struct {
		name           string
		setupProduct   bool
		productID      string
		request        *pb.VariantCreateRequest
		expectedStatus int
		expectedError  *string
		checkResponse  func(t *testing.T, variant *pb.Variant)
	}{
		// ========== HAPPY PATH ==========
		{
			name:         "create variant with SKU and attributes",
			setupProduct: true,
			request: &pb.VariantCreateRequest{
				Sku: "TSHIRT-BLUE-M-PROTO",
				Attributes: map[string]*pb.AttributeValue{
					"size":  models.CreateStringAttribute("M"),
					"color": models.CreateStringAttribute("blue"),
					"price": models.CreateNumberAttribute(29.99),
				},
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, variant *pb.Variant) {
				if variant.Id == "" {
					t.Error("expected id to be present")
				}
				if variant.ProductId == "" {
					t.Error("expected product_id to be present")
				}
				if variant.Sku != "TSHIRT-BLUE-M-PROTO" {
					t.Errorf("expected sku, got %v", variant.Sku)
				}
				if len(variant.Attributes) != 3 {
					t.Errorf("expected 3 attributes, got %d", len(variant.Attributes))
				}
			},
		},
		{
			name:         "create variant with minimal fields",
			setupProduct: true,
			request: &pb.VariantCreateRequest{
				Sku: "MIN-VARIANT-PROTO-001",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, variant *pb.Variant) {
				if len(variant.Attributes) != 0 {
					t.Errorf("expected 0 attributes, got %d", len(variant.Attributes))
				}
			},
		},

		// ========== DATA STATE ==========
		{
			name:         "create variant for non-existent parent",
			setupProduct: false,
			productID:    uuid.New().String(),
			request: &pb.VariantCreateRequest{
				Sku: "NO-PARENT-PROTO-001",
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  stringPtr("not_found"),
		},

		// ========== INPUT VALIDATION ==========
		{
			name:         "empty SKU",
			setupProduct: true,
			request: &pb.VariantCreateRequest{
				Sku: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  stringPtr("bad_request"),
		},
		{
			name:           "SKU at maximum length",
			setupProduct:   true,
			request: &pb.VariantCreateRequest{
				Sku: generateString(100),
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:         "SKU exceeds maximum length",
			setupProduct: true,
			request: &pb.VariantCreateRequest{
				Sku: generateString(101),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  stringPtr("bad_request"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := testutil.BeginTestTransaction(t, db)

			var productID string
			if tt.setupProduct {
				product := testutil.CreateTestProduct(t, tx, "Parent Product", fmt.Sprintf("PARENT-PROTO-%d", len(tt.name)), nil)
				productID = product.ID.String()
			} else if tt.productID != "" {
				productID = tt.productID
			}

			req, _ := testutil.MakeRequest(http.MethodPost, "/api/v1/products/"+productID+"/variants", tt.request)
			rr := httptest.NewRecorder()

			handler := CreateVariantHandler_Proto(tx, productID)
			handler.ServeHTTP(rr, req)

			testutil.AssertStatus(t, rr, tt.expectedStatus)

			if tt.expectedStatus == http.StatusCreated {
				variant := &pb.Variant{}
				testutil.ParseProtoResponse(t, rr, variant)

				if tt.checkResponse != nil {
					tt.checkResponse(t, variant)
				}
			} else if tt.expectedError != nil {
				errorResp := &pb.ErrorResponse{}
				testutil.ParseProtoResponse(t, rr, errorResp)
				if errorResp.Error.Code != *tt.expectedError {
					t.Errorf("expected error code %s, got %v", *tt.expectedError, errorResp.Error.Code)
				}
			}
		})
	}
}

// TestCreateVariantDuplicateSKU tests duplicate SKU constraint using protobuf
func TestCreateVariantDuplicateSKU(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	tx := testutil.BeginTestTransaction(t, db)

	// Create parent product
	product := testutil.CreateTestProduct(t, tx, "Parent Product", "PARENT-DUP-PROTO-002", nil)

	// Create first variant
	variant1 := testutil.CreateTestVariant(t, tx, product.ID, "DUP-VARIANT-PROTO-001", nil)
	if variant1 == nil {
		t.Fatal("failed to create first variant")
	}

	// Try to create second variant with same SKU using protobuf
	req, _ := testutil.MakeRequest(http.MethodPost, "/api/v1/products/"+product.ID.String()+"/variants", &pb.VariantCreateRequest{
		Sku: "DUP-VARIANT-PROTO-001",
	})

	rr := httptest.NewRecorder()
	handler := CreateVariantHandler_Proto(tx, product.ID.String())
	handler.ServeHTTP(rr, req)

	testutil.AssertStatus(t, rr, http.StatusConflict)

	errorResp := &pb.ErrorResponse{}
	testutil.ParseProtoResponse(t, rr, errorResp)
	
	if errorResp.Error.Code != "conflict" {
		t.Errorf("expected error code 'conflict', got %v", errorResp.Error.Code)
	}
}

// TestListVariants tests GET /api/v1/products/{productId}/variants using protobuf
func TestListVariants(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	tests := []struct {
		name           string
		setupProduct   bool
		setupVariants  int
		productID      string
		expectedStatus int
		expectedError  *string
		checkResponse  func(t *testing.T, resp *pb.VariantListResponse)
	}{
		{
			name:           "list all variants for a product",
			setupProduct:   true,
			setupVariants:  3,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *pb.VariantListResponse) {
				if len(resp.Variants) != 3 {
					t.Errorf("expected 3 variants, got %d", len(resp.Variants))
				}
			},
		},
		{
			name:           "product with zero variants",
			setupProduct:   true,
			setupVariants:  0,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *pb.VariantListResponse) {
				if len(resp.Variants) != 0 {
					t.Errorf("expected 0 variants, got %d", len(resp.Variants))
				}
			},
		},
		{
			name:           "list variants for non-existent product",
			setupProduct:   false,
			productID:      uuid.New().String(),
			expectedStatus: http.StatusNotFound,
			expectedError:  stringPtr("not_found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := testutil.BeginTestTransaction(t, db)

			var productID string
			if tt.setupProduct {
				product := testutil.CreateTestProduct(t, tx, "Parent Product", fmt.Sprintf("PARENT-LIST-PROTO-%d", tt.setupVariants), nil)
				productID = product.ID.String()

				for i := 0; i < tt.setupVariants; i++ {
					sku := fmt.Sprintf("VARIANT-LIST-PROTO-%d-%d", tt.setupVariants, i)
					testutil.CreateTestVariant(t, tx, product.ID, sku, nil)
				}
			} else {
				productID = tt.productID
			}

			req, _ := testutil.MakeRequest(http.MethodGet, "/api/v1/products/"+productID+"/variants", nil)
			rr := httptest.NewRecorder()

			handler := ListVariantsHandler_Proto(tx, productID)
			handler.ServeHTTP(rr, req)

			testutil.AssertStatus(t, rr, tt.expectedStatus)

			if tt.expectedStatus == http.StatusOK {
				response := &pb.VariantListResponse{}
				testutil.ParseProtoResponse(t, rr, response)

				if tt.checkResponse != nil {
					tt.checkResponse(t, response)
				}
			} else if tt.expectedError != nil {
				errorResp := &pb.ErrorResponse{}
				testutil.ParseProtoResponse(t, rr, errorResp)
				if errorResp.Error.Code != *tt.expectedError {
					t.Errorf("expected error code %s, got %v", *tt.expectedError, errorResp.Error.Code)
				}
			}
		})
	}
}

// TestGetVariant tests GET /api/v1/variants/{id} using protobuf
func TestGetVariant(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	tests := []struct {
		name           string
		setupVariant   bool
		variantID      string
		expectedStatus int
		expectedError  *string
	}{
		{
			name:           "retrieve existing variant",
			setupVariant:   true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent variant ID returns 404",
			setupVariant:   false,
			variantID:      uuid.New().String(),
			expectedStatus: http.StatusNotFound,
			expectedError:  stringPtr("not_found"),
		},
		{
			name:           "invalid UUID format returns 400",
			setupVariant:   false,
			variantID:      "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  stringPtr("bad_request"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := testutil.BeginTestTransaction(t, db)

			var variantID string
			if tt.setupVariant {
				product := testutil.CreateTestProduct(t, tx, "Parent Product", "PARENT-GET-VAR-PROTO", nil)
				variant := testutil.CreateTestVariant(t, tx, product.ID, "VARIANT-GET-PROTO-001", nil)
				variantID = variant.ID.String()
			} else {
				variantID = tt.variantID
			}

			url := "/api/v1/variants/" + variantID
			req, _ := testutil.MakeRequest(http.MethodGet, url, nil)
			rr := httptest.NewRecorder()

			handler := GetVariantHandler_Proto(tx)
			handler.ServeHTTP(rr, req)

			testutil.AssertStatus(t, rr, tt.expectedStatus)

			if tt.expectedStatus == http.StatusOK {
				variant := &pb.Variant{}
				testutil.ParseProtoResponse(t, rr, variant)
				if variant.Id == "" {
					t.Error("expected variant ID to be present")
				}
			} else if tt.expectedError != nil {
				errorResp := &pb.ErrorResponse{}
				testutil.ParseProtoResponse(t, rr, errorResp)
				if errorResp.Error.Code != *tt.expectedError {
					t.Errorf("expected error code %s, got %v", *tt.expectedError, errorResp.Error.Code)
				}
			}
		})
	}
}

// TestUpdateVariant tests PUT /api/v1/variants/{id} using protobuf
func TestUpdateVariant(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	tests := []struct {
		name           string
		setupVariant   bool
		request        *pb.VariantUpdateRequest
		expectedStatus int
		expectedError  *string
	}{
		{
			name:         "update variant attributes",
			setupVariant: true,
			request: &pb.VariantUpdateRequest{
				Attributes: map[string]*pb.AttributeValue{
					"updated_attr": models.CreateStringAttribute("updated_value"),
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "update non-existent variant returns 404",
			setupVariant: false,
			request: &pb.VariantUpdateRequest{
				Attributes: map[string]*pb.AttributeValue{},
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  stringPtr("not_found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := testutil.BeginTestTransaction(t, db)

			var variantID string
			if tt.setupVariant {
				product := testutil.CreateTestProduct(t, tx, "Parent Product", "PARENT-UPDATE-VAR-PROTO", nil)
				variant := testutil.CreateTestVariant(t, tx, product.ID, "VARIANT-UPDATE-PROTO-001", nil)
				variantID = variant.ID.String()
			} else {
				variantID = uuid.New().String()
			}

			url := "/api/v1/variants/" + variantID
			req, _ := testutil.MakeRequest(http.MethodPut, url, tt.request)
			rr := httptest.NewRecorder()

			handler := UpdateVariantHandler_Proto(tx)
			handler.ServeHTTP(rr, req)

			testutil.AssertStatus(t, rr, tt.expectedStatus)

			if tt.expectedError != nil {
				errorResp := &pb.ErrorResponse{}
				testutil.ParseProtoResponse(t, rr, errorResp)
				if errorResp.Error.Code != *tt.expectedError {
					t.Errorf("expected error code %s, got %v", *tt.expectedError, errorResp.Error.Code)
				}
			}
		})
	}
}

// TestDeleteVariant tests DELETE /api/v1/variants/{id} using protobuf
func TestDeleteVariant(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	tests := []struct {
		name           string
		setupVariant   bool
		variantID      string
		expectedStatus int
		expectedError  *string
	}{
		{
			name:           "delete existing variant",
			setupVariant:   true,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "delete non-existent variant returns 404",
			setupVariant:   false,
			variantID:      uuid.New().String(),
			expectedStatus: http.StatusNotFound,
			expectedError:  stringPtr("not_found"),
		},
		{
			name:           "delete with invalid UUID returns 400",
			setupVariant:   false,
			variantID:      "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  stringPtr("bad_request"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := testutil.BeginTestTransaction(t, db)

			var variantID string
			if tt.setupVariant {
				product := testutil.CreateTestProduct(t, tx, "Parent Product", "PARENT-DELETE-VAR-PROTO", nil)
				variant := testutil.CreateTestVariant(t, tx, product.ID, "VARIANT-DELETE-PROTO-001", nil)
				variantID = variant.ID.String()
			} else {
				variantID = tt.variantID
			}

			url := "/api/v1/variants/" + variantID
			req, _ := testutil.MakeRequest(http.MethodDelete, url, nil)
			rr := httptest.NewRecorder()

			handler := DeleteVariantHandler_Proto(tx)
			handler.ServeHTTP(rr, req)

			testutil.AssertStatus(t, rr, tt.expectedStatus)

			if tt.expectedError != nil && rr.Code != http.StatusNoContent {
				errorResp := &pb.ErrorResponse{}
				testutil.ParseProtoResponse(t, rr, errorResp)
				if errorResp.Error.Code != *tt.expectedError {
					t.Errorf("expected error code %s, got %v", *tt.expectedError, errorResp.Error.Code)
				}
			}
		})
	}
}

// TestCascadeDeleteVariants tests cascade delete using protobuf
func TestCascadeDeleteVariants(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	tx := testutil.BeginTestTransaction(t, db)

	// Create parent product with multiple variants
	product := testutil.CreateTestProduct(t, tx, "Parent Product", "CASCADE-PROTO-001", nil)
	variant1 := testutil.CreateTestVariant(t, tx, product.ID, "CASCADE-VAR-PROTO-001", nil)
	variant2 := testutil.CreateTestVariant(t, tx, product.ID, "CASCADE-VAR-PROTO-002", nil)

	// Delete parent product
	delReq, _ := testutil.MakeRequest(http.MethodDelete, "/api/v1/products/"+product.ID.String(), nil)
	delRR := httptest.NewRecorder()

	productHandler := DeleteProductHandler_Proto(tx)
	productHandler.ServeHTTP(delRR, delReq)

	testutil.AssertStatus(t, delRR, http.StatusNoContent)

	// Try to retrieve variants - should return 404
	varReq1, _ := testutil.MakeRequest(http.MethodGet, "/api/v1/variants/"+variant1.ID.String(), nil)
	varRR1 := httptest.NewRecorder()

	variantHandler := GetVariantHandler_Proto(tx)
	variantHandler.ServeHTTP(varRR1, varReq1)

	testutil.AssertStatus(t, varRR1, http.StatusNotFound)

	// Same for second variant
	varReq2, _ := testutil.MakeRequest(http.MethodGet, "/api/v1/variants/"+variant2.ID.String(), nil)
	varRR2 := httptest.NewRecorder()

	variantHandler.ServeHTTP(varRR2, varReq2)
	testutil.AssertStatus(t, varRR2, http.StatusNotFound)
}

