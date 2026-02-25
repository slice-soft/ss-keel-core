package core

// healthResponse is the default response for the /health endpoint.
type healthResponse struct {
	Status  string `json:"status"  doc:"Service status"  example:"UP"`
	Service string `json:"service" doc:"Service name"    example:"My API"`
	Version string `json:"version" doc:"Service version" example:"1.0.0"`
}

// registerHealth adds the /health route to both Fiber and the OpenAPI spec.
// It is called automatically in Listen() unless DisableHealth is set to true.
func (a *App) registerHealth() {
	a.RegisterController(ControllerFunc(func() []Route {
		return []Route{
			GET("/health", func(c *Ctx) error {
				return c.OK(healthResponse{
					Status:  "UP",
					Service: a.config.ServiceName,
					Version: a.config.Docs.Version,
				})
			}).
				WithResponse(WithResponse[healthResponse](200)).
				Tag("system").
				Describe("Health check", "Returns the current status of the service"),
		}
	}))
}
