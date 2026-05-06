package core

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/slice-soft/ss-keel-core/contracts"
)

// keelLogger provides request logging and optional metrics collection for HTTP requests.
func (a *App) keelLogger() fiber.Handler {
	log := a.logger
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		status := resolveStatus(c, err)
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
			a.metricsCollector.RecordRequest(contracts.RequestMetrics{
				Method:     method,
				Path:       path,
				StatusCode: status,
				Duration:   duration,
			})
		}

		return err
	}
}

// resolveStatus returns the true HTTP status code for the request.
// c.Response().StatusCode() reads 200 before Fiber's error handler runs,
// so we inspect the returned error directly when one is present.
func resolveStatus(c *fiber.Ctx, err error) int {
	if err != nil {
		var ke *KError
		if errors.As(err, &ke) {
			return ke.StatusCode
		}
		if fe, ok := err.(*fiber.Error); ok {
			return fe.Code
		}
	}
	return c.Response().StatusCode()
}
