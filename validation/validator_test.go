package validation

import (
	"testing"
)

func TestValidate(t *testing.T) {
	type loginDTO struct {
		Email    string `validate:"required,email"`
		Password string `validate:"required,min=8"`
	}

	type profileDTO struct {
		Name    string `validate:"required,min=2,max=50"`
		Website string `validate:"omitempty,url"`
		UserID  string `validate:"required,uuid4"`
	}

	tests := []struct {
		name      string
		input     any
		wantNil   bool
		wantCount int
		wantField string
	}{
		{
			name:    "valid struct returns nil",
			input:   loginDTO{Email: "juan@test.com", Password: "secret123"},
			wantNil: true,
		},
		{
			name:      "missing required fields",
			input:     loginDTO{},
			wantNil:   false,
			wantCount: 2,
		},
		{
			name:      "invalid email",
			input:     loginDTO{Email: "notanemail", Password: "secret123"},
			wantNil:   false,
			wantCount: 1,
			wantField: "Email",
		},
		{
			name:      "password too short",
			input:     loginDTO{Email: "juan@test.com", Password: "short"},
			wantNil:   false,
			wantCount: 1,
			wantField: "Password",
		},
		{
			name:    "valid profile",
			input:   profileDTO{Name: "Juan", UserID: "550e8400-e29b-41d4-a716-446655440000"},
			wantNil: true,
		},
		{
			name:      "invalid uuid",
			input:     profileDTO{Name: "Juan", UserID: "not-a-uuid"},
			wantNil:   false,
			wantCount: 1,
			wantField: "UserID",
		},
		{
			name:      "name too short",
			input:     profileDTO{Name: "J", UserID: "550e8400-e29b-41d4-a716-446655440000"},
			wantNil:   false,
			wantCount: 1,
			wantField: "Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := Validate(tt.input)

			if tt.wantNil && errs != nil {
				t.Errorf("expected nil errors, got %v", errs)
				return
			}
			if !tt.wantNil && errs == nil {
				t.Error("expected errors but got nil")
				return
			}
			if tt.wantCount > 0 && len(errs) != tt.wantCount {
				t.Errorf("error count = %v, want %v", len(errs), tt.wantCount)
			}
			if tt.wantField != "" {
				found := false
				for _, e := range errs {
					if e.Field == tt.wantField {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error on field %q", tt.wantField)
				}
			}
		})
	}
}

func TestHumanMessage(t *testing.T) {
	type requiredDTO struct {
		Name string `validate:"required"`
	}
	type emailDTO struct {
		Email string `validate:"required,email"`
	}
	type minDTO struct {
		Name string `validate:"required,min=8"`
	}
	type maxDTO struct {
		Name string `validate:"required,max=3"`
	}
	type uuidDTO struct {
		ID string `validate:"required,uuid4"`
	}
	type numericDTO struct {
		Code string `validate:"required,numeric"`
	}
	type urlDTO struct {
		Link string `validate:"required,url"`
	}

	tests := []struct {
		name        string
		input       any
		wantMessage string
		wantField   string
	}{
		{
			name:        "required message",
			input:       requiredDTO{},
			wantField:   "Name",
			wantMessage: "this field is required",
		},
		{
			name:        "email message",
			input:       emailDTO{Email: "notanemail"},
			wantField:   "Email",
			wantMessage: "must be a valid email",
		},
		{
			name:        "min message",
			input:       minDTO{Name: "short"},
			wantField:   "Name",
			wantMessage: "minimum 8 characters",
		},
		{
			name:        "max message",
			input:       maxDTO{Name: "toolong"},
			wantField:   "Name",
			wantMessage: "maximum 3 characters",
		},
		{
			name:        "uuid message",
			input:       uuidDTO{ID: "not-a-uuid"},
			wantField:   "ID",
			wantMessage: "must be a valid UUID",
		},
		{
			name:        "numeric message",
			input:       numericDTO{Code: "abc"},
			wantField:   "Code",
			wantMessage: "must be a numeric value",
		},
		{
			name:        "url message",
			input:       urlDTO{Link: "not-a-url"},
			wantField:   "Link",
			wantMessage: "must be a valid URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := Validate(tt.input)
			if len(errs) == 0 {
				t.Fatal("expected validation errors but got none")
			}

			var found *FieldError
			for _, e := range errs {
				if e.Field == tt.wantField {
					found = &e
					break
				}
			}
			if found == nil {
				t.Fatalf("expected error on field %q", tt.wantField)
			}
			if found.Message != tt.wantMessage {
				t.Errorf("message = %q, want %q", found.Message, tt.wantMessage)
			}
		})
	}
}
