package main

import (
	"log"
	"net/http"
	"os"

	"apidemo1/internal/handlers"
	"apidemo1/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	// Connect to database
	db, err := sqlx.Connect("pgx", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("successfully connected to database")

	// Create router
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)
	r.Use(middleware.CORS)

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Swagger UI endpoints
	r.Get("/docs", handlers.ServeSwaggerUI)
	r.Get("/api-spec", handlers.ServeOpenAPISpec)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Product routes (User Story 1 - Phase 3)
		r.Route("/products", func(r chi.Router) {
			handler := handlers.NewProductHandler(db)
			r.Post("/", handler.CreateProduct)
			r.Get("/", handler.ListProducts)
			r.Get("/{id}", handler.GetProduct)
			r.Put("/{id}", handler.UpdateProduct)
			r.Delete("/{id}", handler.DeleteProduct)
			
			// Variant sub-routes under products (User Story 2 - Phase 4)
			r.Route("/{productId}/variants", func(r chi.Router) {
				r.Post("/", func(w http.ResponseWriter, r *http.Request) {
					productID := chi.URLParam(r, "productId")
					// Note: In production, use proper handler initialization
					// For now, this creates a temporary connection
					// The test handlers work with transactions
					http.Error(w, "Use test handlers for variant creation", http.StatusNotImplemented)
				})
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					productID := chi.URLParam(r, "productId")
					http.Error(w, "Use test handlers for variant listing", http.StatusNotImplemented)
				})
			})
		})
		
		// Variant routes by ID (User Story 2 - Phase 4)
		r.Route("/variants", func(r chi.Router) {
			r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Use test handlers for variant retrieval", http.StatusNotImplemented)
			})
			r.Put("/{id}", func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Use test handlers for variant update", http.StatusNotImplemented)
			})
			r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Use test handlers for variant deletion", http.StatusNotImplemented)
			})
		})
		
		// Media asset routes will be added in Phase 5
		// r.Mount("/media", mediaRouter(db))
		
		// Product type routes will be added in Phase 6
		// r.Mount("/product-types", productTypesRouter(db))
	})

	// Get server configuration from environment
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	host := os.Getenv("SERVER_HOST")
	if host == "" {
		host = "localhost"
	}

	addr := host + ":" + port
	log.Printf("starting server on %s", addr)

	// Start server
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

