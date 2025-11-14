package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"apidemo1/internal/testutil"

	"github.com/google/uuid"
)

// TestCreateVariant tests POST /api/v1/products/{productId}/variants endpoint (T046)
func TestCreateVariant(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	tests := []struct {
		name           string
		setupProduct   bool
		productID      string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedError  *string
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		// ========== HAPPY PATH ==========
		{
			name:         "create variant with SKU and attributes",
			setupProduct: true,
			requestBody: map[string]interface{}{
				"sku": "TSHIRT-BLUE-M",
				"attributes": map[string]interface{}{
					"size": map[string]interface{}{
						"type":  "string",
						"value": "M",
					},
					"color": map[string]interface{}{
						"type":  "string",
						"value": "blue",
					},
					"price": map[string]interface{}{
						"type":  "number",
						"value": 29.99,
					},
				},
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if body["id"] == nil {
					t.Error("expected id to be present")
				}
				if body["product_id"] == nil {
					t.Error("expected product_id to be present")
				}
				if body["sku"] != "TSHIRT-BLUE-M" {
					t.Errorf("expected sku 'TSHIRT-BLUE-M', got %v", body["sku"])
				}
				attrs := body["attributes"].(map[string]interface{})
				if len(attrs) != 3 {
					t.Errorf("expected 3 attributes, got %d", len(attrs))
				}
			},
		},
		{
			name:         "create variant with minimal fields (SKU only)",
			setupProduct: true,
			requestBody: map[string]interface{}{
				"sku": "MINIMAL-VARIANT-001",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				attrs := body["attributes"].(map[string]interface{})
				if len(attrs) != 0 {
					t.Errorf("expected 0 attributes, got %d", len(attrs))
				}
			},
		},

		// ========== DATA STATE - NON-EXISTENT PARENT ==========
		{
			name:         "create variant for non-existent parent product",
			setupProduct: false,
			productID:    uuid.New().String(),
			requestBody: map[string]interface{}{
				"sku": "NO-PARENT-001",
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  stringPtr("not_found"),
		},

		// ========== INPUT VALIDATION ==========
		{
			name:         "empty SKU",
			setupProduct: true,
			requestBody: map[string]interface{}{
				"sku": "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  stringPtr("bad_request"),
		},
		{
			name:         "missing SKU field",
			setupProduct: true,
			requestBody:  map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  stringPtr("bad_request"),
		},
		{
			name:         "invalid attribute type",
			setupProduct: true,
			requestBody: map[string]interface{}{
				"sku": "INVALID-ATTR-001",
				"attributes": map[string]interface{}{
					"bad_attr": map[string]interface{}{
						"type":  "invalid_type",
						"value": "something",
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  stringPtr("bad_request"),
		},

		// ========== BOUNDARY CONDITIONS ==========
		{
			name:         "SKU at maximum length (100 chars)",
			setupProduct: true,
			requestBody: map[string]interface{}{
				"sku": generateString(100),
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:         "SKU exceeds maximum length (101 chars)",
			setupProduct: true,
			requestBody: map[string]interface{}{
				"sku": generateString(101),
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
				product := testutil.CreateTestProduct(t, tx, "Parent Product", "PARENT-001", nil)
				productID = product.ID.String()
			} else if tt.productID != "" {
				productID = tt.productID
			}

			url := fmt.Sprintf("/api/v1/products/%s/variants", productID)
			req, _ := testutil.MakeRequest(http.MethodPost, url, tt.requestBody)
			rr := httptest.NewRecorder()

			handler := CreateVariantHandler(tx, productID)
			handler.ServeHTTP(rr, req)

			testutil.AssertStatus(t, rr, tt.expectedStatus)

			var response map[string]interface{}
			json.NewDecoder(rr.Body).Decode(&response)

			if tt.expectedError != nil {
				errorResp := response["error"].(map[string]interface{})
				if errorResp["code"] != *tt.expectedError {
					t.Errorf("expected error code %s, got %v", *tt.expectedError, errorResp["code"])
				}
			}

			if tt.checkResponse != nil && tt.expectedError == nil {
				tt.checkResponse(t, response)
			}
		})
	}
}

// TestCreateVariantDuplicateSKU tests duplicate SKU constraint across products and variants
func TestCreateVariantDuplicateSKU(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	tx := testutil.BeginTestTransaction(t, db)

	// Create parent product
	product := testutil.CreateTestProduct(t, tx, "Parent Product", "PARENT-002", nil)

	// Create first variant
	variant1 := testutil.CreateTestVariant(t, tx, product.ID, "DUP-VARIANT-001", nil)
	if variant1 == nil {
		t.Fatal("failed to create first variant")
	}

	// Try to create second variant with same SKU
	url := fmt.Sprintf("/api/v1/products/%s/variants", product.ID.String())
	req, _ := testutil.MakeRequest(http.MethodPost, url, map[string]interface{}{
		"sku": "DUP-VARIANT-001", // Duplicate SKU
	})

	rr := httptest.NewRecorder()
	handler := CreateVariantHandler(tx, product.ID.String())
	handler.ServeHTTP(rr, req)

	// Should return 409 Conflict
	testutil.AssertStatus(t, rr, http.StatusConflict)

	var response map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&response)
	errorResp := response["error"].(map[string]interface{})
	
	if errorResp["code"] != "conflict" {
		t.Errorf("expected error code 'conflict', got %v", errorResp["code"])
	}
}

// TestListVariants tests GET /api/v1/products/{productId}/variants endpoint (T047)
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
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name:           "list all variants for a product",
			setupProduct:   true,
			setupVariants:  3,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				variants := body["variants"].([]interface{})
				if len(variants) != 3 {
					t.Errorf("expected 3 variants, got %d", len(variants))
				}
			},
		},
		{
			name:           "product with zero variants",
			setupProduct:   true,
			setupVariants:  0,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				variants := body["variants"].([]interface{})
				if len(variants) != 0 {
					t.Errorf("expected 0 variants, got %d", len(variants))
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
				product := testutil.CreateTestProduct(t, tx, "Parent Product", fmt.Sprintf("PARENT-%d", tt.setupVariants), nil)
				productID = product.ID.String()

				// Create variants
				for i := 0; i < tt.setupVariants; i++ {
					sku := fmt.Sprintf("VARIANT-%d-%d", tt.setupVariants, i)
					testutil.CreateTestVariant(t, tx, product.ID, sku, nil)
				}
			} else {
				productID = tt.productID
			}

			url := fmt.Sprintf("/api/v1/products/%s/variants", productID)
			req, _ := testutil.MakeRequest(http.MethodGet, url, nil)
			rr := httptest.NewRecorder()

			handler := ListVariantsHandler(tx, productID)
			handler.ServeHTTP(rr, req)

			testutil.AssertStatus(t, rr, tt.expectedStatus)

			var response map[string]interface{}
			json.NewDecoder(rr.Body).Decode(&response)

			if tt.expectedError != nil {
				errorResp := response["error"].(map[string]interface{})
				if errorResp["code"] != *tt.expectedError {
					t.Errorf("expected error code %s, got %v", *tt.expectedError, errorResp["code"])
				}
			}

			if tt.checkResponse != nil && tt.expectedError == nil {
				tt.checkResponse(t, response)
			}
		})
	}
}

