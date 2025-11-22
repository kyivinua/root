package errors

import (
	"errors"
	"fmt"
)

// ErrorType represents the category of error.
type ErrorType string

const (
	// ErrorTypeValidation represents input validation errors
	ErrorTypeValidation ErrorType = "validation"

	// ErrorTypeNotFound represents resource not found errors
	ErrorTypeNotFound ErrorType = "not_found"

	// ErrorTypePermission represents permission/authorization errors
	ErrorTypePermission ErrorType = "permission"

	// ErrorTypeNetwork represents network-related errors
	ErrorTypeNetwork ErrorType = "network"

	// ErrorTypeAPI represents external API errors
	ErrorTypeAPI ErrorType = "api"

	// ErrorTypeConfig represents configuration errors
	ErrorTypeConfig ErrorType = "config"

	// ErrorTypeInternal represents internal/unexpected errors
	ErrorTypeInternal ErrorType = "internal"

	// ErrorTypeTimeout represents timeout errors
	ErrorTypeTimeout ErrorType = "timeout"

	// ErrorTypeRetryable represents errors that can be retried
	ErrorTypeRetryable ErrorType = "retryable"
)

// Error is a structured error with additional context.
type Error struct {
	Type    ErrorType
	Message string
	Cause   error
	Context map[string]interface{}
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the wrapped error.
func (e *Error) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the target.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Type == t.Type
}

// New creates a new structured error.
func New(errType ErrorType, message string) *Error {
	return &Error{
		Type:    errType,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with additional context.
func Wrap(err error, errType ErrorType, message string) *Error {
	return &Error{
		Type:    errType,
		Message: message,
		Cause:   err,
		Context: make(map[string]interface{}),
	}
}

// WithContext adds context to the error.
func (e *Error) WithContext(key string, value interface{}) *Error {
	e.Context[key] = value
	return e
}

// IsType checks if an error is of a specific type.
func IsType(err error, errType ErrorType) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Type == errType
	}
	return false
}

// IsRetryable checks if an error can be retried.
func IsRetryable(err error) bool {
	return IsType(err, ErrorTypeRetryable) ||
		IsType(err, ErrorTypeNetwork) ||
		IsType(err, ErrorTypeTimeout)
}

// Common validation errors
var (
	ErrInvalidPath      = New(ErrorTypeValidation, "invalid path")
	ErrInvalidAPIKey    = New(ErrorTypeValidation, "invalid API key")
	ErrInvalidURL       = New(ErrorTypeValidation, "invalid URL")
	ErrEmptyValue       = New(ErrorTypeValidation, "value cannot be empty")
	ErrInvalidFormat    = New(ErrorTypeValidation, "invalid format")
)

// Common not found errors
var (
	ErrFileNotFound      = New(ErrorTypeNotFound, "file not found")
	ErrDirectoryNotFound = New(ErrorTypeNotFound, "directory not found")
	ErrResourceNotFound  = New(ErrorTypeNotFound, "resource not found")
)

// Common configuration errors
var (
	ErrMissingConfig = New(ErrorTypeConfig, "missing required configuration")
	ErrInvalidConfig = New(ErrorTypeConfig, "invalid configuration")
)

// Common API errors
var (
	ErrAPIUnavailable = New(ErrorTypeAPI, "API unavailable")
	ErrAPIRateLimit   = New(ErrorTypeRetryable, "API rate limit exceeded")
	ErrAPIAuth        = New(ErrorTypePermission, "API authentication failed")
)

// ValidationError creates a validation error.
func ValidationError(message string) *Error {
	return New(ErrorTypeValidation, message)
}

// NotFoundError creates a not found error.
func NotFoundError(message string) *Error {
	return New(ErrorTypeNotFound, message)
}

// PermissionError creates a permission error.
func PermissionError(message string) *Error {
	return New(ErrorTypePermission, message)
}

// NetworkError creates a network error.
func NetworkError(message string) *Error {
	return New(ErrorTypeNetwork, message)
}

// APIError creates an API error.
func APIError(message string) *Error {
	return New(ErrorTypeAPI, message)
}

// ConfigError creates a configuration error.
func ConfigError(message string) *Error {
	return New(ErrorTypeConfig, message)
}

// InternalError creates an internal error.
func InternalError(message string) *Error {
	return New(ErrorTypeInternal, message)
}

// TimeoutError creates a timeout error.
func TimeoutError(message string) *Error {
	return New(ErrorTypeTimeout, message)
}

// RetryableError creates a retryable error.
func RetryableError(message string) *Error {
	return New(ErrorTypeRetryable, message)
}
