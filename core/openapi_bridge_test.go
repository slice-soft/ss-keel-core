package core

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/slice-soft/ss-keel-core/core/httpx"
)

func TestToBuildInputMapsDocsConfig(t *testing.T) {
	cfg := applyDefaults(KConfig{
		ServiceName: "Orders API",
		Docs: DocsConfig{
			Title:       "Orders API Docs",
			Version:     "2.1.0",
			Description: "Public API",
			Contact: &DocsContact{
				Name:  "API Team",
				URL:   "https://example.com/contact",
				Email: "team@example.com",
			},
			License: &DocsLicense{
				Name: "MIT",
				URL:  "https://example.com/license",
			},
			Servers: []string{
				"https://api.example.com - production",
				"http://localhost:3000",
			},
			Tags: []DocsTag{
				{Name: "users", Description: "User operations"},
				{Name: "system", Description: "System endpoints"},
			},
		},
	})

	routes := []httpx.Route{
		httpx.GET("/users/:id", httpx.WrapHandler(func(c *httpx.Ctx) error { return c.NoContent() })).
			Describe("Get user").
			Tag("users"),
	}

	got := toBuildInput(cfg, routes)
	if got.Title != "Orders API Docs" || got.Version != "2.1.0" || got.Description != "Public API" {
		t.Fatalf("unexpected header fields: %+v", got)
	}
	if got.Contact == nil || got.Contact.Name != "API Team" {
		t.Fatalf("contact mapping failed: %+v", got.Contact)
	}
	if got.License == nil || got.License.Name != "MIT" {
		t.Fatalf("license mapping failed: %+v", got.License)
	}
	if len(got.Servers) != 2 {
		t.Fatalf("servers len = %d, want 2", len(got.Servers))
	}
	if got.Servers[0].URL != "https://api.example.com" || got.Servers[0].Description != "production" {
		t.Fatalf("server[0] mapping failed: %+v", got.Servers[0])
	}
	if got.Servers[1].URL != "http://localhost:3000" || got.Servers[1].Description != "" {
		t.Fatalf("server[1] mapping failed: %+v", got.Servers[1])
	}
	if len(got.Tags) != 2 || got.Tags[0].Name != "users" || got.Tags[1].Name != "system" {
		t.Fatalf("tags mapping failed: %+v", got.Tags)
	}
	if len(got.Routes) != 1 || got.Routes[0].Method != "GET" || got.Routes[0].Path != "/users/:id" {
		t.Fatalf("routes mapping failed: %+v", got.Routes)
	}
}

func TestToOpenAPIRoutesMapsRouteMetadata(t *testing.T) {
	type requestDTO struct {
		Name string `json:"name"`
	}
	type responseDTO struct {
		ID string `json:"id"`
	}

	route := httpx.POST("/users", httpx.WrapHandler(func(c *httpx.Ctx) error { return c.NoContent() })).
		WithBody(httpx.WithBody[requestDTO]()).
		WithResponse(httpx.WithResponse[responseDTO](http.StatusCreated)).
		Describe("Create user", "Creates a new user").
		Tag("users").
		WithSecured("bearerAuth", "apiKey").
		WithQueryParam("source", "string", false, "source system").
		WithDeprecated()

	out := toOpenAPIRoutes([]httpx.Route{route})
	if len(out) != 1 {
		t.Fatalf("len(out) = %d, want 1", len(out))
	}
	got := out[0]

	if got.Method != "POST" || got.Path != "/users" {
		t.Fatalf("method/path mapping failed: %+v", got)
	}
	if got.Summary != "Create user" || got.Description != "Creates a new user" {
		t.Fatalf("summary/description mapping failed: %+v", got)
	}
	if !got.Deprecated {
		t.Fatal("deprecated flag mapping failed")
	}
	if len(got.Tags) != 1 || got.Tags[0] != "users" {
		t.Fatalf("tags mapping failed: %+v", got.Tags)
	}
	if len(got.Secured) != 2 || got.Secured[0] != "bearerAuth" || got.Secured[1] != "apiKey" {
		t.Fatalf("secured mapping failed: %+v", got.Secured)
	}
	if got.StatusCode != http.StatusCreated {
		t.Fatalf("status code mapping failed: %d", got.StatusCode)
	}
	if reflect.TypeOf(got.Body) != reflect.TypeOf(requestDTO{}) {
		t.Fatalf("body type mapping failed: %T", got.Body)
	}
	if reflect.TypeOf(got.Response) != reflect.TypeOf(responseDTO{}) {
		t.Fatalf("response type mapping failed: %T", got.Response)
	}
	if len(got.QueryParams) != 1 || got.QueryParams[0].Name != "source" || got.QueryParams[0].Type != "string" {
		t.Fatalf("query params mapping failed: %+v", got.QueryParams)
	}
}
