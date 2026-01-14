package main

import (
	"gb-api/api"
	"gb-api/api/middleware"
	"log/slog"
	"net/http"
)

func addRoutes(m *http.ServeMux) {
	// Health check endpoint for load balancers and orchestration
	m.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	m.HandleFunc("/bigwig", middleware.CORSMiddleware(middleware.RateLimitMiddleware(api.BigWigHandler)))
	m.HandleFunc("/bigbed", middleware.CORSMiddleware(middleware.RateLimitMiddleware(api.BigBedHandler)))
	m.HandleFunc("/transcript", middleware.CORSMiddleware(middleware.RateLimitMiddleware(api.TranscriptHandler)))
	m.HandleFunc("/browser", middleware.CORSMiddleware(middleware.RateLimitMiddleware(api.BrowserHandler)))
	m.HandleFunc("/admin/cache-status", api.CacheSizeHandler)
}

func main() {
	mux := http.NewServeMux()

	addRoutes(mux)
	slog.Info("Server starting", "port", 8080)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		slog.Error("Failed to start server", "error", err)
	}
}
