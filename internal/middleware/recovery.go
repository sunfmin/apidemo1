package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
)

// Recovery is a middleware that recovers from panics and returns a 500 error
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				log.Printf("PANIC: %v\n%s", err, debug.Stack())

				// Return 500 Internal Server Error
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				errorResponse := map[string]interface{}{
					"error": map[string]string{
						"code":    "internal_error",
						"message": "An unexpected error occurred",
					},
				}

				json.NewEncoder(w).Encode(errorResponse)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

