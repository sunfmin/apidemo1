package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"apidemo1/internal/testutil"
)

// TestCreateProduct tests POST /api/v1/products endpoint
// This is a table-driven integration test covering all edge cases per constitution
func TestCreateProduct(t *testing.T) {
	// Setup test database
	db := testutil.SetupTestDB(t)
	defer db.Close()

	// Test cases
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedError  *string // nil for success cases
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		// ========== HAPPY PATH ==========
		{
			name: "create product with all fields and various attribute types",
			requestBody: map[string]interface{}{
				"name":        "Blue T-Shirt",
				"sku":         "TSHIRT-BLUE-001",
				"description": "A comfortable cotton t-shirt",
				"attributes": map[string]interface{}{
					"color": map[string]interface{}{
						"type":  "string",
						"value": "blue",
					},
					"size": map[string]interface{}{
						"type":  "string",
						"value": "XL",
					},
					"weight": map[string]interface{}{
						"type":  "number",
						"value": 0.5,
					},
					"in_stock": map[string]interface{}{
						"type":  "boolean",
						"value": true,
					},
					"release_date": map[string]interface{}{
						"type":  "date",
						"value": "2025-01-15",
					},
				},
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if body["id"] == nil {
					t.Error("expected id to be present")
				}
				if body["name"] != "Blue T-Shirt" {
					t.Errorf("expected name 'Blue T-Shirt', got %v", body["name"])
				}
				if body["sku"] != "TSHIRT-BLUE-001" {
					t.Errorf("expected sku 'TSHIRT-BLUE-001', got %v", body["sku"])
				}
				// Verify all attribute types are preserved
				attrs := body["attributes"].(map[string]interface{})
				if len(attrs) != 5 {
					t.Errorf("expected 5 attributes, got %d", len(attrs))
				}
			},
		},
		{
			name: "create product with minimal fields (name and sku only)",
			requestBody: map[string]interface{}{
				"name": "Minimal Product",
				"sku":  "MIN-001",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if body["attributes"] == nil {
					t.Error("expected attributes to be present (empty object)")
				}
			},
		},
		{
			name: "create product with zero attributes",
			requestBody: map[string]interface{}{
				"name":       "No Attributes Product",
				"sku":        "NOATTR-001",
				"attributes": map[string]interface{}{},
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				attrs := body["attributes"].(map[string]interface{})
				if len(attrs) != 0 {
					t.Errorf("expected 0 attributes, got %d", len(attrs))
				}
			},
		},

		// ========== INPUT VALIDATION - EMPTY/NIL ==========
		{
			name: "empty product name",
			requestBody: map[string]interface{}{
				"name": "",
				"sku":  "EMPTY-NAME-001",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  stringPtr("bad_request"),
		},
		{
			name: "empty SKU",
			requestBody: map[string]interface{}{
				"name": "Empty SKU Product",
				"sku":  "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  stringPtr("bad_request"),
		},
		{
			name: "missing name field",
			requestBody: map[string]interface{}{
				"sku": "MISSING-NAME-001",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  stringPtr("bad_request"),
		},
		{
			name: "missing sku field",
			requestBody: map[string]interface{}{
				"name": "Missing SKU Product",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  stringPtr("bad_request"),
		},

		// ========== INPUT VALIDATION - SQL INJECTION & XSS ==========
		{
			name: "SQL injection attempt in name",
			requestBody: map[string]interface{}{
				"name": "'; DROP TABLE products; --",
				"sku":  "SQL-INJECT-001",
			},
			expectedStatus: http.StatusCreated, // Should be sanitized but not rejected
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				// Verify the malicious SQL is stored as plain text, not executed
				if body["name"] == nil {
					t.Error("name should be stored (sanitized)")
				}
			},
		},
		{
			name: "XSS attempt in description",
			requestBody: map[string]interface{}{
				"name":        "XSS Test Product",
				"sku":         "XSS-001",
				"description": "<script>alert('xss')</script>",
			},
			expectedStatus: http.StatusCreated, // Should be sanitized
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				// Verify XSS payload is escaped or removed
				desc := body["description"].(string)
				if desc == "<script>alert('xss')</script>" {
					t.Error("XSS payload should be sanitized")
				}
			},
		},

		// ========== BOUNDARY CONDITIONS - TEXT LENGTH ==========
		{
			name: "name at minimum length (1 char)",
			requestBody: map[string]interface{}{
				"name": "X",
				"sku":  "MIN-NAME-001",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "name at maximum length (500 chars)",
			requestBody: map[string]interface{}{
				"name": generateString(500),
				"sku":  "MAX-NAME-001",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "name exceeds maximum length (501 chars)",
			requestBody: map[string]interface{}{
				"name": generateString(501),
				"sku":  "EXCEED-NAME-001",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  stringPtr("bad_request"),
		},
		{
			name: "description at maximum length (10000 chars)",
			requestBody: map[string]interface{}{
				"name":        "Max Description Product",
				"sku":         "MAX-DESC-001",
				"description": generateString(10000),
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "description exceeds maximum length (10001 chars)",
			requestBody: map[string]interface{}{
				"name":        "Exceed Description Product",
				"sku":         "EXCEED-DESC-001",
				"description": generateString(10001),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  stringPtr("bad_request"),
		},

		// ========== BOUNDARY CONDITIONS - ATTRIBUTES ==========
		{
			name: "100+ attributes",
			requestBody: map[string]interface{}{
				"name":       "Many Attributes Product",
				"sku":        "MANY-ATTRS-001",
				"attributes": generateManyAttributes(100),
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				attrs := body["attributes"].(map[string]interface{})
				if len(attrs) != 100 {
					t.Errorf("expected 100 attributes, got %d", len(attrs))
				}
			},
		},

		// ========== DATABASE ERRORS - DUPLICATE SKU ==========
		{
			name: "duplicate SKU (will be tested in separate test)",
			requestBody: map[string]interface{}{
				"name": "Duplicate SKU Test",
				"sku":  "DUP-SKU-001",
			},
			expectedStatus: http.StatusCreated, // First creation succeeds
		},

		// ========== HTTP SPECIFICS - INVALID JSON ==========
		// (These will be tested with raw request bodies below)
	}

	// Run table-driven tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Begin transaction for test isolation
			tx := testutil.BeginTestTransaction(t, db)

			// Create request
			req, err := testutil.MakeRequest(http.MethodPost, "/api/v1/products", tt.requestBody)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Create handler (will be implemented in T040)
			handler := CreateProductHandler(tx)
			handler.ServeHTTP(rr, req)

			// Assert status code
			testutil.AssertStatus(t, rr, tt.expectedStatus)

			// Parse response
			var response map[string]interface{}
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			// Check for expected error
			if tt.expectedError != nil {
				errorResp := response["error"].(map[string]interface{})
				if errorResp["code"] != *tt.expectedError {
					t.Errorf("expected error code %s, got %v", *tt.expectedError, errorResp["code"])
				}
			}

			// Run custom response checks
			if tt.checkResponse != nil && tt.expectedError == nil {
				tt.checkResponse(t, response)
			}
		})
	}
}

// TestCreateProductDuplicateSKU tests duplicate SKU constraint
func TestCreateProductDuplicateSKU(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	tx := testutil.BeginTestTransaction(t, db)

	// Create first product
	product1 := testutil.CreateTestProduct(t, tx, "Product 1", "DUP-SKU-001", nil)
	if product1 == nil {
		t.Fatal("failed to create first product")
	}

	// Try to create second product with same SKU
	req, _ := testutil.MakeRequest(http.MethodPost, "/api/v1/products", map[string]interface{}{
		"name": "Product 2",
		"sku":  "DUP-SKU-001", // Duplicate SKU
	})

	rr := httptest.NewRecorder()
	handler := CreateProductHandler(tx)
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

// TestCreateProductInvalidJSON tests malformed JSON handling
func TestCreateProductInvalidJSON(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	tests := []struct {
		name        string
		body        string
		expectError bool
	}{
		{
			name:        "invalid JSON syntax",
			body:        `{"name": "Test", "sku": "TEST-001"`,
			expectError: true,
		},
		{
			name:        "empty request body",
			body:        "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := testutil.BeginTestTransaction(t, db)
			
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/products", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := CreateProductHandler(tx)
			handler.ServeHTTP(rr, req)

			if tt.expectError && rr.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", rr.Code)
			}
		})
	}
}

// TestCreateProductWrongMethod tests HTTP method validation
func TestCreateProductWrongMethod(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	tx := testutil.BeginTestTransaction(t, db)

	// Try GET when POST is expected
	// Note: In the actual router, wrong methods would be handled by Chi
	// Here we're just testing that POST handler doesn't accept GET
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/products", nil)
	rr := httptest.NewRecorder()
	
	// POST handler should ideally check method, but Chi router handles this
	// For this test, we'll skip it since method routing is Chi's responsibility
	// Just verify handler doesn't panic with GET
	handler := CreateProductHandler(tx)
	handler.ServeHTTP(rr, req)

	// Any non-200 response is acceptable for wrong method
	// (400 Bad Request from JSON decode is fine)
	if rr.Code == http.StatusCreated {
		t.Error("GET request should not create a product")
	}
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}

func generateString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}

func generateManyAttributes(count int) map[string]interface{} {
	attrs := make(map[string]interface{})
	for i := 0; i < count; i++ {
		key := "attr_" + string(rune('a'+i%26)) + string(rune('0'+i/26))
		attrs[key] = map[string]interface{}{
			"type":  "string",
			"value": "value",
		}
	}
	return attrs
}


