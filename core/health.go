package core

import (
	"sync"

	"github.com/slice-soft/ss-keel-core/contracts"
	"github.com/slice-soft/ss-keel-core/core/httpx"
)

// RegisterHealthChecker adds a health checker to the app.
func (a *App) RegisterHealthChecker(h contracts.HealthChecker) {
	a.healthCheckers = append(a.healthCheckers, h)
}

// healthResponse is the response for the /health endpoint.
type healthResponse struct {
	Status  string            `json:"status"   doc:"Overall service status"  example:"UP"`
	Service string            `json:"service"  doc:"Service name"            example:"My API"`
	Version string            `json:"version"  doc:"Service version"         example:"1.0.0"`
	Checks  map[string]string `json:"checks,omitempty" doc:"Per-dependency check results"`
}

// registerHealth adds the /health route to both Fiber and the OpenAPI spec.
// It is called automatically in New() unless DisableHealth is set to true.
func (a *App) registerHealth() {
	a.RegisterController(contracts.ControllerFunc[httpx.Route](func() []httpx.Route {
		return []httpx.Route{
			httpx.GET("/health", httpx.WrapHandler(func(c *httpx.Ctx) error {
				status := "UP"
				checks := make(map[string]string)

				if len(a.healthCheckers) > 0 {
					var mu sync.Mutex
					var wg sync.WaitGroup
					ctx := c.Context()

					for _, hc := range a.healthCheckers {
						hc := hc
						wg.Add(1)
						go func() {
							defer wg.Done()
							result := "UP"
							if err := hc.Check(ctx); err != nil {
								result = "DOWN: " + err.Error()
								mu.Lock()
								status = "DOWN"
								mu.Unlock()
							}
							mu.Lock()
							checks[hc.Name()] = result
							mu.Unlock()
						}()
					}
					wg.Wait()
				}

				resp := healthResponse{
					Status:  status,
					Service: a.config.ServiceName,
					Version: a.config.Docs.Version,
				}
				if len(checks) > 0 {
					resp.Checks = checks
				}

				if status == "DOWN" {
					return c.Status(503).JSON(resp)
				}
				return c.OK(resp)
			})).
				WithResponse(httpx.WithResponse[healthResponse](200)).
				Tag("system").
				Describe("Health check", "Returns the current status of the service"),
		}
	}))
}
