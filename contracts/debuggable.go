package contracts

import "time"

// PanelEvent is a single observable event emitted by an addon.
// Addons are free to use the Detail map to carry any domain-specific
// information (e.g. SQL query, cache hit, auth result). Core imposes
// no vocabulary on the contents of Detail.
type PanelEvent struct {
	Timestamp time.Time
	AddonID   string
	Label     string
	Detail    map[string]any
	Level     string // "info", "warn", "error"
}

// Debuggable is implemented by addons that stream observable events to the
// dev panel. The panel consumes the channel and owns all rendering decisions.
type Debuggable interface {
	PanelID() string
	PanelLabel() string
	PanelEvents() <-chan PanelEvent
}

// PanelRegistry is implemented by the dev panel addon.
// Debuggable addons call RegisterAddon during their own Register step.
type PanelRegistry interface {
	RegisterAddon(d Debuggable)
}
