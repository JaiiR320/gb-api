package middleware

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Rate limiter configuration (loaded from env vars)
var (
	rateLimitRPS   float64 = 1.67 // ~100 requests per minute
	rateLimitBurst int     = 10   // Allow short bursts
)

// Client rate limiters stored by IP
var (
	limiters   = make(map[string]*clientLimiter)
	limitersMu sync.Mutex
)

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func init() {
	// Load config from environment
	if rps := os.Getenv("RATE_LIMIT_RPS"); rps != "" {
		if v, err := strconv.ParseFloat(rps, 64); err == nil {
			rateLimitRPS = v
		}
	}
	if burst := os.Getenv("RATE_LIMIT_BURST"); burst != "" {
		if v, err := strconv.Atoi(burst); err == nil {
			rateLimitBurst = v
		}
	}

	// Start cleanup goroutine
	go cleanupLimiters()
}

// cleanupLimiters removes stale limiters every 5 minutes
func cleanupLimiters() {
	for {
		time.Sleep(5 * time.Minute)
		limitersMu.Lock()
		for ip, cl := range limiters {
			if time.Since(cl.lastSeen) > 10*time.Minute {
				delete(limiters, ip)
			}
		}
		limitersMu.Unlock()
	}
}

// getLimiter returns a rate limiter for the given IP
func getLimiter(ip string) *rate.Limiter {
	limitersMu.Lock()
	defer limitersMu.Unlock()

	cl, exists := limiters[ip]
	if !exists {
		cl = &clientLimiter{
			limiter:  rate.NewLimiter(rate.Limit(rateLimitRPS), rateLimitBurst),
			lastSeen: time.Now(),
		}
		limiters[ip] = cl
	} else {
		cl.lastSeen = time.Now()
	}
	return cl.limiter
}

// getClientIP extracts client IP, respecting X-Forwarded-For for proxies
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (set by Fly.io, nginx, etc.)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP (original client)
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr (strip port)
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// RateLimitMiddleware limits requests per client IP using token bucket
func RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		limiter := getLimiter(ip)

		if !limiter.Allow() {
			slog.Warn("Rate limit exceeded", "ip", ip)
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next(w, r)
	}
}
