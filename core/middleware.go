package core

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

// keelLogger is a method on App so it can access a.metricsCollector.
// It replaces Fiber's default logger middleware with one that uses Keel's
// logger for consistent log formatting and optional metrics recording.
func (a *App) keelLogger() fiber.Handler {
	log := a.logger
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		status := c.Response().StatusCode()
		method := c.Method()
		path := c.Path()
		ip := c.IP()
		rid := c.Locals("requestid")

		msg := fmt.Sprintf("%s %s %s [%d] %s (%dms)", ip, rid, method, status, path, duration.Milliseconds())

		if status >= 400 {
			log.Warn("HTTP %s", msg)
		} else {
			log.Info("HTTP %s", msg)
		}

		if a.metricsCollector != nil {
			a.metricsCollector.RecordRequest(RequestMetrics{
				Method:     method,
				Path:       path,
				StatusCode: status,
				Duration:   duration,
			})
		}

		return err
	}
}
