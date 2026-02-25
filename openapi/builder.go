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

// Components groups reusable schemas and security schemes.
type Components struct {
	Schemas         map[string]any            `json:"schemas,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

// SecurityScheme defines an authentication scheme in OpenAPI.
type SecurityScheme struct {
	Type         string `json:"type"`
	Scheme       string `json:"scheme,omitempty"`
	In           string `json:"in,omitempty"`
	Name         string `json:"name,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
}

// RouteInput is the neutral representation of a route.
type RouteInput struct {
	Method      string
	Path        string
	Summary     string
	Description string
	Tags        []string
	Secured     []string
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
	schemas := make(map[string]any)
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
			"responses":   buildResponses(route, schemas),
		}

		if route.Body != nil {
			operation["requestBody"] = buildRequestBody(route.Body, schemas)
		}

		if len(route.Secured) > 0 {
			var security []map[string][]string
			for _, scheme := range route.Secured {
				security = append(security, map[string][]string{scheme: {}})
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
			Schemas:         schemas,
			SecuritySchemes: securitySchemes,
		},
	}
}

// schemaRef registers a struct as a named schema in components and returns a $ref.
// If the type is anonymous or not a struct, falls back to inline schema.
func schemaRef(v any, schemas map[string]any) map[string]any {
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

	name := t.Name()

	// Anonymous structs — generate inline, no $ref
	if name == "" {
		return reflectSchema(v)
	}

	// Register in components/schemas if not already there
	if _, exists := schemas[name]; !exists {
		schemas[name] = reflectSchema(v)
	}

	return map[string]any{
		"$ref": fmt.Sprintf("#/components/schemas/%s", name),
	}
}

func buildRequestBody(dto any, schemas map[string]any) map[string]any {
	return map[string]any{
		"required": true,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": schemaRef(dto, schemas),
			},
		},
	}
}

func buildResponses(route RouteInput, schemas map[string]any) map[string]any {
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
					"schema": schemaRef(route.Response, schemas),
				},
			},
		}
	}
	return responses
}

// reflectSchema generates an OpenAPI schema from a struct.
// Reads tags: json, validate, doc, example, format.
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

		oaType, oaFormat := goTypeToOA(field.Type.Kind())
		prop := map[string]any{
			"type": oaType,
		}

		// format from tag takes priority over inferred format
		if f := field.Tag.Get("format"); f != "" {
			prop["format"] = f
		} else if oaFormat != "" {
			prop["format"] = oaFormat
		}

		// format inferred from validate tags
		if prop["format"] == nil {
			validateTag := field.Tag.Get("validate")
			if strings.Contains(validateTag, "email") {
				prop["format"] = "email"
			} else if strings.Contains(validateTag, "uuid") {
				prop["format"] = "uuid"
			} else if strings.Contains(validateTag, "url") {
				prop["format"] = "uri"
			}
		}

		if doc := field.Tag.Get("doc"); doc != "" {
			prop["description"] = doc
		}
		if example := field.Tag.Get("example"); example != "" {
			prop["example"] = example
		}

		// min/max from validate tag
		validateTag := field.Tag.Get("validate")
		if min := extractParam(validateTag, "min"); min != "" {
			if oaType == "string" {
				prop["minLength"] = toInt(min)
			} else {
				prop["minimum"] = toInt(min)
			}
		}
		if max := extractParam(validateTag, "max"); max != "" {
			if oaType == "string" {
				prop["maxLength"] = toInt(max)
			} else {
				prop["maximum"] = toInt(max)
			}
		}

		properties[name] = prop

		if strings.Contains(validateTag, "required") {
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

// goTypeToOA returns the OpenAPI type and optional format for a Go kind.
func goTypeToOA(k reflect.Kind) (string, string) {
	switch k {
	case reflect.String:
		return "string", ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return "integer", "int32"
	case reflect.Int64:
		return "integer", "int64"
	case reflect.Float32:
		return "number", "float"
	case reflect.Float64:
		return "number", "double"
	case reflect.Bool:
		return "boolean", ""
	default:
		return "string", ""
	}
}

// extractParam extracts the value of a validate param e.g. "min=8" → "8".
func extractParam(tag, key string) string {
	for _, part := range strings.Split(tag, ",") {
		if strings.HasPrefix(part, key+"=") {
			return strings.TrimPrefix(part, key+"=")
		}
	}
	return ""
}

// toInt converts a string to int, returns 0 on failure.
func toInt(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

// inferSecurityScheme infers the scheme type by name convention.
func inferSecurityScheme(name string) SecurityScheme {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "bearer"):
		return SecurityScheme{Type: "http", Scheme: "bearer", BearerFormat: "JWT"}
	case strings.Contains(lower, "basic"):
		return SecurityScheme{Type: "http", Scheme: "basic"}
	default:
		return SecurityScheme{Type: "apiKey", In: "header", Name: "X-API-Key"}
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
