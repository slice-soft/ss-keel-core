package core

import "github.com/gofiber/fiber/v2"

// Route is the result of the builder.
type Route struct {
	method      string
	path        string
	handler     fiber.Handler
	middlewares []fiber.Handler

	// OpenAPI
	summary     string
	description string
	tags        []string
	secured     []string // security schemes: "bearerAuth", "apiKey", etc.
	body        *BodyMeta
	response    *ResponseMeta
}

// Getters internos
func (r Route) Method() string               { return r.method }
func (r Route) Path() string                 { return r.path }
func (r Route) Handler() fiber.Handler       { return r.handler }
func (r Route) Middlewares() []fiber.Handler { return r.middlewares }
func (r Route) Summary() string              { return r.summary }
func (r Route) Description() string          { return r.description }
func (r Route) Tags() []string               { return r.tags }
func (r Route) Secured() []string            { return r.secured }
func (r Route) Body() *BodyMeta              { return r.body }
func (r Route) Response() *ResponseMeta      { return r.response }

// BodyMeta describes the request body.
type BodyMeta struct {
	Type     any
	Required bool
}

// ResponseMeta describes the expected response.
type ResponseMeta struct {
	Type       any
	StatusCode int
}

// — Standalone generic helpers —
// Go does not allow type parameters on methods,
// so WithBody and WithResponse are standalone functions.

// WithBody creates a BodyMeta from a generic type.
// Accepts any struct, inline or imported.
//
//	core.WithBody[dto.CreateUserDTO]()
//	core.WithBody[struct{ Name string `json:"name"` }]()
func WithBody[T any]() *BodyMeta {
	var t T
	return &BodyMeta{Type: t, Required: true}
}

// WithResponse creates a ResponseMeta from a generic type and status code.
//
//	core.WithResponse[dto.UserDTO](201)
func WithResponse[T any](statusCode int) *ResponseMeta {
	var t T
	return &ResponseMeta{Type: t, StatusCode: statusCode}
}

// — Builder methods —

// WithBody declares the request body DTO. Accepts any struct, inline or imported.
func (r Route) WithBody(b *BodyMeta) Route {
	r.body = b
	return r
}

// Res sets the response metadata on the route.
func (r Route) Res(res *ResponseMeta) Route {
	r.response = res
	return r
}

// Tag adds a single OpenAPI tag to the route.
func (r Route) Tag(tag string) Route {
	r.tags = append(r.tags, tag)
	return r
}

// Describe adds summary and optionally description for OpenAPI.
func (r Route) Describe(summary string, description ...string) Route {
	r.summary = summary
	if len(description) > 0 {
		r.description = description[0]
	}
	return r
}

// Secured documents the required security schemes in OpenAPI.
// Each scheme corresponds to an authentication middleware.
// Example: .WithSecured("bearerAuth") → lock icon in Swagger UI
// Example: .WithSecured("bearerAuth", "apiKey") → multiple schemes
func (r Route) WithSecured(schemes ...string) Route {
	r.secured = append(r.secured, schemes...)
	return r
}

// Use adds execution middlewares to the route.
// Middlewares are NOT documented in OpenAPI — use WithSecured() for that.
func (r Route) Use(middlewares ...fiber.Handler) Route {
	r.middlewares = append(r.middlewares, middlewares...)
	return r
}

// — HTTP method constructors —
// Accept both controller methods and inline functions.

func newRoute(method, path string, handler func(*Ctx) error) Route {
	return Route{
		method:  method,
		path:    path,
		handler: WrapHandler(handler),
	}
}

// Constructor for GET method.
func GET(path string, handler func(*Ctx) error) Route {
	return newRoute("GET", path, handler)
}

// Constructor for POST method.
func POST(path string, handler func(*Ctx) error) Route {
	return newRoute("POST", path, handler)
}

// Constructor for PUT method.
func PUT(path string, handler func(*Ctx) error) Route {
	return newRoute("PUT", path, handler)
}

// Constructor for PATCH method.
func PATCH(path string, handler func(*Ctx) error) Route {
	return newRoute("PATCH", path, handler)
}

// Constructor for DELETE method.
func DELETE(path string, handler func(*Ctx) error) Route {
	return newRoute("DELETE", path, handler)
}
