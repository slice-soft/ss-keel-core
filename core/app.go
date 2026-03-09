package core

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/slice-soft/ss-keel-core/contracts"
	"github.com/slice-soft/ss-keel-core/core/httpx"
	"github.com/slice-soft/ss-keel-core/logger"
)

type App struct {
	fiber            *fiber.App
	config           KConfig
	routes           []httpx.Route
	logger           *logger.Logger
	shutdownHooks    []func(context.Context) error
	scheduler        contracts.Scheduler
	metricsCollector contracts.MetricsCollector
	tracer           contracts.Tracer
	translator       contracts.Translator
	healthCheckers   []contracts.HealthChecker
}

// Logger returns the configured logger instance.
func (a *App) Logger() *logger.Logger { return a.logger }

// Fiber returns the underlying Fiber application instance.
func (a *App) Fiber() *fiber.App { return a.fiber }
