package core

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestGroupPrefix(t *testing.T) {
	tests := []struct {
		name       string
		prefix     string
		routePath  string
		requestURL string
		wantCode   int
	}{
		{
			name:       "group prefix prepended to route",
			prefix:     "/v1",
			routePath:  "/users",
			requestURL: "/v1/users",
			wantCode:   http.StatusOK,
		},
		{
			name:       "request without prefix returns 404",
			prefix:     "/v1",
			routePath:  "/users",
			requestURL: "/users",
			wantCode:   http.StatusNotFound,
		},
		{
			name:       "nested prefix",
			prefix:     "/api/v2",
			routePath:  "/products",
			requestURL: "/api/v2/products",
			wantCode:   http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New(KConfig{DisableHealth: true})
			g := app.Group(tt.prefix)
			g.RegisterController(ControllerFunc(func() []Route {
				return []Route{
					GET(tt.routePath, func(c *Ctx) error { return c.OK(nil) }),
				}
			}))

			req := httptest.NewRequest("GET", tt.requestURL, nil)
			resp, err := app.Fiber().Test(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantCode {
				t.Errorf("StatusCode = %v, want %v", resp.StatusCode, tt.wantCode)
			}
		})
	}
}

func TestGroupMiddleware(t *testing.T) {
	t.Run("group middleware runs before handler", func(t *testing.T) {
		app := New(KConfig{DisableHealth: true})

		headerSet := false
		groupMiddleware := func(c *fiber.Ctx) error {
			headerSet = true
			c.Set("X-Group", "yes")
			return c.Next()
		}

		g := app.Group("/v1", groupMiddleware)
		g.RegisterController(ControllerFunc(func() []Route {
			return []Route{
				GET("/ping", func(c *Ctx) error { return c.OK(nil) }),
			}
		}))

		req := httptest.NewRequest("GET", "/v1/ping", nil)
		_, err := app.Fiber().Test(req)
		if err != nil {
			t.Fatal(err)
		}
		if !headerSet {
			t.Error("group middleware should have run")
		}
	})

	t.Run("group middleware prepended before route middlewares", func(t *testing.T) {
		app := New(KConfig{DisableHealth: true})

		order := []string{}
		groupMW := func(c *fiber.Ctx) error {
			order = append(order, "group")
			return c.Next()
		}
		routeMW := func(c *fiber.Ctx) error {
			order = append(order, "route")
			return c.Next()
		}

		g := app.Group("/v1", groupMW)
		g.RegisterController(ControllerFunc(func() []Route {
			return []Route{
				GET("/ping", func(c *Ctx) error { return c.OK(nil) }).Use(routeMW),
			}
		}))

		req := httptest.NewRequest("GET", "/v1/ping", nil)
		_, err := app.Fiber().Test(req)
		if err != nil {
			t.Fatal(err)
		}
		if len(order) != 2 || order[0] != "group" || order[1] != "route" {
			t.Errorf("middleware order = %v, want [group route]", order)
		}
	})
}

func TestGroupRoutesRegisteredInApp(t *testing.T) {
	app := New(KConfig{DisableHealth: true})
	g := app.Group("/api")
	g.RegisterController(ControllerFunc(func() []Route {
		return []Route{
			GET("/users", dummyHandler),
			POST("/users", dummyHandler),
		}
	}))

	if len(app.routes) != 2 {
		t.Errorf("app.routes len = %v, want 2", len(app.routes))
	}
	for _, r := range app.routes {
		if r.Path() != "/api/users" {
			t.Errorf("route path = %v, want /api/users", r.Path())
		}
	}
}

func TestGroupHealthCheckers(t *testing.T) {
	t.Run("all checkers UP returns 200", func(t *testing.T) {
		app := New(KConfig{ServiceName: "Test"})
		app.RegisterHealthChecker(&mockHealthChecker{name: "db", err: nil})
		app.RegisterHealthChecker(&mockHealthChecker{name: "cache", err: nil})

		req := httptest.NewRequest("GET", "/health", nil)
		resp, err := app.Fiber().Test(req)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want 200", resp.StatusCode)
		}

		var body map[string]any
		json.NewDecoder(resp.Body).Decode(&body)
		if body["status"] != "UP" {
			t.Errorf("status = %v, want UP", body["status"])
		}
		checks, ok := body["checks"].(map[string]any)
		if !ok {
			t.Fatal("checks should be present")
		}
		if checks["db"] != "UP" || checks["cache"] != "UP" {
			t.Errorf("checks = %v", checks)
		}
	})

	t.Run("failing checker returns 503 and DOWN status", func(t *testing.T) {
		app := New(KConfig{ServiceName: "Test"})
		app.RegisterHealthChecker(&mockHealthChecker{name: "db", err: nil})
		app.RegisterHealthChecker(&mockHealthChecker{name: "redis", err: NotFound("connection refused")})

		req := httptest.NewRequest("GET", "/health", nil)
		resp, err := app.Fiber().Test(req)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Errorf("StatusCode = %v, want 503", resp.StatusCode)
		}

		var body map[string]any
		json.NewDecoder(resp.Body).Decode(&body)
		if body["status"] != "DOWN" {
			t.Errorf("status = %v, want DOWN", body["status"])
		}
	})
}

// mockHealthChecker is a test double for HealthChecker.
type mockHealthChecker struct {
	name string
	err  error
}

func (m *mockHealthChecker) Name() string               { return m.name }
func (m *mockHealthChecker) Check(_ context.Context) error { return m.err }
