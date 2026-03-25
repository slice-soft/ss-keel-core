package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseApplicationProperties(t *testing.T) {
	t.Setenv("SERVICE_NAME", "env-service")

	values, err := parseApplicationProperties(`
# app
app.name=${SERVICE_NAME:demo}
jwt.issuer=${app.name}
jwt.secret=${JWT_SECRET:change-me}
`)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if values["app.name"] != "env-service" {
		t.Fatalf("app.name = %q, want %q", values["app.name"], "env-service")
	}
	if values["jwt.issuer"] != "env-service" {
		t.Fatalf("jwt.issuer = %q, want %q", values["jwt.issuer"], "env-service")
	}
	if values["jwt.secret"] != "change-me" {
		t.Fatalf("jwt.secret = %q, want %q", values["jwt.secret"], "change-me")
	}
}

func TestLoadApplicationProperties_WalksUp(t *testing.T) {
	resetApplicationPropertiesForTests()

	root := t.TempDir()
	nested := filepath.Join(root, "internal", "modules")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(root, applicationPropertiesFile), []byte("app.name=demo\n"), 0644); err != nil {
		t.Fatalf("failed to write application.properties: %v", err)
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	if err := os.Chdir(nested); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	if err := LoadApplicationProperties(); err != nil {
		t.Fatalf("LoadApplicationProperties returned error: %v", err)
	}

	if got := GetString("app.name"); got != "demo" {
		t.Fatalf("GetString() = %q, want %q", got, "demo")
	}
}

func TestParseApplicationProperties_InvalidLine(t *testing.T) {
	_, err := parseApplicationProperties("invalid-line")
	if err == nil {
		t.Fatal("expected parse error for invalid line")
	}
}

func resetApplicationPropertiesForTests() {
	propertiesMu.Lock()
	defer propertiesMu.Unlock()

	propertiesLoaded = false
	propertiesValues = map[string]string{}
}
