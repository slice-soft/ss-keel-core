package core

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/slice-soft/ss-keel-core/openapi"
)

// Listen starts the HTTP server with graceful shutdown support.
func (a *App) Listen() error {
	if err := a.resolveListenPort(); err != nil {
		return err
	}

	a.registerDocsRoutes()

	a.printBanner()

	if a.scheduler != nil {
		a.scheduler.Start()
	}

	return a.serveWithGracefulShutdown()
}

func (a *App) resolveListenPort() error {
	const maxPortChecks = 100

	selected, err := firstAvailablePort(a.config.Port, maxPortChecks)
	if err != nil {
		return err
	}
	if selected != a.config.Port {
		a.logger.Warn("Port %d is in use, switching to %d", a.config.Port, selected)
		a.config.Port = selected
	}
	return nil
}

func firstAvailablePort(startPort, maxChecks int) (int, error) {
	if startPort < 1 || startPort > 65535 {
		return 0, fmt.Errorf("invalid listen port: %d", startPort)
	}
	if maxChecks <= 0 {
		return 0, fmt.Errorf("invalid maxChecks: %d", maxChecks)
	}

	port := startPort
	for i := 0; i < maxChecks && port <= 65535; i++ {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			_ = ln.Close()
			return port, nil
		}
		port++
	}

	return 0, fmt.Errorf("no available port found from %d after %d attempts", startPort, maxChecks)
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
