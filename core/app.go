package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/slice-soft/ss-keel-core/logger"
	"github.com/slice-soft/ss-keel-core/openapi"
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

// New creates a new App instance with the given configuration.
func New(cfg KConfig) *App {
	cfg = applyDefaults(cfg)
	log := logger.NewLogger(cfg.isProduction())

	app := &App{
		config: cfg,
		logger: log,
		tracer: noopTracer{},
	}

	f := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			var ke *KError
			if errors.As(err, &ke) {
				log.Warn("HTTP Error [%d]: %s", ke.StatusCode, ke.Message)
				return c.Status(ke.StatusCode).JSON(fiber.Map{
					"status_code": ke.StatusCode,
					"code":        ke.Code,
					"message":     ke.Message,
				})
			}
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			log.Warn("HTTP Error [%d]: %s", code, err.Error())
			return c.Status(code).JSON(fiber.Map{
				"status_code": code,
				"message":     err.Error(),
			})
		},
	})

	f.Use(requestid.New())
	f.Use(app.keelLogger())
	f.Use(recover.New())
	f.Use(cors.New())

	// Inject translator into locals via middleware so Ctx.T() can access it.
	f.Use(func(c *fiber.Ctx) error {
		if app.translator != nil {
			c.Locals("_keel_translator", app.translator)
		}
		return c.Next()
	})

	app.fiber = f

	if !cfg.DisableHealth {
		app.registerHealth()
	}

	return app
}

// Use registers a module into the app.
func (a *App) Use(m Module) {
	m.Register(a)
}

// RegisterController registers all routes from a controller into the app.
func (a *App) RegisterController(c Controller) {
	for _, route := range c.Routes() {
		a.routes = append(a.routes, route)
		handlers := append(route.middlewares, route.handler)
		a.fiber.Add(route.method, route.path, handlers...)
		a.logger.Debug("Route registered: [%s] %s", route.method, route.path)
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

// Listen starts the HTTP server with graceful shutdown support.
func (a *App) Listen() error {
	if a.config.docsEnabled() {
		spec := openapi.Build(toBuildInput(a.config, a.routes))
		a.fiber.Get("/docs/openapi.json", func(c *fiber.Ctx) error {
			return c.JSON(spec)
		})
		a.fiber.Get(a.config.Docs.Path, openapi.SwaggerUIHandler("/docs/openapi.json"))
		a.logger.Info("Docs: http://localhost:%d%s", a.config.Port, a.config.Docs.Path)
	}

	if a.scheduler != nil {
		a.scheduler.Start()
	}

	a.printBanner()

	errCh := make(chan error, 1)
	go func() {
		errCh <- a.fiber.Listen(fmt.Sprintf(":%d", a.config.Port))
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-quit:
		a.logger.Info("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		for _, hook := range a.shutdownHooks {
			if err := hook(ctx); err != nil {
				a.logger.Warn("Shutdown hook error: %s", err.Error())
			}
		}

		return a.fiber.ShutdownWithContext(ctx)
	}
}

func (a *App) Logger() *logger.Logger { return a.logger }
func (a *App) Fiber() *fiber.App      { return a.fiber }

func (a *App) printBanner() {
	fmt.Printf(
		"\n  ⚓  K E E L\n  ──────────────────────────────\n  Service  : %s\n  Port     : %d\n  Env      : %s\n  ──────────────────────────────\n\n",
		a.config.ServiceName, a.config.Port, a.config.Env,
	)
}
