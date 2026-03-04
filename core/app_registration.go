package core

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

// Use registers a module into the app.
func (a *App) Use(m Module) {
	m.Register(a)
}

// RegisterController registers all routes from a controller into the app.
func (a *App) RegisterController(c Controller) {
	for _, route := range c.Routes() {
		a.routes = append(a.routes, route)
		handlers := append(append([]fiber.Handler{}, route.Middlewares()...), route.Handler())
		a.fiber.Add(route.Method(), route.Path(), handlers...)
		a.logger.Debug("Route registered: [%s] %s", route.Method(), route.Path())
	}
}

// OnShutdown registers a hook that is called during graceful shutdown.
func (a *App) OnShutdown(fn func(context.Context) error) {
	a.shutdownHooks = append(a.shutdownHooks, fn)
}

// SetMetricsCollector sets the metrics collector.
func (a *App) SetMetricsCollector(mc MetricsCollector) {
	a.metricsCollector = mc
}

// SetTracer sets the tracer. If never called, a noop tracer is used.
func (a *App) SetTracer(t Tracer) {
	a.tracer = t
}

// Tracer returns the configured tracer (never nil).
func (a *App) Tracer() Tracer {
	return a.tracer
}

// SetTranslator sets the i18n translator.
func (a *App) SetTranslator(t Translator) {
	a.translator = t
}

// RegisterScheduler registers a scheduler that will be started in Listen()
// and stopped on shutdown.
func (a *App) RegisterScheduler(s Scheduler) {
	a.scheduler = s
	a.OnShutdown(func(ctx context.Context) error {
		s.Stop(ctx)
		return nil
	})
}
