package config

import (
	"log/slog"
	"os"
	"strings"
	"time"
)

type Config struct {
	Port          string
	AppEnv        string
	DatabaseURL   string
	AutoMigrate   bool
	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration
	CORSOrigins   []string
}

func Load() Config {
	return Config{
		Port:          getEnv("PORT", "8080"),
		AppEnv:        getEnv("APP_ENV", "development"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/parking?sslmode=disable"),
		AutoMigrate:   getBoolEnv("DB_AUTO_MIGRATE", true),
		JWTSecret:     getEnv("JWT_SECRET", "change-me-in-development"),
		JWTAccessTTL:  getDurationEnv("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL: getDurationEnv("JWT_REFRESH_TTL", 7*24*time.Hour),
		CORSOrigins:   getListEnv("CORS_ORIGINS", []string{"http://localhost:3000"}),
	}
}

func getBoolEnv(key string, fallback bool) bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if raw == "" {
		return fallback
	}
	return raw == "1" || raw == "true" || raw == "yes"
}

func (c Config) LogLevel() slog.Level {
	if c.AppEnv == "production" {
		return slog.LevelInfo
	}
	return slog.LevelDebug
}

func getEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return value
}

func getListEnv(key string, fallback []string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			values = append(values, value)
		}
	}
	if len(values) == 0 {
		return fallback
	}
	return values
}
