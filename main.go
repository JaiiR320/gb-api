package main

import (
	"context"
	"gb-api/api"
	"gb-api/api/middleware"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	// Server configuration
	serverPort         = ":8080"
	readTimeout        = 30 * time.Second
	writeTimeout       = 60 * time.Second // Longer for large genomic responses
	idleTimeout        = 120 * time.Second
	maxRequestBodySize = 1 << 20 // 1 MB max request body
	shutdownTimeout    = 30 * time.Second
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

// maxBytesMiddleware limits the size of request bodies
func maxBytesMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()
	addRoutes(mux)

	// Wrap with request body size limit
	handler := maxBytesMiddleware(mux)

	server := &http.Server{
		Addr:         serverPort,
		Handler:      handler,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	// Channel to listen for shutdown signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		slog.Info("Server starting", "port", serverPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	sig := <-shutdown
	slog.Info("Shutdown signal received", "signal", sig)

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Server stopped gracefully")
}
