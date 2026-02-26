package main

import (
	"log"

	"github.com/slice-soft/ss-keel-core/config"
	"github.com/slice-soft/ss-keel-core/core"
)

type HelloResponse struct {
	Message string `json:"message" doc:"Greeting message" example:"Hello, Keel!"`
}

func main() {
	app := core.New(core.KConfig{
		Port:        config.GetEnvIntOrDefault("PORT", 3000),
		ServiceName: config.GetEnvOrDefault("SERVICE_NAME", "My API"),
		Env:         config.GetEnvOrDefault("APP_ENV", "development"),
		Docs: core.DocsConfig{
			Title:   config.GetEnvOrDefault("DOCS_TITLE", "My API"),
			Version: config.GetEnvOrDefault("DOCS_VERSION", "1.0.0"),
		},
	})

	app.RegisterController(core.ControllerFunc(func() []core.Route {
		return []core.Route{
			core.GET("/hello", func(c *core.Ctx) error {
				return c.OK(HelloResponse{Message: "Hello, Keel!"})
			}).
				WithResponse(core.WithResponse[HelloResponse](200)).
				Tag("greeting").
				Describe("Returns a greeting message"),
		}
	}))

	log.Fatal(app.Listen())
}