// TestGetVariant tests GET /api/v1/variants/{id} endpoint (T048)
func TestGetVariant(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	tests := []struct {
		name           string
		setupVariant   bool
		variantID      string
		expectedStatus int
		expectedError  *string
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name:           "retrieve existing variant",
			setupVariant:   true,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if body["id"] == nil {
					t.Error("expected id to be present")
				}
				if body["product_id"] == nil {
					t.Error("expected product_id to be present")
				}
			},
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
				product := testutil.CreateTestProduct(t, tx, "Parent Product", "PARENT-GET-001", nil)
				attrs := json.RawMessage(`{"size":{"type":"string","value":"L"}}`)
				variant := testutil.CreateTestVariant(t, tx, product.ID, "VARIANT-GET-001", attrs)
				variantID = variant.ID.String()
			} else {
				variantID = tt.variantID
			}

			url := fmt.Sprintf("/api/v1/variants/%s", variantID)
			req, _ := testutil.MakeRequest(http.MethodGet, url, nil)
			rr := httptest.NewRecorder()

			handler := GetVariantHandler(tx)
			handler.ServeHTTP(rr, req)

			testutil.AssertStatus(t, rr, tt.expectedStatus)

			var response map[string]interface{}
			json.NewDecoder(rr.Body).Decode(&response)

			if tt.expectedError != nil {
				errorResp := response["error"].(map[string]interface{})
				if errorResp["code"] != *tt.expectedError {
					t.Errorf("expected error code %s, got %v", *tt.expectedError, errorResp["code"])
				}
			}

			if tt.checkResponse != nil && tt.expectedError == nil {
				tt.checkResponse(t, response)
			}
		})
	}
}

