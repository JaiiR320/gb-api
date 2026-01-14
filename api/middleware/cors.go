package middleware

import "net/http"

// CORSMiddleware handles CORS headers and OPTIONS preflight requests
// NOTE: Using "*" for Access-Control-Allow-Origin because this is a public API
// serving genomic data with no authentication. This allows third-party developers
// to build tools using this API and enables access from any origin including the
// Vercel-hosted frontend. If authentication or sensitive user data is added in
// the future, this should be restricted to specific trusted origins.
func CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight OPTIONS request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next(w, r)
	}
}
