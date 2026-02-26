package openapi

import (
	"reflect"
	"testing"
)

func TestFiberPathToOA(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "simple path no params", input: "/users", want: "/users"},
		{name: "single param", input: "/users/:id", want: "/users/{id}"},
		{name: "nested path with param", input: "/users/:id/posts", want: "/users/{id}/posts"},
		{name: "multiple params", input: "/users/:userId/posts/:postId", want: "/users/{userId}/posts/{postId}"},
		{name: "root path", input: "/", want: "/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fiberPathToOA(tt.input)
			if got != tt.want {
				t.Errorf("fiberPathToOA() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInferSecurityScheme(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantType   string
		wantScheme string
		wantIn     string
	}{
		{name: "bearerAuth → HTTP Bearer", input: "bearerAuth", wantType: "http", wantScheme: "bearer"},
		{name: "myBearerToken → HTTP Bearer", input: "myBearerToken", wantType: "http", wantScheme: "bearer"},
		{name: "basicAuth → HTTP Basic", input: "basicAuth", wantType: "http", wantScheme: "basic"},
		{name: "apiKey → API Key in header", input: "apiKey", wantType: "apiKey", wantIn: "header"},
		{name: "unknown → fallback API Key", input: "somethingElse", wantType: "apiKey", wantIn: "header"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferSecurityScheme(tt.input)
			if got.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", got.Type, tt.wantType)
			}
			if tt.wantScheme != "" && got.Scheme != tt.wantScheme {
				t.Errorf("Scheme = %v, want %v", got.Scheme, tt.wantScheme)
			}
			if tt.wantIn != "" && got.In != tt.wantIn {
				t.Errorf("In = %v, want %v", got.In, tt.wantIn)
			}
		})
	}
}

// goTypeToOA returns (type, format) — tests verify both values.
func TestGoTypeToOA(t *testing.T) {
	tests := []struct {
		name     string
		input    reflect.Kind
		wantType string
		wantFmt  string
	}{
		{name: "string", input: reflect.String, wantType: "string", wantFmt: ""},
		{name: "int", input: reflect.Int, wantType: "integer", wantFmt: ""},
		{name: "int32", input: reflect.Int32, wantType: "integer", wantFmt: "int32"},
		{name: "int64", input: reflect.Int64, wantType: "integer", wantFmt: "int64"},
		{name: "float32", input: reflect.Float32, wantType: "number", wantFmt: "float"},
		{name: "float64", input: reflect.Float64, wantType: "number", wantFmt: "double"},
		{name: "bool", input: reflect.Bool, wantType: "boolean", wantFmt: ""},
		{name: "unknown defaults to string", input: reflect.Slice, wantType: "string", wantFmt: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotFmt := goTypeToOA(tt.input)
			if gotType != tt.wantType {
				t.Errorf("type = %v, want %v", gotType, tt.wantType)
			}
			if gotFmt != tt.wantFmt {
				t.Errorf("format = %v, want %v", gotFmt, tt.wantFmt)
			}
		})
	}
}

