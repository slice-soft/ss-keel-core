package httpx

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestConstructors(t *testing.T) {
	handler := func(c *fiber.Ctx) error { return c.SendStatus(http.StatusAccepted) }
	tests := []struct {
		name   string
		route  Route
		method string
		path   string
	}{
		{name: "GET", route: GET("/users", handler), method: "GET", path: "/users"},
		{name: "POST", route: POST("/users", handler), method: "POST", path: "/users"},
		{name: "PUT", route: PUT("/users/:id", handler), method: "PUT", path: "/users/:id"},
		{name: "PATCH", route: PATCH("/users/:id", handler), method: "PATCH", path: "/users/:id"},
		{name: "DELETE", route: DELETE("/users/:id", handler), method: "DELETE", path: "/users/:id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.route.Method() != tt.method {
				t.Fatalf("Method() = %q, want %q", tt.route.Method(), tt.method)
			}
			if tt.route.Path() != tt.path {
				t.Fatalf("Path() = %q, want %q", tt.route.Path(), tt.path)
			}
			if tt.route.Handler() == nil {
				t.Fatal("Handler() is nil")
			}
		})
	}
}

func TestBuilderMetadata(t *testing.T) {
	type req struct {
		Name string `json:"name"`
	}
	type res struct {
		ID string `json:"id"`
	}

	route := POST("/users", func(c *fiber.Ctx) error { return c.SendStatus(http.StatusCreated) }).
		WithBody(WithBody[req]()).
		WithResponse(WithResponse[res](http.StatusCreated)).
		Tag("users").
		Tag("admin").
		Describe("Create user", "Creates a user").
		WithSecured("bearerAuth").
		WithDeprecated().
		WithQueryParam("source", "string", false, "Source system")

	if route.Body() == nil || reflect.TypeOf(route.Body().Type) != reflect.TypeOf(req{}) {
		t.Fatal("Body() not configured correctly")
	}
	if route.Response() == nil || route.Response().StatusCode != http.StatusCreated {
		t.Fatal("Response() not configured correctly")
	}
	if route.Summary() != "Create user" || route.Description() != "Creates a user" {
		t.Fatalf("Describe() not applied: summary=%q description=%q", route.Summary(), route.Description())
	}
	if !route.Deprecated() {
		t.Fatal("Deprecated() should be true")
	}
	if len(route.Tags()) != 2 || route.Tags()[0] != "users" || route.Tags()[1] != "admin" {
		t.Fatalf("Tags() = %v", route.Tags())
	}
	if len(route.Secured()) != 1 || route.Secured()[0] != "bearerAuth" {
		t.Fatalf("Secured() = %v", route.Secured())
	}
	qp := route.QueryParams()
	if len(qp) != 1 || qp[0].Name != "source" || qp[0].Type != "string" {
		t.Fatalf("QueryParams() = %v", qp)
	}
}

func TestMiddlewareOrderAndPathPrefix(t *testing.T) {
	order := []string{}

	mwGroup := func(c *fiber.Ctx) error {
		order = append(order, "group")
		return c.Next()
	}
	mwRoute := func(c *fiber.Ctx) error {
		order = append(order, "route")
		return c.Next()
	}
	handler := func(c *fiber.Ctx) error {
		order = append(order, "handler")
		return c.SendStatus(http.StatusNoContent)
	}

	route := GET("/ping", handler).
		Use(mwRoute).
		PrependMiddlewares(mwGroup).
		WithPathPrefix("/v1")

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	handlers := append(append([]fiber.Handler{}, route.Middlewares()...), route.Handler())
	app.Add(route.Method(), route.Path(), handlers...)

	resp, err := app.Test(httptest.NewRequest("GET", "/v1/ping", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}

	wantOrder := []string{"group", "route", "handler"}
	if !reflect.DeepEqual(order, wantOrder) {
		t.Fatalf("middleware/handler order = %v, want %v", order, wantOrder)
	}
}
