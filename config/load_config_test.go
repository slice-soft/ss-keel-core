package config

import "testing"

type propertyConfig struct {
	AppName string  `keel:"app.name"`
	Port    int     `keel:"server.port,required"`
	Debug   bool    `keel:"feature.debug"`
	Workers uint    `keel:"workers"`
	Ratio   float64 `keel:"limits.ratio"`
	Ignored string
	Skipped string `keel:"-"`
}

func TestLoadConfigWithLookup_LoadsTypedValues(t *testing.T) {
	lookup := func(key string) (string, bool) {
		values := map[string]string{
			"app.name":      "keel-api",
			"server.port":   "8080",
			"feature.debug": "true",
			"workers":       "4",
			"limits.ratio":  "1.5",
			"ignored.value": "x",
			"skipped.value": "y",
		}
		value, ok := values[key]
		return value, ok
	}

	cfg, err := loadConfigWithLookup[propertyConfig](lookup)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.AppName != "keel-api" {
		t.Fatalf("AppName = %q, want %q", cfg.AppName, "keel-api")
	}
	if cfg.Port != 8080 {
		t.Fatalf("Port = %d, want %d", cfg.Port, 8080)
	}
	if !cfg.Debug {
		t.Fatal("Debug = false, want true")
	}
	if cfg.Workers != 4 {
		t.Fatalf("Workers = %d, want %d", cfg.Workers, 4)
	}
	if cfg.Ratio != 1.5 {
		t.Fatalf("Ratio = %f, want %f", cfg.Ratio, 1.5)
	}
	if cfg.Ignored != "" {
		t.Fatal("Ignored should remain zero value")
	}
	if cfg.Skipped != "" {
		t.Fatal("Skipped should remain zero value")
	}
}

func TestLoadConfigWithLookup_ReportsMissingRequired(t *testing.T) {
	lookup := func(key string) (string, bool) {
		if key == "app.name" {
			return "keel-api", true
		}
		return "", false
	}

	_, err := loadConfigWithLookup[propertyConfig](lookup)
	if err == nil {
		t.Fatal("expected missing required error")
	}
	if err.Error() != "missing required environment variables: server.port" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadConfigWithLookup_ReportsInvalidType(t *testing.T) {
	lookup := func(key string) (string, bool) {
		values := map[string]string{
			"app.name":    "keel-api",
			"server.port": "not-a-number",
		}
		value, ok := values[key]
		return value, ok
	}

	_, err := loadConfigWithLookup[propertyConfig](lookup)
	if err == nil {
		t.Fatal("expected invalid type error")
	}
}

func TestIsDevAndIsProd(t *testing.T) {
	resetApplicationPropertiesForTests()
	t.Setenv("APP_ENV", "development")
	if !IsDev() {
		t.Fatal("IsDev() should be true for APP_ENV=development")
	}
	if IsProd() {
		t.Fatal("IsProd() should be false for APP_ENV=development")
	}

	resetApplicationPropertiesForTests()
	t.Setenv("APP_ENV", "production")
	if IsDev() {
		t.Fatal("IsDev() should be false for APP_ENV=production")
	}
	if !IsProd() {
		t.Fatal("IsProd() should be true for APP_ENV=production")
	}
}
