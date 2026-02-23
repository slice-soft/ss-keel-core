package core

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

func dummyHandler(ctx *Ctx) error {
	return nil
}
func dummyMiddleware() fiber.Handler {
	return func(ctx *fiber.Ctx) error { return ctx.Next() }
}

type testDTO struct {
	Name string `json:"name" validate:"required"`
}

type testResponseDTO struct {
	ID string `json:"id"`
}

func TestHTTPConstrcutors(t *testing.T) {
	tests := []struct {
		name       string
		route      Route
		wantMethod string
		wantPath   string
	}{
		{
			name:       "GET constructor",
			route:      GET("/users", dummyHandler),
			wantMethod: "GET",
			wantPath:   "/users",
		},
		{
			name:       "POST simple",
			route:      POST("/users", dummyHandler),
			wantMethod: "POST",
			wantPath:   "/users",
		},
		{
			name:       "PUT simple",
			route:      PUT("/users/:id", dummyHandler),
			wantMethod: "PUT",
			wantPath:   "/users/:id",
		},
		{
			name:       "PATCH simple",
			route:      PATCH("/users/:id", dummyHandler),
			wantMethod: "PATCH",
			wantPath:   "/users/:id",
		},
		{
			name:       "DELETE simple",
			route:      DELETE("/users/:id", dummyHandler),
			wantMethod: "DELETE",
			wantPath:   "/users/:id",
		},
		{
			name:       "path con param",
			route:      GET("/users/:id", dummyHandler),
			wantMethod: "GET",
			wantPath:   "/users/:id",
		},
		{
			name:       "path anidado",
			route:      GET("/users/:id/posts", dummyHandler),
			wantMethod: "GET",
			wantPath:   "/users/:id/posts",
		},
		{
			name:       "path whith parameters",
			route:      GET("/users/:id/books/:bookId", dummyHandler),
			wantMethod: "GET",
			wantPath:   "/users/:id/books/:bookId",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.route.Method() != tt.wantMethod {
				t.Errorf("Method() = %v, want %v", tt.route.Method(), tt.wantMethod)
			}
			if tt.route.Path() != tt.wantPath {
				t.Errorf("Path() = %v, want %v", tt.route.Path(), tt.wantPath)
			}
			if tt.route.Handler() == nil {
				t.Error("Handler() = nil, want non-nil")
			}
		})
	}
}

func TestBody(t *testing.T) {
	tests := []struct {
		name     string
		route    Route
		wantBody bool
		wantType any
	}{
		{
			name:     "whith body required",
			route:    POST("/users", dummyHandler).WithBody(WithBody[testDTO]()),
			wantBody: true,
			wantType: testDTO{},
		},
		{
			name:     "without body",
			route:    GET("/users", dummyHandler),
			wantBody: false,
		},
		{
			name: "body inline struct",
			route: POST("/users", dummyHandler).WithBody(WithBody[struct {
				Email string `json:"email"`
			}]()),
			wantBody: true,
		},
		{
			name: "body whith BodyMeta directly",
			route: POST("/users", dummyHandler).WithBody(&BodyMeta{
				Type:     testDTO{},
				Required: true,
			}),
			wantBody: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantBody && tt.route.Body() == nil {
				t.Error("Body() should not be nil")
			}
			if !tt.wantBody && tt.route.Body() != nil {
				t.Error("Body() should be nil")
			}
			if tt.wantBody && tt.route.Body() != nil && !tt.route.Body().Required {
				t.Error("Body().Required should be true")
			}
		})
	}
}

func TestResponse(t *testing.T) {
	tests := []struct {
		name           string
		route          Route
		wantResponse   bool
		wantStatusCode int
	}{
		{
			name:           "response 200",
			route:          GET("/users", dummyHandler).Res(WithResponse[testResponseDTO](200)),
			wantResponse:   true,
			wantStatusCode: 200,
		},
		{
			name:           "response 201",
			route:          POST("/users", dummyHandler).Res(WithResponse[testResponseDTO](201)),
			wantResponse:   true,
			wantStatusCode: 201,
		},
		{
			name:           "response 204 without body",
			route:          DELETE("/users/:id", dummyHandler).Res(WithResponse[struct{}](204)),
			wantResponse:   true,
			wantStatusCode: 204,
		},
		{
			name:         "without response",
			route:        GET("/users", dummyHandler),
			wantResponse: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantResponse && tt.route.Response() == nil {
				t.Error("Response() should not be nil")
			}
			if !tt.wantResponse && tt.route.Response() != nil {
				t.Error("Response() should be nil")
			}
			if tt.wantResponse && tt.route.Response() != nil {
				if tt.route.Response().StatusCode != tt.wantStatusCode {
					t.Errorf("StatusCode = %v, want %v", tt.route.Response().StatusCode, tt.wantStatusCode)
				}
			}
		})
	}
}

