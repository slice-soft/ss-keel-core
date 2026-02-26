package core

import "github.com/gofiber/fiber/v2"

// QueryParamMeta documents a query string parameter in OpenAPI.
type QueryParamMeta struct {
	Name        string
	Type        string
	Description string
	Required    bool
}

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
	queryParams []QueryParamMeta
	deprecated  bool
}

// Method returns the HTTP method of the route.
func (r Route) Method() string { return r.method }

// Path returns the route path pattern.
func (r Route) Path() string { return r.path }

// Handler returns the Fiber handler function.
func (r Route) Handler() fiber.Handler { return r.handler }

// Middlewares returns the middleware handlers.
func (r Route) Middlewares() []fiber.Handler { return r.middlewares }

// Summary returns the OpenAPI summary.
func (r Route) Summary() string { return r.summary }

// Description returns the OpenAPI description.
func (r Route) Description() string { return r.description }

// Tags returns the OpenAPI tags.
func (r Route) Tags() []string { return r.tags }

// Secured returns the list of security schemes required.
func (r Route) Secured() []string { return r.secured }

// Body returns the request body metadata.
func (r Route) Body() *BodyMeta { return r.body }

// Response returns the response metadata.
func (r Route) Response() *ResponseMeta { return r.response }

// QueryParams returns the query parameter definitions.
func (r Route) QueryParams() []QueryParamMeta { return r.queryParams }

// Deprecated returns whether the route is marked as deprecated.
func (r Route) Deprecated() bool { return r.deprecated }

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

// WithBody sets the request body metadata for the route.
func (r Route) WithBody(b *BodyMeta) Route {
	r.body = b
	return r
}

// WithResponse sets the response metadata on the route.
func (r Route) WithResponse(res *ResponseMeta) Route {
	r.response = res
	return r
}

// Tag adds an OpenAPI tag to classify the route.
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

// WithDeprecated marks the route as deprecated in OpenAPI documentation.
func (r Route) WithDeprecated() Route {
	r.deprecated = true
	return r
}

// WithQueryParam documents a query string parameter in OpenAPI.
//
//	GET("/users").WithQueryParam("status", "string", false, "Filter by status")
func (r Route) WithQueryParam(name, typ string, required bool, desc ...string) Route {
	qp := QueryParamMeta{Name: name, Type: typ, Required: required}
	if len(desc) > 0 {
		qp.Description = desc[0]
	}
	r.queryParams = append(r.queryParams, qp)
	return r
}

// — HTTP method constructors —
// Accept both controller methods and inline functions.

// newRoute creates a new Route with the given HTTP method, path and handler.
func newRoute(method, path string, handler func(*Ctx) error) Route {
	return Route{
		method:  method,
		path:    path,
		handler: WrapHandler(handler),
	}
}

// GET creates a GET route.
func GET(path string, handler func(*Ctx) error) Route {
	return newRoute("GET", path, handler)
}

// POST creates a POST route.
func POST(path string, handler func(*Ctx) error) Route {
	return newRoute("POST", path, handler)
}

// PUT creates a PUT route.
func PUT(path string, handler func(*Ctx) error) Route {
	return newRoute("PUT", path, handler)
}

// PATCH creates a PATCH route.
func PATCH(path string, handler func(*Ctx) error) Route {
	return newRoute("PATCH", path, handler)
}

// DELETE creates a DELETE route.
func DELETE(path string, handler func(*Ctx) error) Route {
	return newRoute("DELETE", path, handler)
}
