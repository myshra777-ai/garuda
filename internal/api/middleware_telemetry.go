package api

import (
	"net/http"
	"os"
)

// PassiveModeMiddleware adds passive mode detection.
func PassiveModeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mode := os.Getenv("GARUDA_MODE")
		if mode == "passive" {
			w.Header().Set("X-Garuda-Mode", "passive")
		} else {
			w.Header().Set("X-Garuda-Mode", "active")
		}
		next.ServeHTTP(w, r)
	})
}
