package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr        string
	DatabaseURL string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		Addr:        getenv("ADDR", ":8080"),
		DatabaseURL: getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/idsai?sslmode=disable"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
