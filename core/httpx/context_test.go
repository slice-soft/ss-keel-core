package httpx

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func newHTTPXTestApp(method, path string, handler func(*Ctx) error) *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Add(method, path, WrapHandler(handler))
	return app
}

func TestWrapHandler(t *testing.T) {
	called := false
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/ping", WrapHandler(func(c *Ctx) error {
		called = true
		return c.NoContent()
	}))

	resp, err := app.Test(httptest.NewRequest("GET", "/ping", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
	if !called {
		t.Fatal("wrapped handler was not executed")
	}
}

func TestParseBody(t *testing.T) {
	type dto struct {
		Name string `json:"name" validate:"required"`
	}

	tests := []struct {
		name     string
		body     []byte
		wantCode int
	}{
		{
			name:     "valid body",
			body:     []byte(`{"name":"juan"}`),
			wantCode: http.StatusOK,
		},
		{
			name:     "invalid json",
			body:     []byte(`{"name":`),
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "validation error",
			body:     []byte(`{}`),
			wantCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newHTTPXTestApp("POST", "/body", func(c *Ctx) error {
				var in dto
				if err := c.ParseBody(&in); err != nil {
					return err
				}
				return c.OK(in)
			})

			req := httptest.NewRequest("POST", "/body", bytes.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantCode {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantCode)
			}
		})
	}
}

func TestUserAndUserAs(t *testing.T) {
	type authUser struct {
		ID string
	}

	var errMsg string
	app := newHTTPXTestApp("GET", "/me", func(c *Ctx) error {
		c.SetUser(authUser{ID: "u-1"})
		v := c.User()
		if v == nil {
			errMsg = "User() returned nil"
			return c.Status(http.StatusInternalServerError).SendString(errMsg)
		}
		u, ok := UserAs[authUser](c)
		if !ok || u.ID != "u-1" {
			errMsg = "UserAs() did not return expected value"
			return c.Status(http.StatusInternalServerError).SendString(errMsg)
		}
		_, ok = UserAs[struct{ Email string }](c)
		if ok {
			errMsg = "UserAs() should fail for wrong type"
			return c.Status(http.StatusInternalServerError).SendString(errMsg)
		}
		return c.NoContent()
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/me", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
	if errMsg != "" {
		t.Fatal(errMsg)
	}
}

func TestLang(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{name: "default", header: "", want: "en"},
		{name: "simple", header: "es", want: "es"},
		{name: "with region", header: "en-US", want: "en-US"},
		{name: "comma separated", header: "fr,en;q=0.9", want: "fr"},
		{name: "semicolon", header: "pt-BR;q=0.8", want: "pt-BR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			app := newHTTPXTestApp("GET", "/lang", func(c *Ctx) error {
				got = c.Lang()
				return c.NoContent()
			})

			req := httptest.NewRequest("GET", "/lang", nil)
			if tt.header != "" {
				req.Header.Set("Accept-Language", tt.header)
			}
			_, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("Lang() = %q, want %q", got, tt.want)
			}
		})
	}
}

type testTranslator struct{}

func (testTranslator) T(locale, key string, _ ...any) string {
	if locale == "es" && key == "hello" {
		return "hola"
	}
	return key
}

func (testTranslator) Locales() []string { return []string{"en", "es"} }

func TestT(t *testing.T) {
	t.Run("without translator returns key", func(t *testing.T) {
		var got string
		app := newHTTPXTestApp("GET", "/t", func(c *Ctx) error {
			got = c.T("hello")
			return c.NoContent()
		})

		_, err := app.Test(httptest.NewRequest("GET", "/t", nil))
		if err != nil {
			t.Fatal(err)
		}
		if got != "hello" {
			t.Fatalf("T() = %q, want %q", got, "hello")
		}
	})

	t.Run("with translator from locals", func(t *testing.T) {
		var got string
		app := fiber.New(fiber.Config{DisableStartupMessage: true})
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("_keel_translator", testTranslator{})
			return c.Next()
		})
		app.Get("/t", WrapHandler(func(c *Ctx) error {
			got = c.T("hello")
			return c.NoContent()
		}))

		req := httptest.NewRequest("GET", "/t", nil)
		req.Header.Set("Accept-Language", "es")
		_, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		if got != "hola" {
			t.Fatalf("T() = %q, want %q", got, "hola")
		}
	})
}

func TestResponseHelpers(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		handler  func(*Ctx) error
		wantCode int
	}{
		{
			name:   "ok",
			method: "GET",
			path:   "/ok",
			handler: func(c *Ctx) error {
				return c.OK(map[string]string{"status": "ok"})
			},
			wantCode: http.StatusOK,
		},
		{
			name:   "created",
			method: "POST",
			path:   "/created",
			handler: func(c *Ctx) error {
				return c.Created(map[string]string{"id": "1"})
			},
			wantCode: http.StatusCreated,
		},
		{
			name:   "no content",
			method: "DELETE",
			path:   "/no-content",
			handler: func(c *Ctx) error {
				return c.NoContent()
			},
			wantCode: http.StatusNoContent,
		},
		{
			name:   "not found",
			method: "GET",
			path:   "/nf",
			handler: func(c *Ctx) error {
				return c.NotFound("missing")
			},
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newHTTPXTestApp(tt.method, tt.path, tt.handler)
			resp, err := app.Test(httptest.NewRequest(tt.method, tt.path, nil))
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantCode {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantCode)
			}
			if tt.wantCode == http.StatusNotFound {
				var body map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if body["message"] != "missing" {
					t.Fatalf("message = %v, want missing", body["message"])
				}
			}
		})
	}
}
