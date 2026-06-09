// Package config loads runtime configuration from environment variables.
package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all runtime configuration for the signaling server.
type Config struct {
	// Port is the TCP port the HTTP server listens on.
	Port string
	// JWTSecret signs and verifies login tokens (HS256).
	JWTSecret string
	// JWTTTLMinutes is how long an issued token stays valid.
	JWTTTLMinutes int
	// AllowedOrigins is the list of origins permitted for CORS and WebSocket
	// upgrades. A single "*" entry disables origin checking (dev only).
	AllowedOrigins []string
}

// Load reads configuration from the environment, applying sane defaults so the
// server runs out of the box in development.
func Load() Config {
	return Config{
		Port:           getEnv("PORT", "8080"),
		JWTSecret:      getEnv("JWT_SECRET", "dev-insecure-secret-change-me"),
		JWTTTLMinutes:  getEnvInt("JWT_TTL_MINUTES", 60),
		AllowedOrigins: splitCSV(getEnv("ALLOWED_ORIGINS", "*")),
	}
}

// AllowAllOrigins reports whether origin checking is disabled.
func (c Config) AllowAllOrigins() bool {
	return len(c.AllowedOrigins) == 1 && c.AllowedOrigins[0] == "*"
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
