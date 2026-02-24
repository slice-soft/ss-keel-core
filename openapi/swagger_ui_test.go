package openapi

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestSwaggerUIHandler(t *testing.T) {
	app := fiber.New()
	app.Get("/docs", SwaggerUIHandler("/docs/openapi.json"))

	req := httptest.NewRequest("GET", "/docs", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %v, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %v, want text/html", ct)
	}

	body := make([]byte, resp.ContentLength)
	resp.Body.Read(body)
	if !strings.Contains(string(body), "/docs/openapi.json") {
		t.Error("HTML should contain the specPath")
	}
}