// TestUpdateVariant tests PUT /api/v1/variants/{id} endpoint (T049)
func TestUpdateVariant(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	tests := []struct {
		name           string
		setupVariant   bool
		requestBody    map[string]interface{}
		expectedStatus int
		expectedError  *string
	}{
		{
			name:         "update variant attributes",
			setupVariant: true,
			requestBody: map[string]interface{}{
				"attributes": map[string]interface{}{
					"updated_attr": map[string]interface{}{
						"type":  "string",
						"value": "updated_value",
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "update non-existent variant returns 404",
			setupVariant: false,
			requestBody: map[string]interface{}{
				"attributes": map[string]interface{}{},
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
				product := testutil.CreateTestProduct(t, tx, "Parent Product", "PARENT-UPDATE-001", nil)
				variant := testutil.CreateTestVariant(t, tx, product.ID, "VARIANT-UPDATE-001", nil)
				variantID = variant.ID.String()
			} else {
				variantID = uuid.New().String()
			}

			url := fmt.Sprintf("/api/v1/variants/%s", variantID)
			req, _ := testutil.MakeRequest(http.MethodPut, url, tt.requestBody)
			rr := httptest.NewRecorder()

			handler := UpdateVariantHandler(tx)
			handler.ServeHTTP(rr, req)

			testutil.AssertStatus(t, rr, tt.expectedStatus)

			if tt.expectedError != nil {
				var response map[string]interface{}
				json.NewDecoder(rr.Body).Decode(&response)
				errorResp := response["error"].(map[string]interface{})
				if errorResp["code"] != *tt.expectedError {
					t.Errorf("expected error code %s, got %v", *tt.expectedError, errorResp["code"])
				}
			}
		})
	}
}

// TestDeleteVariant tests DELETE /api/v1/variants/{id} endpoint (T050)
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
				product := testutil.CreateTestProduct(t, tx, "Parent Product", "PARENT-DELETE-001", nil)
				variant := testutil.CreateTestVariant(t, tx, product.ID, "VARIANT-DELETE-001", nil)
				variantID = variant.ID.String()
			} else {
				variantID = tt.variantID
			}

			url := fmt.Sprintf("/api/v1/variants/%s", variantID)
			req, _ := testutil.MakeRequest(http.MethodDelete, url, nil)
			rr := httptest.NewRecorder()

			handler := DeleteVariantHandler(tx)
			handler.ServeHTTP(rr, req)

			testutil.AssertStatus(t, rr, tt.expectedStatus)

			if tt.expectedError != nil {
				var response map[string]interface{}
				json.NewDecoder(rr.Body).Decode(&response)
				errorResp := response["error"].(map[string]interface{})
				if errorResp["code"] != *tt.expectedError {
					t.Errorf("expected error code %s, got %v", *tt.expectedError, errorResp["code"])
				}
			}
		})
	}
}

// TestCascadeDeleteVariants tests that variants are deleted when parent product is deleted (T051)
func TestCascadeDeleteVariants(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	tx := testutil.BeginTestTransaction(t, db)

	// Create parent product with multiple variants
	product := testutil.CreateTestProduct(t, tx, "Parent Product", "CASCADE-001", nil)
	variant1 := testutil.CreateTestVariant(t, tx, product.ID, "CASCADE-VARIANT-001", nil)
	variant2 := testutil.CreateTestVariant(t, tx, product.ID, "CASCADE-VARIANT-002", nil)

	// Delete parent product
	url := fmt.Sprintf("/api/v1/products/%s", product.ID.String())
	req, _ := testutil.MakeRequest(http.MethodDelete, url, nil)
	rr := httptest.NewRecorder()

	productHandler := DeleteProductHandler(tx)
	productHandler.ServeHTTP(rr, req)

	testutil.AssertStatus(t, rr, http.StatusNoContent)

	// Try to retrieve variants - should return 404
	variantURL1 := fmt.Sprintf("/api/v1/variants/%s", variant1.ID.String())
	variantReq1, _ := testutil.MakeRequest(http.MethodGet, variantURL1, nil)
	variantRR1 := httptest.NewRecorder()

	variantHandler := GetVariantHandler(tx)
	variantHandler.ServeHTTP(variantRR1, variantReq1)

	// Should return 404 because variant was cascade deleted
	testutil.AssertStatus(t, variantRR1, http.StatusNotFound)

	// Same for second variant
	variantURL2 := fmt.Sprintf("/api/v1/variants/%s", variant2.ID.String())
	variantReq2, _ := testutil.MakeRequest(http.MethodGet, variantURL2, nil)
	variantRR2 := httptest.NewRecorder()

	variantHandler.ServeHTTP(variantRR2, variantReq2)
	testutil.AssertStatus(t, variantRR2, http.StatusNotFound)
}


