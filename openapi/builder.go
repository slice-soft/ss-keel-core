package openapi

import (
	"fmt"
	"reflect"
	"strings"
)

// TagInfo describes an OpenAPI tag with a description.
type TagInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// ServerInfo describes an API server.
type ServerInfo struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// Contact holds API contact information.
type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// License holds API license information.
type License struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// Spec is the in-memory representation of an OpenAPI 3.0 spec.
type Spec struct {
	OpenAPI    string                `json:"openapi"`
	Info       Info                  `json:"info"`
	Servers    []ServerInfo          `json:"servers,omitempty"`
	Tags       []TagInfo             `json:"tags,omitempty"`
	Paths      map[string]any        `json:"paths"`
	Components Components            `json:"components"`
	Security   []map[string][]string `json:"security,omitempty"`
}

type Info struct {
	Title       string   `json:"title"`
	Version     string   `json:"version"`
	Description string   `json:"description,omitempty"`
	Contact     *Contact `json:"contact,omitempty"`
	License     *License `json:"license,omitempty"`
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

// QueryParamInput documents a query string parameter.
type QueryParamInput struct {
	Name        string
	Type        string
	Description string
	Required    bool
}

// RouteInput is the neutral representation of a route.
type RouteInput struct {
	Method      string
	Path        string
	Summary     string
	Description string
	Tags        []string
	Secured     []string // security schemes: "bearerAuth", "apiKey", etc.
	Body        any
	Response    any
	StatusCode  int
	QueryParams []QueryParamInput
	Deprecated  bool
}

// BuildInput groups the data to build the spec.
type BuildInput struct {
	Title       string
	Version     string
	Description string
	Contact     *Contact
	License     *License
	Servers     []ServerInfo
	Tags        []TagInfo
	Routes      []RouteInput
}

// Build constructs the OpenAPI 3.0 spec in memory.
func Build(input BuildInput) Spec {
	paths := make(map[string]any)
	schemas := make(map[string]any)
	securitySchemes := make(map[string]SecurityScheme)

	// Pre-register standard error schemas
	registerStandardSchemas(schemas)

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
			"operationId": generateOperationID(route.Method, route.Path),
		}

		// Parameters: path params first, then query params
		pathParams := buildPathParameters(route.Path)
		queryParams := buildQueryParameters(route.QueryParams)
		parameters := append(pathParams, queryParams...)
		if len(parameters) > 0 {
			operation["parameters"] = parameters
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

		if route.Deprecated {
			operation["deprecated"] = true
		}

		method := strings.ToLower(route.Method)
		paths[oaPath].(map[string]any)[method] = operation
	}

	return Spec{
		OpenAPI: "3.0.0",
		Info: Info{
			Title:       input.Title,
			Version:     input.Version,
			Description: input.Description,
			Contact:     input.Contact,
			License:     input.License,
		},
		Servers: input.Servers,
		Tags:    input.Tags,
		Paths:   paths,
		Components: Components{
			Schemas:         schemas,
			SecuritySchemes: securitySchemes,
		},
	}
}

// registerStandardSchemas pre-registers standard error schemas used by auto error responses.
func registerStandardSchemas(schemas map[string]any) {
	schemas["KErrorResponse"] = map[string]any{
		"type": "object",
		"properties": map[string]any{
			"status_code": map[string]any{"type": "integer"},
			"code":        map[string]any{"type": "string"},
			"message":     map[string]any{"type": "string"},
		},
		"required": []string{"status_code", "code", "message"},
	}
	schemas["ValidationErrorItem"] = map[string]any{
		"type": "object",
		"properties": map[string]any{
			"field":   map[string]any{"type": "string"},
			"message": map[string]any{"type": "string"},
		},
		"required": []string{"field", "message"},
	}
	schemas["ValidationErrorResponse"] = map[string]any{
		"type": "object",
		"properties": map[string]any{
			"status_code": map[string]any{"type": "integer"},
			"message":     map[string]any{"type": "string"},
			"errors": map[string]any{
				"type":  "array",
				"items": map[string]any{"$ref": "#/components/schemas/ValidationErrorItem"},
			},
		},
		"required": []string{"status_code", "message", "errors"},
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
		return reflectSchema(v, schemas)
	}

	// Register in components/schemas if not already there
	if _, exists := schemas[name]; !exists {
		schemas[name] = reflectSchema(v, schemas)
	}

	return map[string]any{
		"$ref": fmt.Sprintf("#/components/schemas/%s", name),
	}
}

// buildPathParameters extracts :param segments from a Fiber path and generates OpenAPI path parameters.
func buildPathParameters(fiberPath string) []map[string]any {
	var params []map[string]any
	for _, part := range strings.Split(fiberPath, "/") {
		if strings.HasPrefix(part, ":") {
			params = append(params, map[string]any{
				"name":     part[1:],
				"in":       "path",
				"required": true,
				"schema":   map[string]any{"type": "string"},
			})
		}
	}
	return params
}

