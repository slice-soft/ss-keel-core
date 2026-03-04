package core

import (
	"io"
	"net/http"
)

// TestApp wraps App for use in unit tests.
// It uses Fiber's built-in test helper so no port binding is needed.
type TestApp struct {
	*App
}

// NewTestApp creates a minimal App suitable for controller testing.
func NewTestApp() *TestApp {
	cfg := applyDefaults(KConfig{DisableHealth: true})
	return &TestApp{App: New(cfg)}
}

// Request performs an HTTP request against the app without starting a real server.
// headers is an optional map of header key-value pairs.
func (t *TestApp) Request(method, path string, body io.Reader, headers ...map[string]string) *http.Response {
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		panic(err)
	}
	for _, h := range headers {
		for k, v := range h {
			req.Header.Set(k, v)
		}
	}
	resp, err := t.App.fiber.Test(req, -1)
	if err != nil {
		panic(err)
	}
	return resp
}

// RequestJSON performs a request with Content-Type: application/json.
func (t *TestApp) RequestJSON(method, path string, body io.Reader) *http.Response {
	return t.Request(method, path, body, map[string]string{
		"Content-Type": "application/json",
	})
}
