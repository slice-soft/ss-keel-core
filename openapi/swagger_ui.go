package openapi

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// SwaggerUIHandler returns a Fiber handler that serves embedded Swagger UI.
// specPath is the URL where the openapi.json is located.
func SwaggerUIHandler(specPath string) fiber.Handler {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <title>Keel — API Docs</title>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>
    body { margin: 0; }
    .topbar { display: none; } /* Hide Swagger's topbar with its logo */
  </style>
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>
  SwaggerUIBundle({
    url: "%s",
    dom_id: '#swagger-ui',
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIBundle.SwaggerUIStandalonePreset
    ],
    layout: "BaseLayout",
    deepLinking: true,
    displayRequestDuration: true,
    filter: true,
    persistAuthorization: true,
    tryItOutEnabled: true,
    docExpansion: "list",
    defaultModelsExpandDepth: 3,
  })
</script>
</body>
</html>`, specPath)

	return func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	}
}
