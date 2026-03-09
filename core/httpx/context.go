package httpx

import (
	"github.com/gofiber/fiber/v2"
	"github.com/slice-soft/ss-keel-core/contracts"
	"github.com/slice-soft/ss-keel-core/validation"
)

// Ctx is the Keel wrapper over fiber.Ctx.
type Ctx struct {
	*fiber.Ctx
}

// WrapHandler converts a Keel-style handler into a Fiber handler.
func WrapHandler(h func(*Ctx) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return h(&Ctx{c})
	}
}

// ParseBody parses and validates the request body.
// Returns 400 if JSON is invalid, 422 if validation fails.
func (c *Ctx) ParseBody(dst any) error {
	if err := c.Ctx.BodyParser(dst); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status_code": 400,
			"message":     "invalid request body",
		})
		return fiber.ErrBadRequest
	}

	if errs := validation.Validate(dst); len(errs) > 0 {
		c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"status_code": 422,
			"message":     "validation error",
			"errors":      errs,
		})
		return fiber.ErrUnprocessableEntity
	}

	return nil
}

// SetUser stores the authenticated user in Fiber locals for later retrieval.
func (c *Ctx) SetUser(user any) {
	c.Locals("_keel_user", user)
}

// User retrieves the authenticated user previously stored by SetUser.
func (c *Ctx) User() any {
	return c.Locals("_keel_user")
}

// UserAs extracts the authenticated user from Fiber locals and type-asserts it to T.
func UserAs[T any](c *Ctx) (T, bool) {
	v, ok := c.Locals("_keel_user").(T)
	return v, ok
}

// Lang extracts the language from the Accept-Language header.
// Returns "en" if the header is absent or empty.
func (c *Ctx) Lang() string {
	lang := c.Get("Accept-Language")
	if lang == "" {
		return "en"
	}
	for i := 0; i < len(lang); i++ {
		if lang[i] == ',' || lang[i] == ';' {
			return lang[:i]
		}
	}
	return lang
}

// T translates a key using a translator stored in locals.
// Returns the key unchanged if no translator is registered.
func (c *Ctx) T(key string, args ...any) string {
	t, ok := c.Locals("_keel_translator").(contracts.Translator)
	if !ok || t == nil {
		return key
	}
	return t.T(c.Lang(), key, args...)
}

// OK responds with HTTP 200 and a JSON body.
func (c *Ctx) OK(data any) error {
	return c.Status(fiber.StatusOK).JSON(data)
}

// Created responds with HTTP 201 and a JSON body.
func (c *Ctx) Created(data any) error {
	return c.Status(fiber.StatusCreated).JSON(data)
}

// NoContent responds with HTTP 204 No Content.
func (c *Ctx) NoContent() error {
	return c.Status(fiber.StatusNoContent).Send(nil)
}

// NotFound responds with HTTP 404 and an optional message.
func (c *Ctx) NotFound(message ...string) error {
	msg := "resource not found"
	if len(message) > 0 {
		msg = message[0]
	}
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"status_code": 404,
		"message":     msg,
	})
}
