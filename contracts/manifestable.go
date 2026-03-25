package contracts

// EnvVar describes a single environment variable declared by an addon.
type EnvVar struct {
	Key         string
	Description string
	Required    bool
	Secret      bool
	Default     string
	Source      string // addon ID that declares this variable
}

// AddonManifest describes the capabilities, resources, and env vars of an addon.
// The CLI reads this to merge metadata into keel.toml — the addon never writes
// to keel.toml directly.
type AddonManifest struct {
	ID           string
	Version      string
	Capabilities []string // e.g. "database", "cache", "queue", "auth", "scheduler"
	Resources    []string // e.g. "postgres", "redis", "mongodb"
	EnvVars      []EnvVar
}

// Manifestable is implemented by addons that expose their metadata
// so the CLI and core can merge it into keel.toml.
type Manifestable interface {
	Manifest() AddonManifest
}
