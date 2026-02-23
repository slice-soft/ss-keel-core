package core

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/slicesoft/ss-keel-core/logger"
	"github.com/slicesoft/ss-keel-core/openapi"
)

type App struct {
	fiber  *fiber.App
	config KConfig
	routes []Route
	logger *logger.Logger
}

func New(cfg KConfig) *App {
	cfg = applyDefaults(cfg)
	log := logger.NewLogger(cfg.isProduction())

	f := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
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
	f.Use(fiberLogger.New(fiberLogger.Config{
		Format: "[${ip}]:${port} - ${pid} - ${locals:requestid} - ${status} - ${method} ${path}\n",
	}))
	f.Use(recover.New())
	f.Use(cors.New())

	f.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "UP",
			"service": cfg.ServiceName,
		})
	})

	return &App{fiber: f, config: cfg, logger: log}
}

func (a *App) Use(m Module) {
	m.Register(a)
}

func (a *App) RegisterController(c Controller) {
	for _, route := range c.Routes() {
		a.routes = append(a.routes, route)
		handlers := append(route.middlewares, route.handler)
		a.fiber.Add(route.method, route.path, handlers...)
		a.logger.Debug("Route registered: [%s] %s", route.method, route.path)
	}
}

func (a *App) Listen() error {
	if a.config.docsEnabled() {
		spec := openapi.Build(openapi.BuildInput{
			Title:   a.config.Docs.Title,
			Version: a.config.Docs.Version,
			Routes:  toOpenAPIRoutes(a.routes),
		})
		a.fiber.Get("/docs/openapi.json", func(c *fiber.Ctx) error {
			return c.JSON(spec)
		})
		a.fiber.Get(a.config.Docs.Path, openapi.SwaggerUIHandler("/docs/openapi.json"))
		a.logger.Info("Docs: http://localhost:%d%s", a.config.Port, a.config.Docs.Path)
	}

	a.printBanner()
	return a.fiber.Listen(fmt.Sprintf(":%d", a.config.Port))
}

func (a *App) Logger() *logger.Logger { return a.logger }
func (a *App) Fiber() *fiber.App      { return a.fiber }

func (a *App) printBanner() {
	fmt.Printf("\n  ⚓  K E E L\n  ──────────────────────────────\n  Servicio : %s\n  Puerto   : %d\n  Entorno  : %s\n  ──────────────────────────────\n\n",
		a.config.ServiceName, a.config.Port, a.config.Env)
}
