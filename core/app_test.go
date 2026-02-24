package core

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// testModule is a minimal Module implementation for testing.
type testModule struct {
	controller Controller
}

func (m *testModule) Register(app *App) {
	app.RegisterController(m.controller)
}

// testController is a minimal Controller implementation for testing.
type testController struct {
	routes []Route
}

func (c *testController) Routes() []Route {
	return c.routes
}

func TestNew(t *testing.T) {
	tests := []struct {
		name            string
		cfg             KConfig
		wantPort        int
		wantServiceName string
		wantEnv         string
	}{
		{
			name:            "empty config applies defaults",
			cfg:             KConfig{},
			wantPort:        3000,
			wantServiceName: "Keel App",
			wantEnv:         "development",
		},
		{
			name:            "custom config preserved",
			cfg:             KConfig{Port: 8080, ServiceName: "My API", Env: "staging"},
			wantPort:        8080,
			wantServiceName: "My API",
			wantEnv:         "staging",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New(tt.cfg)
			if app == nil {
				t.Fatal("New() returned nil")
			}
			if app.config.Port != tt.wantPort {
				t.Errorf("Port = %v, want %v", app.config.Port, tt.wantPort)
			}
			if app.config.ServiceName != tt.wantServiceName {
				t.Errorf("ServiceName = %v, want %v", app.config.ServiceName, tt.wantServiceName)
			}
			if app.config.Env != tt.wantEnv {
				t.Errorf("Env = %v, want %v", app.config.Env, tt.wantEnv)
			}
		})
	}
}

func TestHealthEndpoint(t *testing.T) {
	tests := []struct {
		name            string
		cfg             KConfig
		wantCode        int
		wantServiceName string
	}{
		{
			name:            "health returns UP with service name",
			cfg:             KConfig{ServiceName: "Test API"},
			wantCode:        http.StatusOK,
			wantServiceName: "Test API",
		},
		{
			name:            "health with default service name",
			cfg:             KConfig{},
			wantCode:        http.StatusOK,
			wantServiceName: "Keel App",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New(tt.cfg)

			req := httptest.NewRequest("GET", "/health", nil)
			resp, err := app.Fiber().Test(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantCode {
				t.Errorf("StatusCode = %v, want %v", resp.StatusCode, tt.wantCode)
			}

			var body map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["status"] != "UP" {
				t.Errorf("status = %v, want UP", body["status"])
			}
			if body["service"] != tt.wantServiceName {
				t.Errorf("service = %v, want %v", body["service"], tt.wantServiceName)
			}
		})
	}
}

func TestRegisterController(t *testing.T) {
	tests := []struct {
		name       string
		routes     []Route
		wantRoutes int
	}{
		{
			name: "registers single route",
			routes: []Route{
				GET("/users", dummyHandler),
			},
			wantRoutes: 1,
		},
		{
			name: "registers multiple routes",
			routes: []Route{
				GET("/users", dummyHandler),
				POST("/users", dummyHandler),
				DELETE("/users/:id", dummyHandler),
			},
			wantRoutes: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New(KConfig{})
			app.RegisterController(&testController{routes: tt.routes})

			if len(app.routes) != tt.wantRoutes {
				t.Errorf("routes len = %v, want %v", len(app.routes), tt.wantRoutes)
			}
		})
	}
}

func TestUse(t *testing.T) {
	tests := []struct {
		name       string
		routes     []Route
		wantRoutes int
	}{
		{
			name: "module registers its controller",
			routes: []Route{
				GET("/products", dummyHandler),
				POST("/products", dummyHandler),
			},
			wantRoutes: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New(KConfig{})
			app.Use(&testModule{
				controller: &testController{routes: tt.routes},
			})

			if len(app.routes) != tt.wantRoutes {
				t.Errorf("routes len = %v, want %v", len(app.routes), tt.wantRoutes)
			}
		})
	}
}

func TestDocsRoutes(t *testing.T) {
	tests := []struct {
		name         string
		cfg          KConfig
		wantDocsCode int
		wantSpecCode int
	}{
		{
			name:         "docs enabled in development",
			cfg:          KConfig{Env: "development"},
			wantDocsCode: http.StatusOK,
			wantSpecCode: http.StatusOK,
		},
		{
			name:         "docs disabled in production",
			cfg:          KConfig{Env: "production"},
			wantDocsCode: http.StatusNotFound,
			wantSpecCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New(tt.cfg)

			// Registrar rutas de docs manualmente como lo hace Listen()
			if app.config.docsEnabled() {
				app.fiber.Get("/docs/openapi.json", func(c *fiber.Ctx) error {
					return c.JSON(map[string]string{})
				})
				app.fiber.Get(app.config.Docs.Path, func(c *fiber.Ctx) error {
					return c.SendString("swagger ui")
				})
			}

			req := httptest.NewRequest("GET", "/docs/openapi.json", nil)
			resp, err := app.Fiber().Test(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantSpecCode {
				t.Errorf("openapi.json StatusCode = %v, want %v", resp.StatusCode, tt.wantSpecCode)
			}
		})
	}
}
