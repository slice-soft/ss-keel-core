package config

import (
	"testing"
)

type testAppConfig struct {
	AppEnv  string `keel:"APP_ENV"`
	DBHost  string `keel:"DB_HOST,required"`
	Port    int    `keel:"PORT,required"`
	Debug   bool   `keel:"DEBUG"`
	Workers uint   `keel:"WORKERS"`
	Ratio   float64 `keel:"RATIO"`
	Ignored string
	Skipped string `keel:"-"`
}

func TestLoadConfig_AllFromEnv(t *testing.T) {
	t.Setenv("APP_ENV", "staging")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("PORT", "8080")
	t.Setenv("DEBUG", "true")
	t.Setenv("WORKERS", "4")
	t.Setenv("RATIO", "1.5")

	cfg, err := loadConfigWithTOML[testAppConfig](KeelTOML{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.AppEnv != "staging" {
		t.Errorf("AppEnv = %q, want %q", cfg.AppEnv, "staging")
	}
	if cfg.DBHost != "localhost" {
		t.Errorf("DBHost = %q, want %q", cfg.DBHost, "localhost")
	}
	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want 8080", cfg.Port)
	}
	if !cfg.Debug {
		t.Error("Debug = false, want true")
	}
	if cfg.Workers != 4 {
		t.Errorf("Workers = %d, want 4", cfg.Workers)
	}
	if cfg.Ratio != 1.5 {
		t.Errorf("Ratio = %f, want 1.5", cfg.Ratio)
	}
	if cfg.Ignored != "" {
		t.Error("Ignored should remain zero — no keel tag")
	}
	if cfg.Skipped != "" {
		t.Error("Skipped should remain zero — keel:\"-\"")
	}
}

func TestLoadConfig_DefaultsFromTOML(t *testing.T) {
	t.Setenv("DB_HOST", "db.local")
	t.Setenv("PORT", "3000")
	// APP_ENV intentionally unset — default comes from keel.toml

	tomlCfg := KeelTOML{
		Env: []EnvDecl{
			{Key: "APP_ENV", Default: "development", Required: false},
		},
	}

	cfg, err := loadConfigWithTOML[testAppConfig](tomlCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.AppEnv != "development" {
		t.Errorf("AppEnv = %q, want %q (from keel.toml default)", cfg.AppEnv, "development")
	}
}

func TestLoadConfig_EnvOverridesToMLDefault(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DB_HOST", "prod-db")
	t.Setenv("PORT", "443")

	tomlCfg := KeelTOML{
		Env: []EnvDecl{
			{Key: "APP_ENV", Default: "development"},
		},
	}

	cfg, err := loadConfigWithTOML[testAppConfig](tomlCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.AppEnv != "production" {
		t.Errorf("AppEnv = %q, want %q (env overrides keel.toml default)", cfg.AppEnv, "production")
	}
}

func TestLoadConfig_MissingRequiredFromTag(t *testing.T) {
	// DB_HOST and PORT are required via struct tag — not set
	cfg, err := loadConfigWithTOML[testAppConfig](KeelTOML{})
	if err == nil {
		t.Fatalf("expected error for missing required vars, got cfg=%+v", cfg)
	}

	errMsg := err.Error()
	for _, key := range []string{"DB_HOST", "PORT"} {
		if !containsStr(errMsg, key) {
			t.Errorf("error message should mention %q, got: %s", key, errMsg)
		}
	}
}

func TestLoadConfig_MissingRequiredFromTOML(t *testing.T) {
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("PORT", "3000")
	// EXTRA_KEY is required via keel.toml, not via struct tag

	type configWithExtra struct {
		DBHost   string `keel:"DB_HOST,required"`
		Port     int    `keel:"PORT,required"`
		ExtraKey string `keel:"EXTRA_KEY"`
	}

	tomlCfg := KeelTOML{
		Env: []EnvDecl{
			{Key: "EXTRA_KEY", Required: true},
		},
	}

	_, err := loadConfigWithTOML[configWithExtra](tomlCfg)
	if err == nil {
		t.Fatal("expected error for EXTRA_KEY required in keel.toml")
	}
	if !containsStr(err.Error(), "EXTRA_KEY") {
		t.Errorf("error should mention EXTRA_KEY, got: %s", err.Error())
	}
}

func TestLoadConfig_InvalidFieldType(t *testing.T) {
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("PORT", "notanumber")

	_, err := loadConfigWithTOML[testAppConfig](KeelTOML{})
	if err == nil {
		t.Fatal("expected error for invalid integer value")
	}
}

func TestIsDev(t *testing.T) {
	tests := []struct {
		envValue string
		want     bool
	}{
		{"", true},
		{"development", true},
		{"production", false},
		{"staging", false},
	}

	for _, tt := range tests {
		t.Run("APP_ENV="+tt.envValue, func(t *testing.T) {
			if tt.envValue == "" {
				t.Setenv("APP_ENV", "")
			} else {
				t.Setenv("APP_ENV", tt.envValue)
			}
			if got := IsDev(); got != tt.want {
				t.Errorf("IsDev() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsProd(t *testing.T) {
	tests := []struct {
		envValue string
		want     bool
	}{
		{"production", true},
		{"development", false},
		{"", false},
		{"staging", false},
	}

	for _, tt := range tests {
		t.Run("APP_ENV="+tt.envValue, func(t *testing.T) {
			t.Setenv("APP_ENV", tt.envValue)
			if got := IsProd(); got != tt.want {
				t.Errorf("IsProd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
