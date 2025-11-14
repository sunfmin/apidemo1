package handlers

import (
	"net/http"
	"os"
	"path/filepath"
)

// ServeAdminUI serves the PIM admin interface HTML
func ServeAdminUI(w http.ResponseWriter, r *http.Request) {
	// Read the admin HTML file
	adminPath := filepath.Join("web", "admin.html")
	content, err := os.ReadFile(adminPath)
	if err != nil {
		http.Error(w, "Admin interface not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

