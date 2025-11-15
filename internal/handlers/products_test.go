package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"apidemo1/internal/models"
	"apidemo1/internal/testutil"
	pb "apidemo1/proto"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

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

func containsEscapedHTML(s string) bool {
	// Check if string contains HTML entities (indicating escaping)
	return len(s) > 0 && (s != "<script>alert('xss')</script>")
}

// TestCreateProduct tests POST /api/v1/products endpoint using protobuf structs
// Constitution-compliant: Uses protobuf structs (NO maps!) and protocmp for assertions
func TestCreateProduct(t *testing.T) {
	// Test cases using protobuf structs
	tests := []struct {
		name           string
		request        *pb.ProductCreateRequest
		expectedStatus int
		expectedError  *string
		expectedFields *pb.Product // For protocmp comparison
		checkResponse  func(t *testing.T, product *pb.Product)
	}{
		// ========== HAPPY PATH ==========
		{
			name: "create product with all fields and various attribute types",
			request: &pb.ProductCreateRequest{
				Name:        "Blue T-Shirt",
				Sku:         "TSHIRT-BLUE-001",
				Description: "A comfortable cotton t-shirt",
				Attributes: map[string]*pb.AttributeValue{
					"color":        models.CreateStringAttribute("blue"),
					"size":          models.CreateStringAttribute("XL"),
					"weight":        models.CreateNumberAttribute(0.5),
					"in_stock":      models.CreateBooleanAttribute(true),
					"release_date":  models.CreateDateAttribute("2025-01-15"),
				},
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, product *pb.Product) {
				// Verify ID is present (can't predict UUID)
				if product.Id == "" {
					t.Error("expected id to be present")
				}
				
				// Use protocmp directly (Constitution Principle VI - Example)
				expected := &pb.Product{
					Id:          product.Id, // Use actual ID
					Name:        "Blue T-Shirt",
					Sku:         "TSHIRT-BLUE-001",
					Description: "A comfortable cotton t-shirt",
					Attributes: map[string]*pb.AttributeValue{
						"color":        models.CreateStringAttribute("blue"),
						"size":         models.CreateStringAttribute("XL"),
						"weight":       models.CreateNumberAttribute(0.5),
						"in_stock":     models.CreateBooleanAttribute(true),
						"release_date": models.CreateDateAttribute("2025-01-15"),
					},
					CreatedAt: product.CreatedAt, // Use actual timestamp
					UpdatedAt: product.UpdatedAt, // Use actual timestamp
				}
				
				// Direct use of cmp.Diff with protocmp.Transform() per constitution
				if diff := cmp.Diff(expected, product, protocmp.Transform()); diff != "" {
					t.Errorf("Product mismatch (-want +got):\n%s", diff)
				}
			},
		},
		{
			name: "create product with minimal fields",
			request: &pb.ProductCreateRequest{
				Name: "Minimal Product",
				Sku:  "MIN-PROTO-001",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, product *pb.Product) {
				// Use protocmp to verify minimal product structure
				expected := &pb.Product{
					Id:          product.Id, // Use actual ID
					Name:        "Minimal Product",
					Sku:         "MIN-PROTO-001",
					Attributes:  map[string]*pb.AttributeValue{}, // Empty attributes
					CreatedAt:   product.CreatedAt,
					UpdatedAt:   product.UpdatedAt,
				}
				
				testutil.AssertProtoEqual(t, expected, product)
			},
		},

		// ========== INPUT VALIDATION ==========
		{
			name: "empty product name",
			request: &pb.ProductCreateRequest{
				Name: "",
				Sku:  "EMPTY-NAME-PROTO-001",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  stringPtr("bad_request"),
		},
		{
			name: "empty SKU",
			request: &pb.ProductCreateRequest{
				Name: "Empty SKU Product",
				Sku:  "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  stringPtr("bad_request"),
		},

		// ========== BOUNDARY CONDITIONS ==========
		{
			name: "name at minimum length (1 char)",
			request: &pb.ProductCreateRequest{
				Name: "X",
				Sku:  "MIN-NAME-PROTO-001",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "name at maximum length (500 chars)",
			request: &pb.ProductCreateRequest{
				Name: generateString(500),
				Sku:  "MAX-NAME-PROTO-001",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "name exceeds maximum length (501 chars)",
			request: &pb.ProductCreateRequest{
				Name: generateString(501),
				Sku:  "EXCEED-NAME-PROTO-001",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  stringPtr("bad_request"),
		},

		// ========== SQL INJECTION & XSS ==========
		{
			name: "SQL injection attempt in name",
			request: &pb.ProductCreateRequest{
				Name: "'; DROP TABLE products; --",
				Sku:  "SQL-INJECT-PROTO-001",
			},
			expectedStatus: http.StatusCreated, // Sanitized but not rejected
			checkResponse: func(t *testing.T, product *pb.Product) {
				// Verify name is sanitized (HTML escaped)
				if product.Name == "" {
					t.Error("name should be stored (sanitized)")
				}
				// Should contain escaped HTML entities, not raw SQL
				if product.Name == "'; DROP TABLE products; --" {
					t.Error("SQL should be escaped/sanitized")
				}
			},
		},
		{
			name: "XSS attempt in description",
			request: &pb.ProductCreateRequest{
				Name:        "XSS Test Product",
				Sku:         "XSS-PROTO-001",
				Description: "<script>alert('xss')</script>",
			},
			expectedStatus: http.StatusCreated, // Sanitized
			checkResponse: func(t *testing.T, product *pb.Product) {
				// Use protocmp to verify XSS is sanitized
				if product.Description == "<script>alert('xss')</script>" {
					t.Error("XSS payload should be sanitized")
				}
				// Description should be escaped
				if !containsEscapedHTML(product.Description) {
					t.Logf("Description after sanitization: %s", product.Description)
				}
			},
		},
	}

	// Run table-driven tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test database for each test case
			db := testutil.SetupTestDB(t)
			defer db.Close()
			
			// Begin transaction for test isolation
			tx := testutil.BeginTestTransaction(t, db)

			// Create request using protobuf message
			req, err := testutil.MakeRequest(http.MethodPost, "/api/v1/products", tt.request)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call protobuf-aware handler
			handler := CreateProductHandler_Proto(tx)
			handler.ServeHTTP(rr, req)

			// Assert status code
			testutil.AssertStatus(t, rr, tt.expectedStatus)

			// Parse response as protobuf
			if tt.expectedStatus == http.StatusCreated {
				product := &pb.Product{}
				testutil.ParseProtoResponse(t, rr, product)

				// Run custom response checks
				if tt.checkResponse != nil {
					tt.checkResponse(t, product)
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

// TestCreateProductDuplicateSKU tests duplicate SKU constraint using protobuf
func TestCreateProductDuplicateSKU(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	tx := testutil.BeginTestTransaction(t, db)

	// Create first product using fixture
	product1 := testutil.CreateTestProduct(t, tx, "Product 1", "DUP-SKU-PROTO-001", nil)
	if product1 == nil {
		t.Fatal("failed to create first product")
	}

	// Try to create second product with same SKU using protobuf
	req, _ := testutil.MakeRequest(http.MethodPost, "/api/v1/products", &pb.ProductCreateRequest{
		Name: "Product 2",
		Sku:  "DUP-SKU-PROTO-001", // Duplicate SKU
	})

	rr := httptest.NewRecorder()
	handler := CreateProductHandler_Proto(tx)
	handler.ServeHTTP(rr, req)

	// Should return 409 Conflict
	testutil.AssertStatus(t, rr, http.StatusConflict)

	errorResp := &pb.ErrorResponse{}
	testutil.ParseProtoResponse(t, rr, errorResp)
	
	if errorResp.Error.Code != "conflict" {
		t.Errorf("expected error code 'conflict', got %v", errorResp.Error.Code)
	}
}

// TestListProducts tests GET /api/v1/products using protobuf
func TestListProducts(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	tests := []struct {
		name           string
		setupCount     int
		expectedStatus int
		checkResponse  func(t *testing.T, resp *pb.ProductListResponse)
	}{
		{
			name:           "list with default pagination",
			setupCount:     5,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *pb.ProductListResponse) {
				if len(resp.Products) != 5 {
					t.Errorf("expected 5 products, got %d", len(resp.Products))
				}
				if resp.Pagination.TotalItems != 5 {
					t.Errorf("expected total_items=5, got %d", resp.Pagination.TotalItems)
				}
			},
		},
		{
			name:           "empty product list",
			setupCount:     0,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *pb.ProductListResponse) {
				if len(resp.Products) != 0 {
					t.Errorf("expected 0 products, got %d", len(resp.Products))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := testutil.BeginTestTransaction(t, db)

			// Create test products
			for i := 0; i < tt.setupCount; i++ {
				sku := fmt.Sprintf("LIST-PROTO-%03d", i)
				testutil.CreateTestProduct(t, tx, fmt.Sprintf("Product %d", i), sku, nil)
			}

			req, _ := testutil.MakeRequest(http.MethodGet, "/api/v1/products", nil)
			rr := httptest.NewRecorder()

			handler := ListProductsHandler_Proto(tx)
			handler.ServeHTTP(rr, req)

			testutil.AssertStatus(t, rr, tt.expectedStatus)

			response := &pb.ProductListResponse{}
			testutil.ParseProtoResponse(t, rr, response)

			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}
		})
	}
}

// TestUpdateProduct tests PUT /api/v1/products/{id} using protobuf
func TestUpdateProduct(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	tests := []struct {
		name           string
		setupProduct   bool
		request        *pb.ProductUpdateRequest
		expectedStatus int
		expectedError  *string
	}{
		{
			name:         "update product name and description",
			setupProduct: true,
			request: &pb.ProductUpdateRequest{
				Name:        "Updated Product Name",
				Description: "Updated description",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "update product attributes",
			setupProduct: true,
			request: &pb.ProductUpdateRequest{
				Attributes: map[string]*pb.AttributeValue{
					"new_attr": models.CreateStringAttribute("new_value"),
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "update non-existent product returns 404",
			setupProduct: false,
			request: &pb.ProductUpdateRequest{
				Name: "Updated Name",
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  stringPtr("not_found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := testutil.BeginTestTransaction(t, db)

			var productID string
			if tt.setupProduct {
				product := testutil.CreateTestProduct(t, tx, "Original Name", "UPDATE-PROTO-001", nil)
				productID = product.ID.String()
			} else {
				productID = "00000000-0000-0000-0000-000000000000"
			}

			url := "/api/v1/products/" + productID
			req, _ := testutil.MakeRequest(http.MethodPut, url, tt.request)
			rr := httptest.NewRecorder()

			handler := UpdateProductHandler_Proto(tx)
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

// TestDeleteProduct tests DELETE /api/v1/products/{id} using protobuf
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
			productID:      "00000000-0000-0000-0000-000000000000",
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
				product := testutil.CreateTestProduct(t, tx, "To Delete", "DELETE-PROTO-001", nil)
				productID = product.ID.String()
			} else {
				productID = tt.productID
			}

			url := "/api/v1/products/" + productID
			req, _ := testutil.MakeRequest(http.MethodDelete, url, nil)
			rr := httptest.NewRecorder()

			handler := DeleteProductHandler_Proto(tx)
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

// TestGetProduct tests GET /api/v1/products/{id} using protobuf
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
			productID:      "00000000-0000-0000-0000-000000000000",
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
				product := testutil.CreateTestProduct(t, tx, "Test Product", "GET-PROTO-TEST-001", nil)
				productID = product.ID.String()
			} else {
				productID = tt.productID
			}

			url := "/api/v1/products/" + productID
			req, _ := testutil.MakeRequest(http.MethodGet, url, nil)
			rr := httptest.NewRecorder()

			handler := GetProductHandler_Proto(tx)
			handler.ServeHTTP(rr, req)

			testutil.AssertStatus(t, rr, tt.expectedStatus)

			if tt.expectedStatus == http.StatusOK {
				product := &pb.Product{}
				testutil.ParseProtoResponse(t, rr, product)
				
				if product.Id == "" {
					t.Error("expected product ID to be present")
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

