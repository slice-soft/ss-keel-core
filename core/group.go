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
		prefixed := route
		prefixed.path = g.prefix + route.path
		prefixed.middlewares = append(g.middlewares, route.middlewares...)
		g.app.routes = append(g.app.routes, prefixed)
		handlers := append(prefixed.middlewares, prefixed.handler)
		g.app.fiber.Add(prefixed.method, prefixed.path, handlers...)
		g.app.logger.Debug("Route registered: [%s] %s", prefixed.method, prefixed.path)
	}
}

// Use registers a module under the group.
func (g *Group) Use(m Module) {
	m.Register(g.app)
}
