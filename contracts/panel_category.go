package contracts

// Panel category constants used in PanelEvent.Category.
const (
	CategoryQuery   = "query"   // gorm, mongo
	CategoryAuth    = "auth"    // jwt, oauth
	CategoryCache   = "cache"   // redis
	CategoryRequest = "request" // core
	CategoryGeneric = "generic" // fallback / third-party
)