func TestTags(t *testing.T) {
	tests := []struct {
		name     string
		route    Route
		wantTags []string
	}{
		{
			name:     "single tag",
			route:    GET("/users", dummyHandler).Tag("users"),
			wantTags: []string{"users"},
		},
		{
			name:     "multiple tags with Tag()",
			route:    GET("/users", dummyHandler).Tag("users").Tag("admin").Tag("backoffice"),
			wantTags: []string{"users", "admin", "backoffice"},
		},
		{
			name:     "without tags",
			route:    GET("/users", dummyHandler),
			wantTags: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.route.Tags()
			if len(got) != len(tt.wantTags) {
				t.Errorf("Tags() len = %v, want %v", len(got), len(tt.wantTags))
				return
			}
			for i, tag := range got {
				if tag != tt.wantTags[i] {
					t.Errorf("Tags()[%d] = %v, want %v", i, tag, tt.wantTags[i])
				}
			}
		})
	}
}

func TestDescribe(t *testing.T) {
	tests := []struct {
		name            string
		route           Route
		wantSummary     string
		wantDescription string
	}{
		{
			name:        "only summary",
			route:       GET("/users", dummyHandler).Describe("List users"),
			wantSummary: "List users",
		},
		{
			name:            "summary and description",
			route:           GET("/users", dummyHandler).Describe("List users", "Returns all users"),
			wantSummary:     "List users",
			wantDescription: "Returns all users",
		},
		{
			name:        "without describe",
			route:       GET("/users", dummyHandler),
			wantSummary: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.route.Summary() != tt.wantSummary {
				t.Errorf("Summary() = %v, want %v", tt.route.Summary(), tt.wantSummary)
			}
			if tt.route.Description() != tt.wantDescription {
				t.Errorf("Description() = %v, want %v", tt.route.Description(), tt.wantDescription)
			}
		})
	}
}

func TestSecured(t *testing.T) {
	tests := []struct {
		name        string
		route       Route
		wantSecured []string
	}{
		{
			name:        "a single scheme",
			route:       GET("/users", dummyHandler).WithSecured("bearerAuth"),
			wantSecured: []string{"bearerAuth"},
		},
		{
			name:        "multiple schemes",
			route:       GET("/users", dummyHandler).WithSecured("bearerAuth", "apiKey"),
			wantSecured: []string{"bearerAuth", "apiKey"},
		},
		{
			name:        "without secured",
			route:       GET("/users", dummyHandler),
			wantSecured: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.route.Secured()
			if len(got) != len(tt.wantSecured) {
				t.Errorf("Secured() len = %v, want %v", len(got), len(tt.wantSecured))
				return
			}
			for i, s := range got {
				if s != tt.wantSecured[i] {
					t.Errorf("Secured()[%d] = %v, want %v", i, s, tt.wantSecured[i])
				}
			}
		})
	}
}

func TestMiddlewares(t *testing.T) {
	tests := []struct {
		name            string
		route           Route
		wantMiddlewares int
	}{
		{
			name:            "a single middleware",
			route:           GET("/users", dummyHandler).Use(dummyMiddleware()),
			wantMiddlewares: 1,
		},
		{
			name:            "multiple middlewares",
			route:           GET("/users", dummyHandler).Use(dummyMiddleware(), dummyMiddleware()),
			wantMiddlewares: 2,
		},
		{
			name:            "without middlewares",
			route:           GET("/users", dummyHandler),
			wantMiddlewares: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := len(tt.route.Middlewares())
			if got != tt.wantMiddlewares {
				t.Errorf("Middlewares() len = %v, want %v", got, tt.wantMiddlewares)
			}
		})
	}
}

