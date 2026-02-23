package core

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// newTestApp creates a minimal Fiber app for testing Ctx methods.
func newTestApp(method, path string, handler func(*Ctx) error) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	app.Add(method, path, WrapHandler(handler))
	return app
}

func TestOK(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		wantCode int
	}{
		{
			name:     "responds with 200",
			data:     map[string]string{"message": "ok"},
			wantCode: http.StatusOK,
		},
		{
			name:     "responds with 200 and struct",
			data:     struct{ ID string }{ID: "123"},
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApp("GET", "/test", func(ctx *Ctx) error {
				return ctx.OK(tt.data)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantCode {
				t.Errorf("StatusCode = %v, want %v", resp.StatusCode, tt.wantCode)
			}
		})
	}
}

func TestCreated(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		wantCode int
	}{
		{
			name:     "responds with 201",
			data:     map[string]string{"id": "123"},
			wantCode: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApp("POST", "/test", func(ctx *Ctx) error {
				return ctx.Created(tt.data)
			})

			req := httptest.NewRequest("POST", "/test", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantCode {
				t.Errorf("StatusCode = %v, want %v", resp.StatusCode, tt.wantCode)
			}
		})
	}
}

func TestNoContent(t *testing.T) {
	tests := []struct {
		name     string
		wantCode int
	}{
		{
			name:     "responds with 204",
			wantCode: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApp("DELETE", "/test", func(ctx *Ctx) error {
				return ctx.NoContent()
			})

			req := httptest.NewRequest("DELETE", "/test", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantCode {
				t.Errorf("StatusCode = %v, want %v", resp.StatusCode, tt.wantCode)
			}
		})
	}
}

func TestNotFound(t *testing.T) {
	tests := []struct {
		name        string
		message     []string
		wantCode    int
		wantMessage string
	}{
		{
			name:        "default message",
			message:     []string{},
			wantCode:    http.StatusNotFound,
			wantMessage: "resource not found",
		},
		{
			name:        "custom message",
			message:     []string{"user not found"},
			wantCode:    http.StatusNotFound,
			wantMessage: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApp("GET", "/test", func(ctx *Ctx) error {
				return ctx.NotFound(tt.message...)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req)
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
			if body["message"] != tt.wantMessage {
				t.Errorf("message = %v, want %v", body["message"], tt.wantMessage)
			}
		})
	}
}

func TestParseBody(t *testing.T) {
	type testDTO struct {
		Name  string `json:"name"  validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	tests := []struct {
		name     string
		body     any
		wantCode int
	}{
		{
			name:     "valid body",
			body:     map[string]string{"name": "Juan", "email": "juan@test.com"},
			wantCode: http.StatusOK,
		},
		{
			name:     "invalid JSON",
			body:     "not json",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "missing required fields",
			body:     map[string]string{},
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name:     "invalid email",
			body:     map[string]string{"name": "Juan", "email": "notanemail"},
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name:     "missing name only",
			body:     map[string]string{"email": "juan@test.com"},
			wantCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApp("POST", "/test", func(ctx *Ctx) error {
				var dto testDTO
				if err := ctx.ParseBody(&dto); err != nil {
					return err
				}
				return ctx.OK(dto)
			})

			var bodyBytes []byte
			var err error
			if s, ok := tt.body.(string); ok {
				bodyBytes = []byte(s)
			} else {
				bodyBytes, err = json.Marshal(tt.body)
				if err != nil {
					t.Fatal(err)
				}
			}

			req := httptest.NewRequest("POST", "/test", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantCode {
				t.Errorf("StatusCode = %v, want %v", resp.StatusCode, tt.wantCode)
			}
		})
	}
}
