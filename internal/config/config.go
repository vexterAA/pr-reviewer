package config

import "os"

type Config struct {
	HTTPPort string
	DBDSN    string
}

func Load() (*Config, error) {
	cfg := &Config{
		HTTPPort: getenvDefault("HTTP_PORT", "8080"),
		DBDSN:    getenvDefault("DB_DSN", "postgres://user:password@db:5432/pr_review?sslmode=disable"),
	}
	return cfg, nil
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
