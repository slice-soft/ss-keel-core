package core

// Module is the basic unit of organization.
// Groups controller, service and repository of a domain.
type Module interface {
	Register(app *App)
}

// Controller exposes the routes of a module.
type Controller interface {
	Routes() []Route
}

// ControllerFunc is a helper to create controllers from simple functions.
type ControllerFunc func() []Route

func (f ControllerFunc) Routes() []Route {
	return f()
}
