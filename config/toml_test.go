package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadKeelTOMLFromDir_NotFound(t *testing.T) {
	// Empty temp dir — no keel.toml anywhere up the chain from it
	// Use os.TempDir() directly so the walk doesn't find our project's keel.toml
	tmp := t.TempDir()

	_, err := loadKeelTOMLFromDir(tmp)
	if err == nil {
		t.Fatal("expected error when keel.toml is not found, got nil")
	}
}

func TestLoadKeelTOMLFromDir_ParsesCorrectly(t *testing.T) {
	tmp := t.TempDir()

	content := `
[keel]
version = "1.0.0"

[[env]]
key      = "APP_ENV"
source   = "core"
required = false
default  = "development"

[[env]]
key      = "DB_DSN"
source   = "gorm"
required = true
secret   = true

[[addons]]
id           = "gorm"
version      = "0.3.0"
capabilities = ["database"]
resources    = ["postgres"]
`
	if err := os.WriteFile(filepath.Join(tmp, "keel.toml"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write keel.toml: %v", err)
	}

	cfg, err := loadKeelTOMLFromDir(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Keel.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", cfg.Keel.Version, "1.0.0")
	}
	if len(cfg.Env) != 2 {
		t.Errorf("len(Env) = %d, want 2", len(cfg.Env))
	}
	if cfg.Env[0].Key != "APP_ENV" || cfg.Env[0].Default != "development" {
		t.Errorf("Env[0] = %+v, want APP_ENV with default=development", cfg.Env[0])
	}
	if !cfg.Env[1].Required || !cfg.Env[1].Secret {
		t.Errorf("Env[1] should be required and secret: %+v", cfg.Env[1])
	}
	if len(cfg.Addons) != 1 || cfg.Addons[0].ID != "gorm" {
		t.Errorf("Addons = %+v, want one gorm addon", cfg.Addons)
	}
}

func TestLoadKeelTOMLFromDir_WalksUp(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}

	content := "[keel]\nversion = \"2.0.0\"\n"
	if err := os.WriteFile(filepath.Join(root, "keel.toml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadKeelTOMLFromDir(nested)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Keel.Version != "2.0.0" {
		t.Errorf("Version = %q, want %q (found by walking up)", cfg.Keel.Version, "2.0.0")
	}
}

func TestLoadKeelTOMLFromDir_InvalidTOML(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "keel.toml"), []byte("not = valid [[[toml"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := loadKeelTOMLFromDir(tmp)
	if err == nil {
		t.Fatal("expected error for invalid TOML, got nil")
	}
}
