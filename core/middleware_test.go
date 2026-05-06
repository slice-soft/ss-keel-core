package core

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestResolveStatus_noError(t *testing.T) {
	app := fiber.New()
	var captured int
	// Middleware must be registered before routes to apply to them.
	app.Use(func(c *fiber.Ctx) error {
		err := c.Next()
		captured = resolveStatus(c, err)
		return err
	})
	app.Get("/ok", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/ok", nil)
	app.Test(req)

	if captured != 200 {
		t.Fatalf("resolveStatus = %d, want 200", captured)
	}
}

func TestResolveStatus_KError(t *testing.T) {
	var captured int
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		},
	})
	app.Use(func(c *fiber.Ctx) error {
		err := c.Next()
		captured = resolveStatus(c, err)
		return err
	})
	app.Get("/missing", func(c *fiber.Ctx) error {
		return NotFound("not here")
	})

	req := httptest.NewRequest("GET", "/missing", nil)
	app.Test(req)

	if captured != 404 {
		t.Fatalf("resolveStatus = %d, want 404; must read KError.StatusCode before error handler runs", captured)
	}
}

func TestResolveStatus_fiberError(t *testing.T) {
	var captured int
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		err := c.Next()
		captured = resolveStatus(c, err)
		return err
	})
	app.Get("/deny", func(c *fiber.Ctx) error {
		return fiber.ErrForbidden
	})

	req := httptest.NewRequest("GET", "/deny", nil)
	app.Test(req)

	if captured != 403 {
		t.Fatalf("resolveStatus = %d, want 403", captured)
	}
}
