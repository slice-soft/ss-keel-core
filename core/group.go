package core

import "github.com/gofiber/fiber/v2"

// Group is a set of routes sharing a common path prefix and middlewares.
type Group struct {
	prefix      string
	middlewares []fiber.Handler
	app         *App
}

// Group creates a new route group with the given prefix and optional middlewares.
func (a *App) Group(prefix string, middlewares ...fiber.Handler) *Group {
	return &Group{prefix: prefix, middlewares: middlewares, app: a}
}

// RegisterController registers a controller's routes under the group prefix,
// prepending the group middlewares before each route's own middlewares.
func (g *Group) RegisterController(c Controller) {
	for _, route := range c.Routes() {
		prefixed := route.WithPathPrefix(g.prefix).PrependMiddlewares(g.middlewares...)
		g.app.routes = append(g.app.routes, prefixed)
		handlers := append(append([]fiber.Handler{}, prefixed.Middlewares()...), prefixed.Handler())
		g.app.fiber.Add(prefixed.Method(), prefixed.Path(), handlers...)
		g.app.logger.Debug("Route registered: [%s] %s", prefixed.Method(), prefixed.Path())
	}
}

// Use registers a module under the group.
func (g *Group) Use(m Module) {
	m.Register(g.app)
}
