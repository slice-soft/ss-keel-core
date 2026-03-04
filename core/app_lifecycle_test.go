package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterDocsRoutes(t *testing.T) {
	t.Run("registers docs routes when enabled", func(t *testing.T) {
		app := New(KConfig{
			DisableHealth: true,
			Env:           "development",
			Docs: DocsConfig{
				Path:    "/docs",
				Title:   "Docs",
				Version: "1.0.0",
			},
		})

		app.registerDocsRoutes()

		resp, err := app.Fiber().Test(httptest.NewRequest("GET", "/docs/openapi.json", nil))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("openapi status = %d, want %d", resp.StatusCode, http.StatusOK)
		}

		resp, err = app.Fiber().Test(httptest.NewRequest("GET", "/docs", nil))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("docs status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("does not register docs routes in production", func(t *testing.T) {
		app := New(KConfig{
			DisableHealth: true,
			Env:           "production",
		})

		app.registerDocsRoutes()

		resp, err := app.Fiber().Test(httptest.NewRequest("GET", "/docs/openapi.json", nil))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("openapi status = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}
	})
}

func TestShutdownRunsHooks(t *testing.T) {
	app := New(KConfig{DisableHealth: true})
	called := 0

	app.OnShutdown(func(context.Context) error {
		called++
		return nil
	})
	app.OnShutdown(func(context.Context) error {
		called++
		return nil
	})

	// App is not listening in this test; shutdown may return an error depending
	// on Fiber internals, but hooks must run regardless.
	_ = app.shutdown()

	if called != 2 {
		t.Fatalf("shutdown hooks called = %d, want 2", called)
	}
}

func TestListenReturnsErrorOnInvalidPort(t *testing.T) {
	app := New(KConfig{
		DisableHealth: true,
		Port:          -1,
		Env:           "production",
	})
	s := &schedulerSpy{}
	app.RegisterScheduler(s)

	err := app.Listen()
	if err == nil {
		t.Fatal("Listen() should return error for invalid port")
	}
	if !s.started {
		t.Fatal("scheduler Start() should be called before listen failure")
	}
}