func TestReflectSchema(t *testing.T) {
	type userDTO struct {
		Name  string `json:"name"  validate:"required" doc:"Full name"   example:"Juan"`
		Email string `json:"email" validate:"required" doc:"Email address"`
		Age   int    `json:"age"`
		Admin bool   `json:"admin"`
	}

	tests := []struct {
		name           string
		input          any
		wantProperties []string
		wantRequired   []string
		wantType       string
	}{
		{
			name:           "struct with all field types",
			input:          userDTO{},
			wantProperties: []string{"name", "email", "age", "admin"},
			wantRequired:   []string{"name", "email"},
			wantType:       "object",
		},
		{
			name:     "nil input returns object",
			input:    nil,
			wantType: "object",
		},
		{
			name:     "non-struct returns object",
			input:    "string value",
			wantType: "object",
		},
		{
			name: "struct with ignored field",
			input: struct {
				Name    string `json:"name"`
				Ignored string `json:"-"`
				Empty   string
			}{},
			wantProperties: []string{"name"},
			wantType:       "object",
		},
		{
			name: "struct with doc and example tags",
			input: struct {
				Name string `json:"name" doc:"Full name" example:"Juan"`
			}{},
			wantProperties: []string{"name"},
			wantType:       "object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reflectSchema(tt.input, map[string]any{})

			if got["type"] != tt.wantType {
				t.Errorf("type = %v, want %v", got["type"], tt.wantType)
			}

			if len(tt.wantProperties) > 0 {
				props, ok := got["properties"].(map[string]any)
				if !ok {
					t.Fatal("properties should be a map")
				}
				for _, prop := range tt.wantProperties {
					if _, exists := props[prop]; !exists {
						t.Errorf("missing property %q", prop)
					}
				}
			}

			if len(tt.wantRequired) > 0 {
				required, ok := got["required"].([]string)
				if !ok {
					t.Fatal("required should be a []string")
				}
				for _, req := range tt.wantRequired {
					found := false
					for _, r := range required {
						if r == req {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("missing required field %q", req)
					}
				}
			}
		})
	}
}

func TestReflectSchemaFormats(t *testing.T) {
	type formatsDTO struct {
		Email string  `json:"email" validate:"required,email"`
		ID    string  `json:"id"    validate:"required,uuid4"`
		URL   string  `json:"url"   validate:"required,url"`
		Age   int32   `json:"age"`
		Score float64 `json:"score"`
	}

	got := reflectSchema(formatsDTO{}, map[string]any{})
	props, ok := got["properties"].(map[string]any)
	if !ok {
		t.Fatal("properties should be a map")
	}

	cases := []struct {
		field      string
		wantFormat string
	}{
		{"email", "email"},
		{"id", "uuid"},
		{"url", "uri"},
		{"age", "int32"},
		{"score", "double"},
	}

	for _, c := range cases {
		prop, ok := props[c.field].(map[string]any)
		if !ok {
			t.Errorf("field %q not found in properties", c.field)
			continue
		}
		if prop["format"] != c.wantFormat {
			t.Errorf("field %q format = %v, want %v", c.field, prop["format"], c.wantFormat)
		}
	}
}

func TestReflectSchemaMinMax(t *testing.T) {
	type minMaxDTO struct {
		Name string `json:"name" validate:"required,min=2,max=50"`
		Age  int    `json:"age"  validate:"min=18,max=120"`
	}

	got := reflectSchema(minMaxDTO{}, map[string]any{})
	props, ok := got["properties"].(map[string]any)
	if !ok {
		t.Fatal("properties should be a map")
	}

	name, ok := props["name"].(map[string]any)
	if !ok {
		t.Fatal("name property not found")
	}
	if name["minLength"] != 2 {
		t.Errorf("name minLength = %v, want 2", name["minLength"])
	}
	if name["maxLength"] != 50 {
		t.Errorf("name maxLength = %v, want 50", name["maxLength"])
	}

	age, ok := props["age"].(map[string]any)
	if !ok {
		t.Fatal("age property not found")
	}
	if age["minimum"] != 18 {
		t.Errorf("age minimum = %v, want 18", age["minimum"])
	}
	if age["maximum"] != 120 {
		t.Errorf("age maximum = %v, want 120", age["maximum"])
	}
}

func TestReflectSchemaNested(t *testing.T) {
	type AddressDTO struct {
		Street string `json:"street"`
	}
	type PersonDTO struct {
		Name    string     `json:"name"`
		Address AddressDTO `json:"address"`
	}

	schemas := map[string]any{}
	got := reflectSchema(PersonDTO{}, schemas)

	props, ok := got["properties"].(map[string]any)
	if !ok {
		t.Fatal("properties should be a map")
	}
	addr, ok := props["address"].(map[string]any)
	if !ok {
		t.Fatal("address property not found")
	}
	if addr["$ref"] != "#/components/schemas/AddressDTO" {
		t.Errorf("address $ref = %v, want #/components/schemas/AddressDTO", addr["$ref"])
	}
	if _, exists := schemas["AddressDTO"]; !exists {
		t.Error("AddressDTO should be registered in schemas")
	}
}

