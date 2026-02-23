package validation

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// FieldError represents a validation error on a specific field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Validate validates a struct with `validate` tags.
// Returns nil if there are no errors.
func Validate(s any) []FieldError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}
	var errs []FieldError
	for _, e := range err.(validator.ValidationErrors) {
		errs = append(errs, FieldError{
			Field:   e.Field(),
			Message: humanMessage(e),
		})
	}
	return errs
}

func humanMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email"
	case "min":
		return fmt.Sprintf("minimum %s characters", e.Param())
	case "max":
		return fmt.Sprintf("maximum %s characters", e.Param())
	case "uuid", "uuid4":
		return "must be a valid UUID"
	case "numeric":
		return "must be a numeric value"
	case "url":
		return "must be a valid URL"
	default:
		return fmt.Sprintf("validation failed: %s", e.Tag())
	}
}
