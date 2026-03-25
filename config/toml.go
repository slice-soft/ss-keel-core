package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// KeelTOML represents the structure of a keel.toml file.
type KeelTOML struct {
	Keel   KeelMeta    `toml:"keel"`
	Addons []AddonEntry `toml:"addons"`
	Env    []EnvDecl   `toml:"env"`
}

// KeelMeta holds top-level keel metadata.
type KeelMeta struct {
	Version string `toml:"version"`
}

// EnvDecl represents a single environment variable declaration in keel.toml.
type EnvDecl struct {
	Key      string `toml:"key"`
	Source   string `toml:"source"`
	Required bool   `toml:"required"`
	Secret   bool   `toml:"secret"`
	Default  string `toml:"default"`
}

// AddonEntry represents an installed addon entry in keel.toml.
type AddonEntry struct {
	ID           string   `toml:"id"`
	Version      string   `toml:"version"`
	Capabilities []string `toml:"capabilities"`
	Resources    []string `toml:"resources"`
}

// LoadKeelTOML finds and parses the nearest keel.toml file, walking up from the
// current working directory. Returns an error if no keel.toml is found — every
// Keel project must have one at the project root.
func LoadKeelTOML() (KeelTOML, error) {
	dir, err := os.Getwd()
	if err != nil {
		return KeelTOML{}, fmt.Errorf("could not determine working directory: %w", err)
	}
	return loadKeelTOMLFromDir(dir)
}

// loadKeelTOMLFromDir finds and parses keel.toml starting from dir, walking up.
// Returns an error if no keel.toml is found.
func loadKeelTOMLFromDir(dir string) (KeelTOML, error) {
	path, found := findKeelTOML(dir)
	if !found {
		return KeelTOML{}, fmt.Errorf("keel.toml not found: every Keel project requires a keel.toml at its root")
	}

	var cfg KeelTOML
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return KeelTOML{}, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return cfg, nil
}

// findKeelTOML walks up from dir looking for a keel.toml file.
// Returns the path and true if found, or empty string and false if not.
func findKeelTOML(dir string) (string, bool) {
	for {
		candidate := filepath.Join(dir, "keel.toml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}
