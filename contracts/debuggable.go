package contracts

import (
	"context"
	"io"
	"time"
)

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

// PanelComponent is the minimal interface satisfied by any templ.Component.
// The contracts package does not import a-h/templ to keep the dependency
// surface minimal — any templ.Component satisfies this interface via Go's
// structural typing.
type PanelComponent interface {
	Render(ctx context.Context, w io.Writer) error
}

// DebuggableWithView is an optional extension of Debuggable for addons that
// want to render a custom view in the dev panel.
// If an addon does not implement this, the panel falls back to its generic
// key/value table renderer using the Detail map from PanelEvent.
type DebuggableWithView interface {
	Debuggable
	PanelView() PanelComponent
}
