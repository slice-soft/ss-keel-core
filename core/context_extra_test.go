package core

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// — SetUser / UserAs —

func TestSetUserAndUser(t *testing.T) {
	type User struct{ ID string }

	app := newTestApp("GET", "/test", func(c *Ctx) error {
		c.SetUser(User{ID: "42"})
		u := c.User()
		if u == nil {
			return fmt.Errorf("User() returned nil")
		}
		return c.OK(nil)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %v, want 200", resp.StatusCode)
	}
}

func TestUserAs(t *testing.T) {
	type AuthUser struct{ ID string; Role string }

	t.Run("correct type returns value and true", func(t *testing.T) {
		app := newTestApp("GET", "/test", func(c *Ctx) error {
			c.SetUser(AuthUser{ID: "1", Role: "admin"})
			u, ok := UserAs[AuthUser](c)
			if !ok {
				return fmt.Errorf("UserAs returned false")
			}
			if u.ID != "1" || u.Role != "admin" {
				return fmt.Errorf("unexpected user: %+v", u)
			}
			return c.OK(nil)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want 200", resp.StatusCode)
		}
	})

	t.Run("wrong type returns zero and false", func(t *testing.T) {
		type OtherUser struct{ Email string }
		app := newTestApp("GET", "/test", func(c *Ctx) error {
			c.SetUser(AuthUser{ID: "1", Role: "admin"})
			_, ok := UserAs[OtherUser](c)
			if ok {
				return fmt.Errorf("UserAs should return false for wrong type")
			}
			return c.OK(nil)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want 200", resp.StatusCode)
		}
	})

	t.Run("no user set returns false", func(t *testing.T) {
		app := newTestApp("GET", "/test", func(c *Ctx) error {
			_, ok := UserAs[AuthUser](c)
			if ok {
				return fmt.Errorf("UserAs should return false when no user is set")
			}
			return c.OK(nil)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want 200", resp.StatusCode)
		}
	})
}

// — ParsePagination —

func TestParsePagination(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantPage  int
		wantLimit int
	}{
		{
			name:      "defaults when no query",
			query:     "",
			wantPage:  1,
			wantLimit: 20,
		},
		{
			name:      "custom page and limit",
			query:     "?page=3&limit=50",
			wantPage:  3,
			wantLimit: 50,
		},
		{
			name:      "limit clamped to 100",
			query:     "?page=1&limit=999",
			wantPage:  1,
			wantLimit: 100,
		},
		{
			name:      "page below 1 defaults to 1",
			query:     "?page=0&limit=10",
			wantPage:  1,
			wantLimit: 10,
		},
		{
			name:      "negative limit defaults to 20",
			query:     "?page=1&limit=-5",
			wantPage:  1,
			wantLimit: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPage, gotLimit int
			app := newTestApp("GET", "/test", func(c *Ctx) error {
				q := c.ParsePagination()
				gotPage = q.Page
				gotLimit = q.Limit
				return c.OK(nil)
			})

			req := httptest.NewRequest("GET", "/test"+tt.query, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Errorf("StatusCode = %v, want 200", resp.StatusCode)
			}
			if gotPage != tt.wantPage {
				t.Errorf("Page = %v, want %v", gotPage, tt.wantPage)
			}
			if gotLimit != tt.wantLimit {
				t.Errorf("Limit = %v, want %v", gotLimit, tt.wantLimit)
			}
		})
	}
}

// — Lang / T (i18n) —

func TestLang(t *testing.T) {
	tests := []struct {
		name           string
		acceptLanguage string
		wantLang       string
	}{
		{name: "no header returns en", acceptLanguage: "", wantLang: "en"},
		{name: "simple lang", acceptLanguage: "es", wantLang: "es"},
		{name: "lang with region", acceptLanguage: "en-US", wantLang: "en-US"},
		{name: "multiple langs picks first", acceptLanguage: "fr,en;q=0.9", wantLang: "fr"},
		{name: "lang with quality picks before semicolon", acceptLanguage: "en-US,en;q=0.9", wantLang: "en-US"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotLang string
			app := newTestApp("GET", "/test", func(c *Ctx) error {
				gotLang = c.Lang()
				return c.OK(nil)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.acceptLanguage != "" {
				req.Header.Set("Accept-Language", tt.acceptLanguage)
			}
			app.Test(req) //nolint
			if gotLang != tt.wantLang {
				t.Errorf("Lang() = %v, want %v", gotLang, tt.wantLang)
			}
		})
	}
}

func TestCtxT(t *testing.T) {
	t.Run("returns key as-is when no translator registered", func(t *testing.T) {
		var result string
		app := newTestApp("GET", "/test", func(c *Ctx) error {
			result = c.T("greeting.hello")
			return c.OK(nil)
		})
		req := httptest.NewRequest("GET", "/test", nil)
		app.Test(req) //nolint
		if result != "greeting.hello" {
			t.Errorf("T() = %v, want greeting.hello", result)
		}
	})

	t.Run("uses translator when registered", func(t *testing.T) {
		var result string
		keelApp := New(KConfig{DisableHealth: true})
		keelApp.SetTranslator(&mockTranslator{})
		keelApp.RegisterController(ControllerFunc(func() []Route {
			return []Route{
				GET("/test", func(c *Ctx) error {
					result = c.T("hello")
					return c.OK(nil)
				}),
			}
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Language", "es")
		keelApp.Fiber().Test(req) //nolint
		if result != "hola" {
			t.Errorf("T() = %v, want hola", result)
		}
	})
}

// mockTranslator is a simple test double that supports en/es.
type mockTranslator struct{}

func (m *mockTranslator) T(locale, key string, _ ...any) string {
	translations := map[string]map[string]string{
		"en": {"hello": "hello"},
		"es": {"hello": "hola"},
	}
	if loc, ok := translations[locale]; ok {
		if val, ok := loc[key]; ok {
			return val
		}
	}
	return key
}

func (m *mockTranslator) Locales() []string { return []string{"en", "es"} }

// — Metrics collector —

func TestMetricsCollector(t *testing.T) {
	t.Run("RecordRequest called after each request", func(t *testing.T) {
		mc := &mockMetricsCollector{}
		keelApp := New(KConfig{DisableHealth: true})
		keelApp.SetMetricsCollector(mc)
		keelApp.RegisterController(ControllerFunc(func() []Route {
			return []Route{
				GET("/ping", func(c *Ctx) error { return c.OK(nil) }),
			}
		}))

		req := httptest.NewRequest("GET", "/ping", nil)
		keelApp.Fiber().Test(req) //nolint

		if mc.calls != 1 {
			t.Errorf("RecordRequest calls = %v, want 1", mc.calls)
		}
		if mc.lastMetrics.Method != "GET" {
			t.Errorf("Method = %v, want GET", mc.lastMetrics.Method)
		}
		if mc.lastMetrics.Path != "/ping" {
			t.Errorf("Path = %v, want /ping", mc.lastMetrics.Path)
		}
		if mc.lastMetrics.StatusCode != 200 {
			t.Errorf("StatusCode = %v, want 200", mc.lastMetrics.StatusCode)
		}
	})
}

type mockMetricsCollector struct {
	calls       int
	lastMetrics RequestMetrics
}

func (m *mockMetricsCollector) RecordRequest(metrics RequestMetrics) {
	m.calls++
	m.lastMetrics = metrics
}

// — Tracer —

func TestTracer(t *testing.T) {
	t.Run("default tracer is noop and never nil", func(t *testing.T) {
		keelApp := New(KConfig{DisableHealth: true})
		if keelApp.Tracer() == nil {
			t.Error("Tracer() should never return nil")
		}
		// noop tracer must not panic
		ctx, span := keelApp.Tracer().Start(t.Context(), "test-op")
		if ctx == nil {
			t.Error("Start() context should not be nil")
		}
		span.SetAttribute("key", "value")
		span.RecordError(nil)
		span.End()
	})

	t.Run("custom tracer replaces noop", func(t *testing.T) {
		keelApp := New(KConfig{DisableHealth: true})
		custom := &mockTracer{}
		keelApp.SetTracer(custom)
		// Confirm the custom tracer is actually used by starting a span.
		_, _ = keelApp.Tracer().Start(context.Background(), "op")
		if custom.started != 1 {
			t.Errorf("custom tracer Start() calls = %v, want 1", custom.started)
		}
	})
}

type mockTracer struct{ started int }

func (m *mockTracer) Start(ctx context.Context, _ string) (context.Context, Span) {
	m.started++
	return ctx, noopSpan{}
}

// — OnShutdown —

func TestOnShutdown(t *testing.T) {
	t.Run("hooks are appended", func(t *testing.T) {
		keelApp := New(KConfig{DisableHealth: true})
		keelApp.OnShutdown(func(ctx context.Context) error { return nil })
		keelApp.OnShutdown(func(ctx context.Context) error { return nil })
		if len(keelApp.shutdownHooks) != 2 {
			t.Errorf("shutdownHooks len = %v, want 2", len(keelApp.shutdownHooks))
		}
	})
}
