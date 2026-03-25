package contracts

// EnvVar describes a single environment variable declared by an addon.
type EnvVar struct {
	Key         string
	Description string
	Required    bool
	Secret      bool
	Default     string
	Source      string // name of the addon that declares it
}

// AddonManifest describes an addon's identity, capabilities, and requirements.
// The core uses this to merge addon metadata into keel.toml — never the addon directly.
type AddonManifest struct {
	ID           string
	Version      string
	Capabilities []string // "database","cache","queue","auth","scheduler"
	Resources    []string // "postgres","redis","mongodb","rabbitmq"
	EnvVars      []EnvVar
}

// Manifestable is the contract every addon must implement.
// The core uses Manifest() to merge addon metadata into keel.toml.
type Manifestable interface {
	Manifest() AddonManifest
}
