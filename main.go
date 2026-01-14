package main

import (
	"context"
	"gb-api/api"
	"gb-api/api/middleware"
	"gb-api/config"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// Global config loaded at startup
var cfg *config.Config

// API version prefix - change this to bump all endpoints
const apiVersion = "/v1"

func addRoutes(m *http.ServeMux) {
	// Health check endpoint for load balancers and orchestration (unversioned)
	m.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// API endpoints
	m.HandleFunc(apiVersion+"/bigwig", middleware.CORSMiddleware(middleware.RateLimitMiddleware(api.BigWigHandler)))
	m.HandleFunc(apiVersion+"/bigbed", middleware.CORSMiddleware(middleware.RateLimitMiddleware(api.BigBedHandler)))
	m.HandleFunc(apiVersion+"/transcript", middleware.CORSMiddleware(middleware.RateLimitMiddleware(api.TranscriptHandler)))
	m.HandleFunc(apiVersion+"/browser", middleware.CORSMiddleware(middleware.RateLimitMiddleware(api.BrowserHandler)))

	// Admin endpoints (unversioned)
	m.HandleFunc("/admin/cache-status", api.CacheSizeHandler)
}

// maxBytesMiddleware limits the size of request bodies
func maxBytesMiddleware(maxSize int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)
			next.ServeHTTP(w, r)
		})
	}
}

func main() {
	// Load configuration from environment
	cfg = config.Load()

	mux := http.NewServeMux()
	addRoutes(mux)

	// Chain middleware: security headers -> body size limit -> routes
	handler := middleware.SecurityHeadersMiddleware(
		maxBytesMiddleware(cfg.MaxRequestBody)(mux),
	)

	server := &http.Server{
		Addr:         cfg.Port,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// Channel to listen for shutdown signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		slog.Info("Server starting", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	sig := <-shutdown
	slog.Info("Shutdown signal received", "signal", sig)

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Server stopped gracefully")
}