func TestReflectSchemaSlice(t *testing.T) {
	type TagDTO struct {
		Name string `json:"name"`
	}
	type PostDTO struct {
		Tags []TagDTO `json:"tags"`
	}

	schemas := map[string]any{}
	got := reflectSchema(PostDTO{}, schemas)

	props, ok := got["properties"].(map[string]any)
	if !ok {
		t.Fatal("properties should be a map")
	}
	tags, ok := props["tags"].(map[string]any)
	if !ok {
		t.Fatal("tags property not found")
	}
	if tags["type"] != "array" {
		t.Errorf("tags type = %v, want array", tags["type"])
	}
	items, ok := tags["items"].(map[string]any)
	if !ok {
		t.Fatal("tags items not found")
	}
	if items["$ref"] != "#/components/schemas/TagDTO" {
		t.Errorf("items.$ref = %v, want #/components/schemas/TagDTO", items["$ref"])
	}
	if _, exists := schemas["TagDTO"]; !exists {
		t.Error("TagDTO should be registered in schemas")
	}
}

func TestReflectSchemaPointer(t *testing.T) {
	type DTO struct {
		Name *string `json:"name"`
	}

	schemas := map[string]any{}
	got := reflectSchema(DTO{}, schemas)

	props, ok := got["properties"].(map[string]any)
	if !ok {
		t.Fatal("properties should be a map")
	}
	name, ok := props["name"].(map[string]any)
	if !ok {
		t.Fatal("name property not found")
	}
	if name["nullable"] != true {
		t.Errorf("name nullable = %v, want true", name["nullable"])
	}
	if name["type"] != "string" {
		t.Errorf("name type = %v, want string", name["type"])
	}
}

func TestReflectSchemaEnum(t *testing.T) {
	type DTO struct {
		Role string `json:"role" validate:"required,oneof=admin user"`
	}

	schemas := map[string]any{}
	got := reflectSchema(DTO{}, schemas)

	props, ok := got["properties"].(map[string]any)
	if !ok {
		t.Fatal("properties should be a map")
	}
	role, ok := props["role"].(map[string]any)
	if !ok {
		t.Fatal("role property not found")
	}
	enum, ok := role["enum"].([]string)
	if !ok {
		t.Fatalf("role enum should be []string, got %T", role["enum"])
	}
	if len(enum) != 2 || enum[0] != "admin" || enum[1] != "user" {
		t.Errorf("role enum = %v, want [admin user]", enum)
	}
}

func TestReflectSchemaDefault(t *testing.T) {
	type DTO struct {
		Status string `json:"status" default:"active"`
	}

	schemas := map[string]any{}
	got := reflectSchema(DTO{}, schemas)

	props, ok := got["properties"].(map[string]any)
	if !ok {
		t.Fatal("properties should be a map")
	}
	status, ok := props["status"].(map[string]any)
	if !ok {
		t.Fatal("status property not found")
	}
	if status["default"] != "active" {
		t.Errorf("status default = %v, want active", status["default"])
	}
}

