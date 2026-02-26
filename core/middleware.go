package core

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/slice-soft/ss-keel-core/logger"
)

// keelLogger replaces Fiber's default logger middleware with one
// that uses Keel's logger for consistent log formatting.
func keelLogger(log *logger.Logger) fiber.Handler {
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

		return err
	}
}
