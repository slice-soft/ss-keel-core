package core

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

// keelLogger provides request logging and optional metrics collection for HTTP requests.
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
