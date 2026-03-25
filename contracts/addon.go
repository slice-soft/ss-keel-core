package contracts

// Addon is the base contract every addon must satisfy.
// It provides a stable identifier used by the runtime to locate
// registered addons (e.g. app.GetAddon("devpanel")).
type Addon interface {
	// ID returns the unique identifier for this addon (e.g. "gorm", "jwt", "redis").
	ID() string
}