func buildQueryParameters(params []QueryParamInput) []map[string]any {
	var out []map[string]any
	for _, p := range params {
		typ := p.Type
		if typ == "" {
			typ = "string"
		}
		param := map[string]any{
			"name":     p.Name,
			"in":       "query",
			"required": p.Required,
			"schema":   map[string]any{"type": typ},
		}
		if p.Description != "" {
			param["description"] = p.Description
		}
		out = append(out, param)
	}
	return out
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

// buildAutoErrorResponses generates automatic error responses based on route properties.
func buildAutoErrorResponses(route RouteInput) map[string]any {
	errs := map[string]any{}

	kerrorContent := map[string]any{
		"application/json": map[string]any{
			"schema": map[string]any{"$ref": "#/components/schemas/KErrorResponse"},
		},
	}

	if route.Body != nil {
		errs["400"] = map[string]any{
			"description": "Bad Request",
			"content":     kerrorContent,
		}
		errs["422"] = map[string]any{
			"description": "Validation Error",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{"$ref": "#/components/schemas/ValidationErrorResponse"},
				},
			},
		}
	}

	if len(route.Secured) > 0 {
		errs["401"] = map[string]any{
			"description": "Unauthorized",
			"content":     kerrorContent,
		}
		errs["403"] = map[string]any{
			"description": "Forbidden",
			"content":     kerrorContent,
		}
	}

	for _, part := range strings.Split(route.Path, "/") {
		if strings.HasPrefix(part, ":") {
			errs["404"] = map[string]any{
				"description": "Not Found",
				"content":     kerrorContent,
			}
			break
		}
	}

	errs["500"] = map[string]any{
		"description": "Internal Server Error",
		"content":     kerrorContent,
	}

	return errs
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

	// Merge auto error responses
	for k, v := range buildAutoErrorResponses(route) {
		responses[k] = v
	}

	return responses
}

// generateOperationID generates an operationId from the HTTP method and path.
// Examples: GET /users/:id → getUsersById, POST /v1/users → postV1Users
func generateOperationID(method, path string) string {
	result := strings.ToLower(method)
	for _, part := range strings.Split(path, "/") {
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, ":") {
			param := part[1:]
			result += "By" + strings.ToUpper(param[:1]) + param[1:]
		} else {
			result += strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return result
}

// fieldSchema generates an OpenAPI schema for a single struct field, handling complex types.
func fieldSchema(field reflect.StructField, schemas map[string]any) map[string]any {
	t := field.Type

	// Special case: time.Time → date-time string
	if t.PkgPath() == "time" && t.Name() == "Time" {
		return map[string]any{"type": "string", "format": "date-time"}
	}

	switch t.Kind() {
	case reflect.Struct:
		return schemaRef(reflect.New(t).Interface(), schemas)
	case reflect.Slice:
		elem := t.Elem()
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		if elem.Kind() == reflect.Struct {
			return map[string]any{
				"type":  "array",
				"items": schemaRef(reflect.New(elem).Interface(), schemas),
			}
		}
		oaType, _ := goTypeToOA(elem.Kind())
		return map[string]any{
			"type":  "array",
			"items": map[string]any{"type": oaType},
		}
	case reflect.Ptr:
		elem := t.Elem()
		if elem.Kind() == reflect.Struct {
			ref := schemaRef(reflect.New(elem).Interface(), schemas)
			return map[string]any{
				"allOf":    []any{ref},
				"nullable": true,
			}
		}
		oaType, oaFormat := goTypeToOA(elem.Kind())
		prop := map[string]any{"type": oaType, "nullable": true}
		if oaFormat != "" {
			prop["format"] = oaFormat
		}
		return prop
	case reflect.Map:
		return map[string]any{"type": "object", "additionalProperties": true}
	default:
		oaType, oaFormat := goTypeToOA(t.Kind())
		prop := map[string]any{"type": oaType}
		if oaFormat != "" {
			prop["format"] = oaFormat
		}
		return prop
	}
}

// reflectSchema generates an OpenAPI schema from a struct.
// Reads tags: json, validate, doc, example, format, default.
func reflectSchema(v any, schemas map[string]any) map[string]any {
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
		if name == "" || name == "-" {
			continue
		}

		prop := fieldSchema(field, schemas)
		validateTag := field.Tag.Get("validate")

		// Universal metadata tags apply to all field types
		if doc := field.Tag.Get("doc"); doc != "" {
			prop["description"] = doc
		}
		if example := field.Tag.Get("example"); example != "" {
			prop["example"] = example
		}
		if def := field.Tag.Get("default"); def != "" {
			prop["default"] = def
		}

		// Primitive-specific enrichments (not structs, slices, ptrs, or maps)
		kind := field.Type.Kind()
		isPrimitive := kind != reflect.Struct && kind != reflect.Slice && kind != reflect.Ptr && kind != reflect.Map

		if isPrimitive {
			// format from tag takes priority over inferred format
			if f := field.Tag.Get("format"); f != "" {
				prop["format"] = f
			} else if prop["format"] == nil {
				if strings.Contains(validateTag, "email") {
					prop["format"] = "email"
				} else if strings.Contains(validateTag, "uuid") {
					prop["format"] = "uuid"
				} else if strings.Contains(validateTag, "url") {
					prop["format"] = "uri"
				}
			}

			if min := extractParam(validateTag, "min"); min != "" {
				if prop["type"] == "string" {
					prop["minLength"] = toInt(min)
				} else {
					prop["minimum"] = toInt(min)
				}
			}
			if max := extractParam(validateTag, "max"); max != "" {
				if prop["type"] == "string" {
					prop["maxLength"] = toInt(max)
				} else {
					prop["maximum"] = toInt(max)
				}
			}

			// enum from oneof validate tag
			if oneof := extractParam(validateTag, "oneof"); oneof != "" {
				prop["enum"] = strings.Split(oneof, " ")
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
	case reflect.Int, reflect.Int8, reflect.Int16: // ← Int sin format
		return "integer", ""
	case reflect.Int32:
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