func TestBuild(t *testing.T) {
	type responseDTO struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	type bodyDTO struct {
		Name string `json:"name" validate:"required"`
	}

	tests := []struct {
		name           string
		input          BuildInput
		wantTitle      string
		wantVersion    string
		wantPaths      []string
		wantSecSchemes []string
		wantSchemas    []string
	}{
		{
			name: "basic build with title and version",
			input: BuildInput{
				Title:   "Test API",
				Version: "1.0.0",
				Routes:  []RouteInput{},
			},
			wantTitle:   "Test API",
			wantVersion: "1.0.0",
		},
		{
			name: "build registers paths",
			input: BuildInput{
				Title:   "Test API",
				Version: "1.0.0",
				Routes: []RouteInput{
					{Method: "GET", Path: "/users", Summary: "List users"},
					{Method: "POST", Path: "/users", Summary: "Create user"},
				},
			},
			wantTitle:   "Test API",
			wantVersion: "1.0.0",
			wantPaths:   []string{"/users"},
		},
		{
			name: "build converts fiber params to openapi",
			input: BuildInput{
				Title:   "Test API",
				Version: "1.0.0",
				Routes: []RouteInput{
					{Method: "GET", Path: "/users/:id"},
				},
			},
			wantTitle:   "Test API",
			wantVersion: "1.0.0",
			wantPaths:   []string{"/users/{id}"},
		},
		{
			name: "build registers security schemes from routes",
			input: BuildInput{
				Title:   "Test API",
				Version: "1.0.0",
				Routes: []RouteInput{
					{
						Method:  "GET",
						Path:    "/users",
						Secured: []string{"bearerAuth", "apiKey"},
					},
				},
			},
			wantTitle:      "Test API",
			wantVersion:    "1.0.0",
			wantSecSchemes: []string{"bearerAuth", "apiKey"},
		},
		{
			name: "build with body and response registers named schemas",
			input: BuildInput{
				Title:   "Test API",
				Version: "1.0.0",
				Routes: []RouteInput{
					{
						Method:     "POST",
						Path:       "/users",
						Body:       bodyDTO{},
						Response:   responseDTO{},
						StatusCode: 201,
					},
				},
			},
			wantTitle:   "Test API",
			wantVersion: "1.0.0",
			wantPaths:   []string{"/users"},
			wantSchemas: []string{"bodyDTO", "responseDTO"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Build(tt.input)

			if got.Info.Title != tt.wantTitle {
				t.Errorf("Title = %v, want %v", got.Info.Title, tt.wantTitle)
			}
			if got.Info.Version != tt.wantVersion {
				t.Errorf("Version = %v, want %v", got.Info.Version, tt.wantVersion)
			}
			if got.OpenAPI != "3.0.0" {
				t.Errorf("OpenAPI = %v, want 3.0.0", got.OpenAPI)
			}

			for _, path := range tt.wantPaths {
				if _, exists := got.Paths[path]; !exists {
					t.Errorf("missing path %q in spec", path)
				}
			}

			for _, scheme := range tt.wantSecSchemes {
				if _, exists := got.Components.SecuritySchemes[scheme]; !exists {
					t.Errorf("missing security scheme %q in components", scheme)
				}
			}

			for _, schema := range tt.wantSchemas {
				if _, exists := got.Components.Schemas[schema]; !exists {
					t.Errorf("missing schema %q in components/schemas", schema)
				}
			}
		})
	}
}

func TestBuildQueryParameters(t *testing.T) {
	tests := []struct {
		name         string
		params       []QueryParamInput
		wantLen      int
		wantNames    []string
		wantTypes    []string
		wantRequired []bool
		wantDescs    []string
	}{
		{
			name: "single optional string param",
			params: []QueryParamInput{
				{Name: "status", Type: "string", Required: false},
			},
			wantLen:      1,
			wantNames:    []string{"status"},
			wantTypes:    []string{"string"},
			wantRequired: []bool{false},
		},
		{
			name: "required param with description",
			params: []QueryParamInput{
				{Name: "q", Type: "string", Required: true, Description: "Search query"},
			},
			wantLen:      1,
			wantNames:    []string{"q"},
			wantRequired: []bool{true},
			wantDescs:    []string{"Search query"},
		},
		{
			name: "empty type defaults to string",
			params: []QueryParamInput{
				{Name: "filter"},
			},
			wantLen:   1,
			wantNames: []string{"filter"},
			wantTypes: []string{"string"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildQueryParameters(tt.params)
			if len(got) != tt.wantLen {
				t.Errorf("len = %v, want %v", len(got), tt.wantLen)
				return
			}
			for i, p := range got {
				if len(tt.wantNames) > i && p["name"] != tt.wantNames[i] {
					t.Errorf("[%d] name = %v, want %v", i, p["name"], tt.wantNames[i])
				}
				if p["in"] != "query" {
					t.Errorf("[%d] in = %v, want query", i, p["in"])
				}
				if len(tt.wantRequired) > i && p["required"] != tt.wantRequired[i] {
					t.Errorf("[%d] required = %v, want %v", i, p["required"], tt.wantRequired[i])
				}
				if len(tt.wantDescs) > i && tt.wantDescs[i] != "" && p["description"] != tt.wantDescs[i] {
					t.Errorf("[%d] description = %v, want %v", i, p["description"], tt.wantDescs[i])
				}
				schema, ok := p["schema"].(map[string]any)
				if !ok {
					t.Errorf("[%d] schema missing", i)
					continue
				}
				if len(tt.wantTypes) > i && schema["type"] != tt.wantTypes[i] {
					t.Errorf("[%d] schema.type = %v, want %v", i, schema["type"], tt.wantTypes[i])
				}
			}
		})
	}
}

