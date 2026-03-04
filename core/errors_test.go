package core

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
)

func TestKErrorConstructors(t *testing.T) {
	tests := []struct {
		name           string
		err            *KError
		wantCode       string
		wantStatusCode int
		wantMessage    string
	}{
		{
			name:           "NotFound",
			err:            NotFound("user not found"),
			wantCode:       "NOT_FOUND",
			wantStatusCode: 404,
			wantMessage:    "user not found",
		},
		{
			name:           "Unauthorized",
			err:            Unauthorized("invalid token"),
			wantCode:       "UNAUTHORIZED",
			wantStatusCode: 401,
			wantMessage:    "invalid token",
		},
		{
			name:           "Forbidden",
			err:            Forbidden("access denied"),
			wantCode:       "FORBIDDEN",
			wantStatusCode: 403,
			wantMessage:    "access denied",
		},
		{
			name:           "Conflict",
			err:            Conflict("email already exists"),
			wantCode:       "CONFLICT",
			wantStatusCode: 409,
			wantMessage:    "email already exists",
		},
		{
			name:           "BadRequest",
			err:            BadRequest("invalid input"),
			wantCode:       "BAD_REQUEST",
			wantStatusCode: 400,
			wantMessage:    "invalid input",
		},
		{
			name:           "Internal without cause",
			err:            Internal("something broke", nil),
			wantCode:       "INTERNAL_ERROR",
			wantStatusCode: 500,
			wantMessage:    "something broke",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.wantCode {
				t.Errorf("Code = %v, want %v", tt.err.Code, tt.wantCode)
			}
			if tt.err.StatusCode != tt.wantStatusCode {
				t.Errorf("StatusCode = %v, want %v", tt.err.StatusCode, tt.wantStatusCode)
			}
			if tt.err.Message != tt.wantMessage {
				t.Errorf("Message = %v, want %v", tt.err.Message, tt.wantMessage)
			}
		})
	}
}

func TestKErrorError(t *testing.T) {
	t.Run("without cause", func(t *testing.T) {
		err := NotFound("item not found")
		if err.Error() != "item not found" {
			t.Errorf("Error() = %v, want %v", err.Error(), "item not found")
		}
	})

	t.Run("with cause includes cause in message", func(t *testing.T) {
		cause := errors.New("connection refused")
		err := Internal("db error", cause)
		if err.Error() != "db error: connection refused" {
			t.Errorf("Error() = %v", err.Error())
		}
	})
}

func TestKErrorUnwrap(t *testing.T) {
	t.Run("Unwrap returns nil when no cause", func(t *testing.T) {
		err := NotFound("not found")
		if err.Unwrap() != nil {
			t.Error("Unwrap() should return nil when no cause")
		}
	})

	t.Run("Unwrap returns the original cause", func(t *testing.T) {
		cause := errors.New("original error")
		err := Internal("wrapped", cause)
		if !errors.Is(err, cause) {
			t.Error("errors.Is() should unwrap to the original cause")
		}
	})

	t.Run("errors.As works on wrapped KError", func(t *testing.T) {
		cause := NotFound("inner not found")
		outer := Internal("outer", cause)
		var ke *KError
		if !errors.As(outer, &ke) {
			t.Error("errors.As() should find *KError in the chain")
		}
	})
}

func TestKErrorHTTPHandler(t *testing.T) {
	tests := []struct {
		name       string
		kerr       *KError
		wantStatus int
		wantCode   string
	}{
		{"NotFound maps to 404", NotFound("user not found"), 404, "NOT_FOUND"},
		{"Unauthorized maps to 401", Unauthorized("no token"), 401, "UNAUTHORIZED"},
		{"Forbidden maps to 403", Forbidden("no access"), 403, "FORBIDDEN"},
		{"Conflict maps to 409", Conflict("duplicate"), 409, "CONFLICT"},
		{"BadRequest maps to 400", BadRequest("bad input"), 400, "BAD_REQUEST"},
		{"Internal maps to 500", Internal("boom", nil), 500, "INTERNAL_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New(KConfig{DisableHealth: true})
			ke := tt.kerr
			app.RegisterController(ControllerFunc(func() []Route {
				return []Route{
					GET("/test", func(c *Ctx) error {
						return ke
					}),
				}
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Fiber().Test(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("StatusCode = %v, want %v", resp.StatusCode, tt.wantStatus)
			}

			var body map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["code"] != tt.wantCode {
				t.Errorf("code = %v, want %v", body["code"], tt.wantCode)
			}
		})
	}
}
