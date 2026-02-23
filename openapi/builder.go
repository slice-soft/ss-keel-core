package openapi

import (
	"fmt"
	"reflect"
	"strings"
)

// Spec is the in-memory representation of an OpenAPI 3.0 spec.
type Spec struct {
	OpenAPI    string                `json:"openapi"`
	Info       Info                  `json:"info"`
	Paths      map[string]any        `json:"paths"`
	Components Components            `json:"components"`
	Security   []map[string][]string `json:"security,omitempty"`
}

type Info struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

// Components groups global security schemes.
type Components struct {
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

// SecurityScheme defines an authentication scheme in OpenAPI.
type SecurityScheme struct {
	Type         string `json:"type"`
	Scheme       string `json:"scheme,omitempty"`       // para http
	In           string `json:"in,omitempty"`           // para apiKey
	Name         string `json:"name,omitempty"`         // para apiKey
	BearerFormat string `json:"bearerFormat,omitempty"` // para http bearer
}

// RouteInput is the neutral representation of a route.
type RouteInput struct {
	Method      string
	Path        string
	Summary     string
	Description string
	Tags        []string
	Secured     []string // required security schemes
	Body        any
	Response    any
	StatusCode  int
}

// BuildInput groups the data to build the spec.
type BuildInput struct {
	Title   string
	Version string
	Routes  []RouteInput
}

// Build constructs the OpenAPI 3.0 spec in memory.
func Build(input BuildInput) Spec {
	paths := make(map[string]any)
	securitySchemes := make(map[string]SecurityScheme)

	for _, route := range input.Routes {
		oaPath := fiberPathToOA(route.Path)

		if paths[oaPath] == nil {
			paths[oaPath] = make(map[string]any)
		}

		operation := map[string]any{
			"summary":     route.Summary,
			"description": route.Description,
			"tags":        route.Tags,
			"responses":   buildResponses(route),
		}

		if route.Body != nil {
			operation["requestBody"] = buildRequestBody(route.Body)
		}

		// Security in the operation — one requirement per scheme
		if len(route.Secured) > 0 {
			var security []map[string][]string
			for _, scheme := range route.Secured {
				security = append(security, map[string][]string{scheme: {}})
				// Register the scheme in components if it doesn't exist
				if _, exists := securitySchemes[scheme]; !exists {
					securitySchemes[scheme] = inferSecurityScheme(scheme)
				}
			}
			operation["security"] = security
		}

		method := strings.ToLower(route.Method)
		paths[oaPath].(map[string]any)[method] = operation
	}

	return Spec{
		OpenAPI: "3.0.0",
		Info:    Info{Title: input.Title, Version: input.Version},
		Paths:   paths,
		Components: Components{
			SecuritySchemes: securitySchemes,
		},
	}
}

// inferSecurityScheme infers the scheme type by name convention.
// bearerAuth → HTTP Bearer JWT
// apiKey     → API Key in header
// basicAuth  → HTTP Basic
func inferSecurityScheme(name string) SecurityScheme {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "bearer"):
		return SecurityScheme{
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		}
	case strings.Contains(lower, "basic"):
		return SecurityScheme{
			Type:   "http",
			Scheme: "basic",
		}
	default:
		// Fallback: API Key en header X-API-Key
		return SecurityScheme{
			Type: "apiKey",
			In:   "header",
			Name: "X-API-Key",
		}
	}
}

// fiberPathToOA converts /users/:id → /users/{id}
func fiberPathToOA(p string) string {
	parts := strings.Split(p, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "{" + part[1:] + "}"
		}
	}
	return strings.Join(parts, "/")
}

func buildRequestBody(dto any) map[string]any {
	return map[string]any{
		"required": true,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": reflectSchema(dto),
			},
		},
	}
}

func buildResponses(route RouteInput) map[string]any {
	code := route.StatusCode
	if code == 0 {
		code = 200
	}
	responses := map[string]any{}
	if route.Response != nil {
		responses[fmt.Sprintf("%d", code)] = map[string]any{
			"description": "Success",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": reflectSchema(route.Response),
				},
			},
		}
	}
	return responses
}

// reflectSchema generates an OpenAPI schema from a struct.
// Reads tags: json, validate, doc, example.
func reflectSchema(v any) map[string]any {
	t := reflect.TypeOf(v)
	if t == nil {
		return map[string]any{"type": "object"}
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return map[string]any{"type": "object"}
	}

	properties := map[string]any{}
	var required []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		name := strings.Split(jsonTag, ",")[0]

		prop := map[string]any{
			"type": goTypeToOA(field.Type.Kind()),
		}
		if doc := field.Tag.Get("doc"); doc != "" {
			prop["description"] = doc
		}
		if example := field.Tag.Get("example"); example != "" {
			prop["example"] = example
		}

		properties[name] = prop

		if strings.Contains(field.Tag.Get("validate"), "required") {
			required = append(required, name)
		}
	}

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func goTypeToOA(k reflect.Kind) string {
	switch k {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	default:
		return "string"
	}
}
