package core

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/slice-soft/ss-keel-core/openapi"
)

// Listen starts the HTTP server with graceful shutdown support.
func (a *App) Listen() error {
	a.registerDocsRoutes()

	if a.scheduler != nil {
		a.scheduler.Start()
	}

	a.printBanner()

	return a.serveWithGracefulShutdown()
}

func (a *App) registerDocsRoutes() {
	if !a.config.docsEnabled() {
		return
	}

	spec := openapi.Build(toBuildInput(a.config, a.routes))
	a.fiber.Get("/docs/openapi.json", func(c *fiber.Ctx) error {
		return c.JSON(spec)
	})
	a.fiber.Get(a.config.Docs.Path, openapi.SwaggerUIHandler("/docs/openapi.json"))
	a.logger.Info("Docs: http://localhost:%d%s", a.config.Port, a.config.Docs.Path)
}

func (a *App) serveWithGracefulShutdown() error {
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
		return a.shutdown()
	}
}

func (a *App) shutdown() error {
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

// printBanner prints the Keel service banner with service name, port and environment.
func (a *App) printBanner() {
	fmt.Printf(
		"\n  ⚓  K E E L\n  ──────────────────────────────\n  Service  : %s\n  Port     : %d\n  Env      : %s\n  ──────────────────────────────\n\n",
		a.config.ServiceName, a.config.Port, a.config.Env,
	)
}
