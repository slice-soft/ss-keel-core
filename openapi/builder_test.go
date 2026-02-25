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
			got := reflectSchema(tt.input)

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

	got := reflectSchema(formatsDTO{})
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

	got := reflectSchema(minMaxDTO{})
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
