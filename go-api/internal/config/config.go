package config

import (
	"os"
	"strings"
	"time"
)

// Config is process configuration loaded from the environment.
type Config struct {
	Port        string
	NodeAPIURL  string
	JWTSecret   string
	JWTUsername string
	JWTPassword string
	JWTTTL      time.Duration
	CORSOrigins []string
}

// FromEnv reads process settings, with local-development defaults.
func FromEnv() Config {
	return Config{
		Port:        getenv("PORT", "8080"),
		NodeAPIURL:  getenv("NODE_API_URL", "http://localhost:3000"),
		JWTSecret:   getenv("JWT_SECRET", "dev-only-change-me-please"),
		JWTUsername: getenv("JWT_USERNAME", "admin"),
		JWTPassword: getenv("JWT_PASSWORD", "admin"),
		JWTTTL:      parseTTL(getenv("JWT_TTL", "1h")),
		CORSOrigins: parseList(getenv("CORS_ORIGINS", "http://localhost:5173")),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseList(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseTTL(raw string) time.Duration {
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return time.Hour
	}
	return d
}
