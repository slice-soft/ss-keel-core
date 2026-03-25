package contracts

// Addon is the base contract that every Keel addon must implement.
type Addon interface {
	// ID returns the unique identifier for this addon (e.g. "gorm", "redis").
	ID() string
}
