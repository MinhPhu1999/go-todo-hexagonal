package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfigReadsDotEnvFile(t *testing.T) {
	restoreEnv(t, "DB_NAME")
	restoreEnv(t, "HTTP_READ_TIMEOUT")
	restoreEnv(t, "MONGODB_URI")
	restoreEnv(t, "SERVER_READ_TIMEOUT")
	restoreEnv(t, "TODO_COLLECTION")
	restoreEnv(t, "JWT_SECRET")

	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	if err := os.WriteFile(envFile, []byte("DB_NAME=from-dot-env\nMONGODB_URI=mongodb://dot-env:27017\nTODO_COLLECTION=dot_env_todos\nHTTP_READ_TIMEOUT=17s\nJWT_SECRET=test-secret-with-enough-length\n"), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	cfg, err := LoadConfig(dir, "test")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.Database.Name != "from-dot-env" {
		t.Fatalf("Database.Name = %q, want %q", cfg.Database.Name, "from-dot-env")
	}
	if cfg.Database.URI != "mongodb://dot-env:27017" {
		t.Fatalf("Database.URI = %q, want %q", cfg.Database.URI, "mongodb://dot-env:27017")
	}
	if cfg.Database.TodoCollection != "dot_env_todos" {
		t.Fatalf("Database.TodoCollection = %q, want %q", cfg.Database.TodoCollection, "dot_env_todos")
	}
	if cfg.Server.ReadTimeout != 17*time.Second {
		t.Fatalf("Server.ReadTimeout = %s, want 17s", cfg.Server.ReadTimeout)
	}
}

func TestLoadConfigRejectsShortJWTSecret(t *testing.T) {
	restoreEnv(t, "JWT_SECRET")

	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	if err := os.WriteFile(envFile, []byte("JWT_SECRET=short\n"), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	if _, err := LoadConfig(dir, "test"); err == nil {
		t.Fatal("LoadConfig() error = nil, want jwt secret validation error")
	}
}

func restoreEnv(t *testing.T, key string) {
	t.Helper()

	value, ok := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}

	t.Cleanup(func() {
		if ok {
			_ = os.Setenv(key, value)
			return
		}
		_ = os.Unsetenv(key)
	})
}
