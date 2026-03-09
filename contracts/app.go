package contracts

// Module is the basic unit of organization for a host application.
// A is the application/container type exposed by the host package.
type Module[A any] interface {
	Register(app A)
}

// Controller exposes the routes of a module.
// R is the route type exposed by the host package.
type Controller[R any] interface {
	Routes() []R
}

// ControllerFunc is a helper to create controllers from simple functions.
type ControllerFunc[R any] func() []R

// Routes returns the routes produced by the function.
func (f ControllerFunc[R]) Routes() []R {
	return f()
}
