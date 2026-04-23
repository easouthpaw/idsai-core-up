package main

import (
	"path/filepath"
	"testing"
)

func TestResolveMigrationsDirUsesExplicitEnv(t *testing.T) {
	t.Setenv("MIGRATIONS_DIR", " custom/migrations ")

	if got := resolveMigrationsDir(); got != "custom/migrations" {
		t.Fatalf("resolveMigrationsDir() = %q, want explicit env path", got)
	}
}

func TestResolveMigrationsDirFindsRepoMigrations(t *testing.T) {
	t.Setenv("MIGRATIONS_DIR", "")

	got := resolveMigrationsDir()
	want := filepath.Join("..", "..", migrationsDir)
	if got != migrationsDir && got != want && got != filepath.Join("/app", migrationsDir) {
		t.Fatalf("resolveMigrationsDir() = %q, want known migrations candidate", got)
	}
}

func TestResolveMigrationsDirFallsBackWhenCandidatesAreMissing(t *testing.T) {
	t.Setenv("MIGRATIONS_DIR", "")
	t.Chdir(t.TempDir())

	if got := resolveMigrationsDir(); got != "./"+migrationsDir {
		t.Fatalf("resolveMigrationsDir() = %q, want fallback migrations path", got)
	}
}
