package contracts

// Well-known panel category constants for PanelEvent.Category.
//
// PanelEvent.Category is a free-form string — any addon can define its own
// category value without modifying this package. These constants represent the
// categories that the devpanel renders with a specialized view out of the box.
// Any other value falls back to the generic key/value renderer.
const (
	CategoryQuery   = "query"   // database queries
	CategoryAuth    = "auth"    // authentication and authorization flows
	CategoryCache   = "cache"   // cache operations (hit, miss, eviction)
	CategoryRequest = "request" // incoming HTTP requests
	CategoryGeneric = "generic" // catch-all for custom or third-party addons
)
