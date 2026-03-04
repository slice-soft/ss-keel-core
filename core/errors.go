package core

import "fmt"

// KError is the standard error type that the App error handler maps to HTTP responses.
// All modules should return *KError so the handler can set the correct status code.
type KError struct {
	Code       string
	StatusCode int
	Message    string
	Cause      error
}

func (e *KError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Cause.Error())
	}
	return e.Message
}

func (e *KError) Unwrap() error { return e.Cause }

// NotFound creates a 404 KError.
func NotFound(msg string) *KError {
	return &KError{Code: "NOT_FOUND", StatusCode: 404, Message: msg}
}

// Unauthorized creates a 401 KError.
func Unauthorized(msg string) *KError {
	return &KError{Code: "UNAUTHORIZED", StatusCode: 401, Message: msg}
}

// Forbidden creates a 403 KError.
func Forbidden(msg string) *KError {
	return &KError{Code: "FORBIDDEN", StatusCode: 403, Message: msg}
}

// Conflict creates a 409 KError.
func Conflict(msg string) *KError {
	return &KError{Code: "CONFLICT", StatusCode: 409, Message: msg}
}

// BadRequest creates a 400 KError.
func BadRequest(msg string) *KError {
	return &KError{Code: "BAD_REQUEST", StatusCode: 400, Message: msg}
}

// Internal creates a 500 KError with an optional cause.
func Internal(msg string, cause error) *KError {
	return &KError{Code: "INTERNAL_ERROR", StatusCode: 500, Message: msg, Cause: cause}
}