func TestBuilderCombinations(t *testing.T) {
	tests := []struct {
		name            string
		route           Route
		wantMethod      string
		wantPath        string
		wantSummary     string
		wantTags        []string
		wantSecured     []string
		wantBody        bool
		wantResponse    bool
		wantMiddlewares int
		wantStatusCode  int
	}{
		{
			name: "complete GET route",
			route: GET("/users/:id", dummyHandler).
				Res(WithResponse[testResponseDTO](200)).
				Tag("users").
				Describe("Get user", "Returns a user by ID").
				WithSecured("bearerAuth").
				Use(dummyMiddleware()),
			wantMethod:      "GET",
			wantPath:        "/users/:id",
			wantSummary:     "Get user",
			wantTags:        []string{"users"},
			wantSecured:     []string{"bearerAuth"},
			wantBody:        false,
			wantResponse:    true,
			wantMiddlewares: 1,
			wantStatusCode:  200,
		},
		{
			name: "complete POST route",
			route: POST("/users", dummyHandler).
				WithBody(WithBody[testDTO]()).
				Res(WithResponse[testResponseDTO](201)).
				Tag("users").
				Tag("admin").
				Describe("Create user").
				WithSecured("bearerAuth", "apiKey").
				Use(dummyMiddleware(), dummyMiddleware()),
			wantMethod:      "POST",
			wantPath:        "/users",
			wantSummary:     "Create user",
			wantTags:        []string{"users", "admin"},
			wantSecured:     []string{"bearerAuth", "apiKey"},
			wantBody:        true,
			wantResponse:    true,
			wantMiddlewares: 2,
			wantStatusCode:  201,
		},
		{
			name: "inline function",
			route: GET("/ping", func(ctx *Ctx) error {
				return nil
			}).Tag("health").Describe("Ping"),
			wantMethod:  "GET",
			wantPath:    "/ping",
			wantSummary: "Ping",
			wantTags:    []string{"health"},
		},
		{
			name: "complete chained builder with DELETE",
			route: DELETE("/users/:id", dummyHandler).
				Use(dummyMiddleware(), dummyMiddleware()).
				WithSecured("bearerAuth").
				Tag("users").
				Tag("admin").
				Tag("backoffice").
				Describe("Delete user", "Deletes a user by ID"),
			wantMethod:      "DELETE",
			wantPath:        "/users/:id",
			wantSummary:     "Delete user",
			wantTags:        []string{"users", "admin", "backoffice"},
			wantSecured:     []string{"bearerAuth"},
			wantBody:        false,
			wantResponse:    false,
			wantMiddlewares: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.route.Method() != tt.wantMethod {
				t.Errorf("Method() = %v, want %v", tt.route.Method(), tt.wantMethod)
			}
			if tt.route.Path() != tt.wantPath {
				t.Errorf("Path() = %v, want %v", tt.route.Path(), tt.wantPath)
			}
			if tt.route.Summary() != tt.wantSummary {
				t.Errorf("Summary() = %v, want %v", tt.route.Summary(), tt.wantSummary)
			}
			if tt.wantBody && tt.route.Body() == nil {
				t.Error("Body() should not be nil")
			}
			if !tt.wantBody && tt.route.Body() != nil {
				t.Error("Body() should be nil")
			}
			if tt.wantResponse && tt.route.Response() == nil {
				t.Error("Response() should not be nil")
			}
			if !tt.wantResponse && tt.route.Response() != nil {
				t.Error("Response() should be nil")
			}
			if tt.wantResponse && tt.route.Response() != nil && tt.wantStatusCode != 0 {
				if tt.route.Response().StatusCode != tt.wantStatusCode {
					t.Errorf("StatusCode = %v, want %v", tt.route.Response().StatusCode, tt.wantStatusCode)
				}
			}
			if len(tt.route.Tags()) != len(tt.wantTags) {
				t.Errorf("Tags() len = %v, want %v", len(tt.route.Tags()), len(tt.wantTags))
			}
			if len(tt.route.Secured()) != len(tt.wantSecured) {
				t.Errorf("Secured() len = %v, want %v", len(tt.route.Secured()), len(tt.wantSecured))
			}
			if len(tt.route.Middlewares()) != tt.wantMiddlewares {
				t.Errorf("Middlewares() len = %v, want %v", len(tt.route.Middlewares()), tt.wantMiddlewares)
			}
		})
	}
}
