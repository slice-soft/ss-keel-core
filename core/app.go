package core

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/slice-soft/ss-keel-core/logger"
)

type App struct {
	fiber            *fiber.App
	config           KConfig
	routes           []Route
	logger           *logger.Logger
	shutdownHooks    []func(context.Context) error
	scheduler        Scheduler
	metricsCollector MetricsCollector
	tracer           Tracer
	translator       Translator
	healthCheckers   []HealthChecker
}

// Logger returns the configured logger instance.
func (a *App) Logger() *logger.Logger { return a.logger }

// Fiber returns the underlying Fiber application instance.
func (a *App) Fiber() *fiber.App { return a.fiber }