func TestBuildIncludesQueryParamsInOperation(t *testing.T) {
	spec := Build(BuildInput{
		Title:   "Test",
		Version: "1.0.0",
		Routes: []RouteInput{
			{
				Method: "GET",
				Path:   "/users",
				QueryParams: []QueryParamInput{
					{Name: "page", Type: "integer", Required: false},
					{Name: "limit", Type: "integer", Required: false},
				},
			},
		},
	})

	pathItem, ok := spec.Paths["/users"].(map[string]any)
	if !ok {
		t.Fatal("path /users not found")
	}
	operation, ok := pathItem["get"].(map[string]any)
	if !ok {
		t.Fatal("operation get not found")
	}
	params, ok := operation["parameters"].([]map[string]any)
	if !ok {
		t.Fatal("parameters should be []map[string]any")
	}
	if len(params) != 2 {
		t.Errorf("parameters len = %v, want 2", len(params))
	}
}

func TestBuildPathParameters(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantLen  int
		wantName string
	}{
		{name: "no path params", path: "/users", wantLen: 0},
		{name: "single path param", path: "/users/:id", wantLen: 1, wantName: "id"},
		{name: "multiple path params", path: "/users/:userId/posts/:postId", wantLen: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildPathParameters(tt.path)
			if len(got) != tt.wantLen {
				t.Errorf("len = %v, want %v", len(got), tt.wantLen)
				return
			}
			if tt.wantLen > 0 {
				if got[0]["in"] != "path" {
					t.Errorf("in = %v, want path", got[0]["in"])
				}
				if got[0]["required"] != true {
					t.Errorf("required = %v, want true", got[0]["required"])
				}
				if tt.wantName != "" && got[0]["name"] != tt.wantName {
					t.Errorf("name = %v, want %v", got[0]["name"], tt.wantName)
				}
			}
		})
	}
}

func TestBuildAutoErrorResponses(t *testing.T) {
	t.Run("body present adds 400 and 422", func(t *testing.T) {
		type B struct{ Name string `json:"name"` }
		route := RouteInput{Method: "POST", Path: "/users", Body: B{}}
		got := buildAutoErrorResponses(route)
		if _, ok := got["400"]; !ok {
			t.Error("missing 400 response")
		}
		if _, ok := got["422"]; !ok {
			t.Error("missing 422 response")
		}
	})

	t.Run("secured adds 401 and 403", func(t *testing.T) {
		route := RouteInput{Method: "GET", Path: "/users", Secured: []string{"bearerAuth"}}
		got := buildAutoErrorResponses(route)
		if _, ok := got["401"]; !ok {
			t.Error("missing 401 response")
		}
		if _, ok := got["403"]; !ok {
			t.Error("missing 403 response")
		}
	})

	t.Run("path params adds 404", func(t *testing.T) {
		route := RouteInput{Method: "GET", Path: "/users/:id"}
		got := buildAutoErrorResponses(route)
		if _, ok := got["404"]; !ok {
			t.Error("missing 404 response")
		}
	})

	t.Run("no path params no 404", func(t *testing.T) {
		route := RouteInput{Method: "GET", Path: "/users"}
		got := buildAutoErrorResponses(route)
		if _, ok := got["404"]; ok {
			t.Error("404 should not be present for route without path params")
		}
	})

	t.Run("always adds 500", func(t *testing.T) {
		route := RouteInput{Method: "GET", Path: "/users"}
		got := buildAutoErrorResponses(route)
		if _, ok := got["500"]; !ok {
			t.Error("missing 500 response")
		}
	})
}

