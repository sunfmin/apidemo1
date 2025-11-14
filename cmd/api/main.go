package main

import (
	"log"
	"net/http"
	"os"

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

	// API routes will be mounted here
	r.Route("/api/v1", func(r chi.Router) {
		// Product routes will be added in Phase 3
		// r.Mount("/products", productsRouter(db))
		
		// Variant routes will be added in Phase 4
		// r.Mount("/variants", variantsRouter(db))
		
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

