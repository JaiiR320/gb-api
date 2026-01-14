package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Server settings
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	MaxRequestBody  int64
	ShutdownTimeout time.Duration

	// Cache settings
	CacheSize int
}

// Default configuration values
var (
	DefaultPort            = ":8080"
	DefaultReadTimeout     = 30 * time.Second
	DefaultWriteTimeout    = 60 * time.Second
	DefaultIdleTimeout     = 120 * time.Second
	DefaultMaxRequestBody  = int64(1 << 20) // 1 MB
	DefaultShutdownTimeout = 30 * time.Second
	DefaultCacheSize       = 25
)

// Load reads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		Port:            getEnvOrDefault("PORT", DefaultPort),
		ReadTimeout:     getDurationEnv("READ_TIMEOUT", DefaultReadTimeout),
		WriteTimeout:    getDurationEnv("WRITE_TIMEOUT", DefaultWriteTimeout),
		IdleTimeout:     getDurationEnv("IDLE_TIMEOUT", DefaultIdleTimeout),
		MaxRequestBody:  getInt64Env("MAX_REQUEST_BODY", DefaultMaxRequestBody),
		ShutdownTimeout: getDurationEnv("SHUTDOWN_TIMEOUT", DefaultShutdownTimeout),
		CacheSize:       getIntEnv("CACHE_SIZE", DefaultCacheSize),
	}
}

// GetCacheSize returns the cache size from environment or default
// This can be called from init() functions
func GetCacheSize() int {
	return getIntEnv("CACHE_SIZE", DefaultCacheSize)
}

// getEnvOrDefault returns the environment variable value or a default
func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// getIntEnv returns an integer from environment or default
func getIntEnv(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

// getInt64Env returns an int64 from environment or default
func getInt64Env(key string, defaultVal int64) int64 {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	}
	return defaultVal
}

// getDurationEnv returns a duration from environment (in seconds) or default
func getDurationEnv(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if secs, err := strconv.Atoi(val); err == nil {
			return time.Duration(secs) * time.Second
		}
	}
	return defaultVal
}
