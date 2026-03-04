package core

import (
	"github.com/gofiber/fiber/v2"
	"github.com/slice-soft/ss-keel-core/core/httpx"
)

// HTTP aliases keep core's public API stable while organizing HTTP concerns
// under core/httpx.
type (
	Ctx = httpx.Ctx

	PageQuery   = httpx.PageQuery
	Page[T any] = httpx.Page[T]

	QueryParamMeta = httpx.QueryParamMeta
	Route          = httpx.Route
	BodyMeta       = httpx.BodyMeta
	ResponseMeta   = httpx.ResponseMeta
)

// WrapHandler converts a Keel handler function into a Fiber handler.
func WrapHandler(h func(*Ctx) error) fiber.Handler {
	return httpx.WrapHandler(h)
}

// UserAs extracts the authenticated user stored in Fiber locals and type-asserts it to T.
func UserAs[T any](c *Ctx) (T, bool) {
	return httpx.UserAs[T](c)
}

// NewPage constructs a Page from a slice of data and pagination parameters.
func NewPage[T any](data []T, total, page, limit int) Page[T] {
	return httpx.NewPage(data, total, page, limit)
}

// WithBody creates a BodyMeta from a generic type.
func WithBody[T any]() *BodyMeta {
	return httpx.WithBody[T]()
}

// WithResponse creates a ResponseMeta from a generic type and status code.
func WithResponse[T any](statusCode int) *ResponseMeta {
	return httpx.WithResponse[T](statusCode)
}

// GET creates a GET route.
func GET(path string, handler func(*Ctx) error) Route {
	return httpx.GET(path, WrapHandler(handler))
}

// POST creates a POST route.
func POST(path string, handler func(*Ctx) error) Route {
	return httpx.POST(path, WrapHandler(handler))
}

// PUT creates a PUT route.
func PUT(path string, handler func(*Ctx) error) Route {
	return httpx.PUT(path, WrapHandler(handler))
}

// PATCH creates a PATCH route.
func PATCH(path string, handler func(*Ctx) error) Route {
	return httpx.PATCH(path, WrapHandler(handler))
}

// DELETE creates a DELETE route.
func DELETE(path string, handler func(*Ctx) error) Route {
	return httpx.DELETE(path, WrapHandler(handler))
}
