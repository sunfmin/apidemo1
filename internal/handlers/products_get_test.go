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

// TestGetProduct tests GET /api/v1/products/{id} endpoint (T032)
func TestGetProduct(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	tests := []struct {
		name           string
		setupProduct   bool
		productID      string
		expectedStatus int
		expectedError  *string
	}{
		{
			name:           "retrieve existing product with all attributes",
			setupProduct:   true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent product ID returns 404",
			setupProduct:   false,
			productID:      uuid.New().String(),
			expectedStatus: http.StatusNotFound,
			expectedError:  stringPtr("not_found"),
		},
		{
			name:           "invalid UUID format returns 400",
			setupProduct:   false,
			productID:      "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  stringPtr("bad_request"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := testutil.BeginTestTransaction(t, db)

			var productID string
			if tt.setupProduct {
				attrs := json.RawMessage(`{"color":{"type":"string","value":"blue"}}`)
				product := testutil.CreateTestProduct(t, tx, "Test Product", "GET-TEST-001", attrs)
				productID = product.ID.String()
			} else if tt.productID != "" {
				productID = tt.productID
			}

			url := fmt.Sprintf("/api/v1/products/%s", productID)
			req, _ := testutil.MakeRequest(http.MethodGet, url, nil)
			rr := httptest.NewRecorder()

			handler := GetProductHandler(tx)
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

// TestListProducts tests GET /api/v1/products endpoint with pagination (T033)
func TestListProducts(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	tests := []struct {
		name           string
		setupCount     int
		queryParams    string
		expectedStatus int
		expectedError  *string
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name:           "list with default pagination",
			setupCount:     5,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				products := body["products"].([]interface{})
				if len(products) != 5 {
					t.Errorf("expected 5 products, got %d", len(products))
				}
			},
		},
		{
			name:           "empty product list",
			setupCount:     0,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				products := body["products"].([]interface{})
				if len(products) != 0 {
					t.Errorf("expected 0 products, got %d", len(products))
				}
			},
		},
		{
			name:           "pagination with page_size",
			setupCount:     10,
			queryParams:    "?page=1&page_size=5",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				products := body["products"].([]interface{})
				if len(products) > 5 {
					t.Errorf("expected max 5 products, got %d", len(products))
				}
			},
		},
		{
			name:           "invalid page number (zero)",
			queryParams:    "?page=0",
			expectedStatus: http.StatusBadRequest,
			expectedError:  stringPtr("bad_request"),
		},
		{
			name:           "invalid page number (negative)",
			queryParams:    "?page=-1",
			expectedStatus: http.StatusBadRequest,
			expectedError:  stringPtr("bad_request"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := testutil.BeginTestTransaction(t, db)

			// Create test products
			for i := 0; i < tt.setupCount; i++ {
				sku := fmt.Sprintf("LIST-TEST-%03d", i)
				testutil.CreateTestProduct(t, tx, fmt.Sprintf("Product %d", i), sku, nil)
			}

			url := "/api/v1/products" + tt.queryParams
			req, _ := testutil.MakeRequest(http.MethodGet, url, nil)
			rr := httptest.NewRecorder()

			handler := ListProductsHandler(tx)
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

// TestUpdateProduct tests PUT /api/v1/products/{id} endpoint (T034)
func TestUpdateProduct(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	tests := []struct {
		name           string
		setupProduct   bool
		requestBody    map[string]interface{}
		expectedStatus int
		expectedError  *string
	}{
		{
			name:         "update product name and description",
			setupProduct: true,
			requestBody: map[string]interface{}{
				"name":        "Updated Product Name",
				"description": "Updated description",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "update product attributes",
			setupProduct: true,
			requestBody: map[string]interface{}{
				"attributes": map[string]interface{}{
					"new_attr": map[string]interface{}{
						"type":  "string",
						"value": "new_value",
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "update non-existent product returns 404",
			setupProduct: false,
			requestBody: map[string]interface{}{
				"name": "Updated Name",
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  stringPtr("not_found"),
		},
		{
			name:         "update with empty name returns 400",
			setupProduct: true,
			requestBody: map[string]interface{}{
				"name": "",
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
				product := testutil.CreateTestProduct(t, tx, "Original Name", "UPDATE-TEST-001", nil)
				productID = product.ID.String()
			} else {
				productID = uuid.New().String()
			}

			url := fmt.Sprintf("/api/v1/products/%s", productID)
			req, _ := testutil.MakeRequest(http.MethodPut, url, tt.requestBody)
			rr := httptest.NewRecorder()

			handler := UpdateProductHandler(tx)
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

// TestDeleteProduct tests DELETE /api/v1/products/{id} endpoint (T035)
func TestDeleteProduct(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	tests := []struct {
		name           string
		setupProduct   bool
		productID      string
		expectedStatus int
		expectedError  *string
	}{
		{
			name:           "delete existing product",
			setupProduct:   true,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "delete non-existent product returns 404",
			setupProduct:   false,
			productID:      uuid.New().String(),
			expectedStatus: http.StatusNotFound,
			expectedError:  stringPtr("not_found"),
		},
		{
			name:           "delete with invalid UUID returns 400",
			setupProduct:   false,
			productID:      "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  stringPtr("bad_request"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := testutil.BeginTestTransaction(t, db)

			var productID string
			if tt.setupProduct {
				product := testutil.CreateTestProduct(t, tx, "To Delete", "DELETE-TEST-001", nil)
				productID = product.ID.String()
			} else {
				productID = tt.productID
			}

			url := fmt.Sprintf("/api/v1/products/%s", productID)
			req, _ := testutil.MakeRequest(http.MethodDelete, url, nil)
			rr := httptest.NewRecorder()

			handler := DeleteProductHandler(tx)
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

// Placeholder handler functions - will be implemented in T040-T044
func GetProductHandler(db interface{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(`{"error":{"code":"not_implemented","message":"Handler not yet implemented"}}`))
	})
}

func ListProductsHandler(db interface{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(`{"error":{"code":"not_implemented","message":"Handler not yet implemented"}}`))
	})
}

func UpdateProductHandler(db interface{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(`{"error":{"code":"not_implemented","message":"Handler not yet implemented"}}`))
	})
}

func DeleteProductHandler(db interface{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(`{"error":{"code":"not_implemented","message":"Handler not yet implemented"}}`))
	})
}

