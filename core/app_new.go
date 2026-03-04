package core

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/slice-soft/ss-keel-core/logger"
)

// New creates a new App instance with the given configuration.
func New(cfg KConfig) *App {
	cfg = applyDefaults(cfg)
	log := logger.NewLogger(cfg.isProduction())

	app := &App{
		config: cfg,
		logger: log,
		tracer: noopTracer{},
	}

	app.fiber = app.buildFiber()

	if !cfg.DisableHealth {
		app.registerHealth()
	}

	return app
}

func (a *App) buildFiber() *fiber.App {
	f := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler:          a.errorHandler(),
	})

	f.Use(requestid.New())
	f.Use(a.keelLogger())
	f.Use(recover.New())
	f.Use(cors.New())
	f.Use(a.translatorMiddleware())

	return f
}

func (a *App) errorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		var ke *KError
		if errors.As(err, &ke) {
			a.logger.Warn("HTTP Error [%d]: %s", ke.StatusCode, ke.Message)
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
		a.logger.Warn("HTTP Error [%d]: %s", code, err.Error())
		return c.Status(code).JSON(fiber.Map{
			"status_code": code,
			"message":     err.Error(),
		})
	}
}

func (a *App) translatorMiddleware() fiber.Handler {
	// Inject translator into locals so Ctx.T() can access it.
	return func(c *fiber.Ctx) error {
		if a.translator != nil {
			c.Locals("_keel_translator", a.translator)
		}
		return c.Next()
	}
}
