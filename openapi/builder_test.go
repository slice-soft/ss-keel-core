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
		{
			name:  "simple path no params",
			input: "/users",
			want:  "/users",
		},
		{
			name:  "single param",
			input: "/users/:id",
			want:  "/users/{id}",
		},
		{
			name:  "nested path with param",
			input: "/users/:id/posts",
			want:  "/users/{id}/posts",
		},
		{
			name:  "multiple params",
			input: "/users/:userId/posts/:postId",
			want:  "/users/{userId}/posts/{postId}",
		},
		{
			name:  "root path",
			input: "/",
			want:  "/",
		},
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
		{
			name:       "bearerAuth → HTTP Bearer",
			input:      "bearerAuth",
			wantType:   "http",
			wantScheme: "bearer",
		},
		{
			name:       "myBearerToken → HTTP Bearer",
			input:      "myBearerToken",
			wantType:   "http",
			wantScheme: "bearer",
		},
		{
			name:       "basicAuth → HTTP Basic",
			input:      "basicAuth",
			wantType:   "http",
			wantScheme: "basic",
		},
		{
			name:     "apiKey → API Key in header",
			input:    "apiKey",
			wantType: "apiKey",
			wantIn:   "header",
		},
		{
			name:     "unknown → fallback API Key",
			input:    "somethingElse",
			wantType: "apiKey",
			wantIn:   "header",
		},
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

func TestGoTypeToOA(t *testing.T) {
	tests := []struct {
		name  string
		input reflect.Kind
		want  string
	}{
		{name: "string", input: reflect.String, want: "string"},
		{name: "int", input: reflect.Int, want: "integer"},
		{name: "int32", input: reflect.Int32, want: "integer"},
		{name: "int64", input: reflect.Int64, want: "integer"},
		{name: "float32", input: reflect.Float32, want: "number"},
		{name: "float64", input: reflect.Float64, want: "number"},
		{name: "bool", input: reflect.Bool, want: "boolean"},
		{name: "unknown defaults to string", input: reflect.Slice, want: "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := goTypeToOA(tt.input)
			if got != tt.want {
				t.Errorf("goTypeToOA() = %v, want %v", got, tt.want)
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
			name: "build with body and response",
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
		})
	}
}
