package contracts

import (
	"time"

	"github.com/a-h/templ"
)

// PanelEvent is a single event emitted by a Debuggable addon.
//
// Category is a free-form string defined by each addon (e.g. "query", "auth",
// "cache"). The devpanel uses it to select a renderer; unknown values fall back
// to the generic key/value view. Addons should define their own Category
// constant locally — this package imposes no restrictions on the value.
//
// Level is one of "info", "warn", or "error".
// Detail must never contain sensitive data (passwords, tokens, secrets).
type PanelEvent struct {
	Timestamp time.Time
	AddonID   string
	Category  string
	Label     string
	Detail    map[string]any
	Level     string // "info", "warn", "error"
}

// Debuggable is the contract every addon must implement to participate
// in the Keel dev panel. It exposes identification and a stream of events.
type Debuggable interface {
	PanelID() string
	PanelLabel() string
	PanelEvents() <-chan PanelEvent
}

// DebuggableWithView extends Debuggable with a custom templ component
// that the panel renders instead of the default key/value table.
// Implementing this interface is optional.
type DebuggableWithView interface {
	Debuggable
	PanelView() templ.Component
}

// PanelRegistry is implemented by the devpanel addon.
// Addons auto-register themselves by calling RegisterAddon inside Register().
type PanelRegistry interface {
	RegisterAddon(d Debuggable)
}
