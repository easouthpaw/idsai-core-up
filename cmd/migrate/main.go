package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const migrationsDir = "migrations"

func main() {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		log.Fatal("DATABASE_URL must be set")
	}

	command := "up"
	if len(os.Args) > 1 {
		command = strings.TrimSpace(strings.ToLower(os.Args[1]))
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("set goose dialect: %v", err)
	}

	dir := resolveMigrationsDir()
	switch command {
	case "up":
		if err := goose.Up(db, dir); err != nil {
			log.Fatalf("apply migrations: %v", err)
		}
		log.Printf("migrations applied successfully from %s", dir)
	case "status":
		if err := goose.Status(db, dir); err != nil {
			log.Fatalf("migration status: %v", err)
		}
	default:
		log.Fatalf("unsupported migration command %q, expected one of: up, status", command)
	}
}

func resolveMigrationsDir() string {
	if dir := strings.TrimSpace(os.Getenv("MIGRATIONS_DIR")); dir != "" {
		return dir
	}

	candidates := []string{
		migrationsDir,
		filepath.Join("..", "..", migrationsDir),
		filepath.Join("/app", migrationsDir),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return fmt.Sprintf("./%s", migrationsDir)
}
