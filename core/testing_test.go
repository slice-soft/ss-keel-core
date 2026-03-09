package core

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/slice-soft/ss-keel-core/contracts"
	"github.com/slice-soft/ss-keel-core/core/httpx"
)

func TestNewTestApp(t *testing.T) {
	app := NewTestApp()
	if app == nil || app.App == nil {
		t.Fatal("NewTestApp() returned nil app")
	}

	// Health is disabled in NewTestApp.
	resp := app.Request("GET", "/health", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestTestAppRequestHelpers(t *testing.T) {
	type bodyDTO struct {
		Name string `json:"name" validate:"required"`
	}

	app := NewTestApp()
	app.RegisterController(contracts.ControllerFunc[httpx.Route](func() []httpx.Route {
		return []httpx.Route{
			httpx.GET("/headers", httpx.WrapHandler(func(c *httpx.Ctx) error {
				if c.Get("X-Test") != "1" {
					return c.Status(http.StatusBadRequest).JSON(map[string]string{"error": "missing header"})
				}
				return c.OK(map[string]string{"status": "ok"})
			})),
			httpx.POST("/echo", httpx.WrapHandler(func(c *httpx.Ctx) error {
				var in bodyDTO
				if err := c.ParseBody(&in); err != nil {
					return err
				}
				return c.OK(in)
			})),
		}
	}))

	resp := app.Request("GET", "/headers", nil, map[string]string{"X-Test": "1"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Request() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	resp = app.RequestJSON("POST", "/echo", strings.NewReader(`{"name":"ana"}`))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("RequestJSON() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var out bodyDTO
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Name != "ana" {
		t.Fatalf("decoded body = %+v, want name=ana", out)
	}
}