func TestBuildOperationID(t *testing.T) {
	tests := []struct {
		method string
		path   string
		want   string
	}{
		{"GET", "/users/:id", "getUsersById"},
		{"POST", "/v1/users", "postV1Users"},
		{"DELETE", "/users/:id", "deleteUsersById"},
		{"GET", "/users", "getUsers"},
		{"PATCH", "/users/:id/posts/:postId", "patchUsersByIdPostsByPostId"},
	}
	for _, tt := range tests {
		got := generateOperationID(tt.method, tt.path)
		if got != tt.want {
			t.Errorf("generateOperationID(%q, %q) = %v, want %v", tt.method, tt.path, got, tt.want)
		}
	}
}

func TestBuildDeprecated(t *testing.T) {
	spec := Build(BuildInput{
		Title:   "Test",
		Version: "1.0.0",
		Routes: []RouteInput{
			{Method: "GET", Path: "/users", Deprecated: true},
		},
	})

	pathItem, ok := spec.Paths["/users"].(map[string]any)
	if !ok {
		t.Fatal("path /users not found")
	}
	operation, ok := pathItem["get"].(map[string]any)
	if !ok {
		t.Fatal("operation get not found")
	}
	if operation["deprecated"] != true {
		t.Errorf("deprecated = %v, want true", operation["deprecated"])
	}
}

func TestBuildServersAndTags(t *testing.T) {
	spec := Build(BuildInput{
		Title:   "Test",
		Version: "1.0.0",
		Servers: []ServerInfo{{URL: "https://api.example.com", Description: "Production"}},
		Tags:    []TagInfo{{Name: "users", Description: "User operations"}},
		Routes:  []RouteInput{},
	})

	if len(spec.Servers) != 1 || spec.Servers[0].URL != "https://api.example.com" {
		t.Errorf("servers = %v, want [{URL: https://api.example.com}]", spec.Servers)
	}
	if len(spec.Tags) != 1 || spec.Tags[0].Name != "users" {
		t.Errorf("tags = %v, want [{Name: users}]", spec.Tags)
	}
}

func TestBuildStandardSchemas(t *testing.T) {
	spec := Build(BuildInput{
		Title:   "Test",
		Version: "1.0.0",
		Routes:  []RouteInput{},
	})

	for _, name := range []string{"KErrorResponse", "ValidationErrorResponse", "ValidationErrorItem"} {
		if _, exists := spec.Components.Schemas[name]; !exists {
			t.Errorf("missing schema %q in components/schemas", name)
		}
	}
}

func TestBuildOperationIncludesPathParamsWhenPresent(t *testing.T) {
	spec := Build(BuildInput{
		Title:   "Test",
		Version: "1.0.0",
		Routes: []RouteInput{
			{Method: "GET", Path: "/users/:id"},
		},
	})

	pathItem, ok := spec.Paths["/users/{id}"].(map[string]any)
	if !ok {
		t.Fatal("path /users/{id} not found")
	}
	operation, ok := pathItem["get"].(map[string]any)
	if !ok {
		t.Fatal("operation get not found")
	}
	params, ok := operation["parameters"].([]map[string]any)
	if !ok {
		t.Fatal("parameters should be []map[string]any")
	}
	if len(params) != 1 {
		t.Errorf("parameters len = %v, want 1", len(params))
		return
	}
	if params[0]["name"] != "id" {
		t.Errorf("param name = %v, want id", params[0]["name"])
	}
	if params[0]["in"] != "path" {
		t.Errorf("param in = %v, want path", params[0]["in"])
	}
	if params[0]["required"] != true {
		t.Errorf("param required = %v, want true", params[0]["required"])
	}
}
