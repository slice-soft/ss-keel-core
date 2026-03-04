package httpx

import "github.com/gofiber/fiber/v2"

// QueryParamMeta documents a query string parameter in OpenAPI.
type QueryParamMeta struct {
	Name        string
	Type        string
	Description string
	Required    bool
}

// Route is the result of the route builder.
type Route struct {
	method      string
	path        string
	handler     fiber.Handler
	middlewares []fiber.Handler

	summary     string
	description string
	tags        []string
	secured     []string
	body        *BodyMeta
	response    *ResponseMeta
	queryParams []QueryParamMeta
	deprecated  bool
}

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

// WithBody creates a BodyMeta from a generic type.
func WithBody[T any]() *BodyMeta {
	var t T
	return &BodyMeta{Type: t, Required: true}
}

// WithResponse creates a ResponseMeta from a generic type and status code.
func WithResponse[T any](statusCode int) *ResponseMeta {
	var t T
	return &ResponseMeta{Type: t, StatusCode: statusCode}
}

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

// WithSecured documents the required security schemes in OpenAPI.
func (r Route) WithSecured(schemes ...string) Route {
	r.secured = append(r.secured, schemes...)
	return r
}

// Use adds execution middlewares to the route.
func (r Route) Use(middlewares ...fiber.Handler) Route {
	r.middlewares = append(r.middlewares, middlewares...)
	return r
}

// PrependMiddlewares prepends middlewares before existing route middlewares.
func (r Route) PrependMiddlewares(middlewares ...fiber.Handler) Route {
	r.middlewares = append(append([]fiber.Handler{}, middlewares...), r.middlewares...)
	return r
}

// WithPathPrefix prepends a path prefix to the route path.
func (r Route) WithPathPrefix(prefix string) Route {
	r.path = prefix + r.path
	return r
}

// WithDeprecated marks the route as deprecated in OpenAPI documentation.
func (r Route) WithDeprecated() Route {
	r.deprecated = true
	return r
}

// WithQueryParam documents a query string parameter in OpenAPI.
func (r Route) WithQueryParam(name, typ string, required bool, desc ...string) Route {
	qp := QueryParamMeta{Name: name, Type: typ, Required: required}
	if len(desc) > 0 {
		qp.Description = desc[0]
	}
	r.queryParams = append(r.queryParams, qp)
	return r
}

func newRoute(method, path string, handler fiber.Handler) Route {
	return Route{
		method:  method,
		path:    path,
		handler: handler,
	}
}

// GET creates a GET route.
func GET(path string, handler fiber.Handler) Route {
	return newRoute("GET", path, handler)
}

// POST creates a POST route.
func POST(path string, handler fiber.Handler) Route {
	return newRoute("POST", path, handler)
}

// PUT creates a PUT route.
func PUT(path string, handler fiber.Handler) Route {
	return newRoute("PUT", path, handler)
}

// PATCH creates a PATCH route.
func PATCH(path string, handler fiber.Handler) Route {
	return newRoute("PATCH", path, handler)
}

// DELETE creates a DELETE route.
func DELETE(path string, handler fiber.Handler) Route {
	return newRoute("DELETE", path, handler)
}
