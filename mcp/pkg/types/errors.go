package types

import "errors"

// Request validation errors
var (
	ErrInvalidJSONRPC        = errors.New("jsonrpc must be '2.0'")
	ErrMissingMethod         = errors.New("method is required")
	ErrInvalidFieldSelection = errors.New("cannot specify both include and exclude")
	ErrInvalidMaxDepth       = errors.New("maxDepth must be between 1 and 20")
	ErrInvalidStrategy       = errors.New("invalid filtering strategy")
)

// Response validation errors
var (
	ErrBothResultAndError    = errors.New("response cannot have both result and error")
	ErrNeitherResultNorError = errors.New("response must have either result or error")
)

// Filtering errors
var (
	ErrFilteringFailed  = errors.New("filtering operation failed")
	ErrInvalidFieldName = errors.New("invalid field name")
	ErrMaxDepthExceeded = errors.New("maximum nesting depth exceeded")
	ErrTokenCountFailed = errors.New("token counting failed")
)

// Configuration errors
var (
	ErrInvalidConfig        = errors.New("invalid configuration")
	ErrMissingServerURL     = errors.New("server URL is required")
	ErrInvalidServerURL     = errors.New("invalid server URL")
	ErrInvalidTimeout       = errors.New("invalid timeout value")
	ErrInvalidPort          = errors.New("invalid port number")
	ErrInvalidListenAddress = errors.New("invalid listen address")
)

// Schema errors
var (
	ErrSchemaNotFound        = errors.New("schema not found")
	ErrSchemaInvalid         = errors.New("schema is invalid")
	ErrSchemaVersionMismatch = errors.New("schema version mismatch")
	ErrRequiredFieldMissing  = errors.New("required field is missing")
)

// Downstream errors
var (
	ErrDownstreamTimeout     = errors.New("downstream server timeout")
	ErrDownstreamUnavailable = errors.New("downstream server unavailable")
	ErrDownstreamError       = errors.New("downstream server error")
	ErrCircuitBreakerOpen    = errors.New("circuit breaker open")
)

// FilterError represents a filtering error with context
type FilterError struct {
	Field   string
	Message string
	Err     error
}

func (e *FilterError) Error() string {
	if e.Field != "" {
		return "filtering error on field '" + e.Field + "': " + e.Message
	}
	return "filtering error: " + e.Message
}

func (e *FilterError) Unwrap() error {
	return e.Err
}

// NewFilterError creates a new filter error
func NewFilterError(field, message string, err error) *FilterError {
	return &FilterError{
		Field:   field,
		Message: message,
		Err:     err,
	}
}

// ValidationError represents a validation error with context
type ValidationError struct {
	Field   string
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return "validation error on field '" + e.Field + "': " + e.Message
	}
	return "validation error: " + e.Message
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string, err error) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Err:     err,
	}
}

// TimeoutError represents a timeout error with context
type TimeoutError struct {
	Operation string
	Duration  string
	Err       error
}

func (e *TimeoutError) Error() string {
	return "timeout during " + e.Operation + " after " + e.Duration
}

func (e *TimeoutError) Unwrap() error {
	return e.Err
}

// NewTimeoutError creates a new timeout error
func NewTimeoutError(operation, duration string, err error) *TimeoutError {
	return &TimeoutError{
		Operation: operation,
		Duration:  duration,
		Err:       err,
	}
}
