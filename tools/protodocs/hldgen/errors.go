package hldgen

import (
	"errors"
	"fmt"
)

// Common errors
var (
	ErrInvalidConfig        = errors.New("invalid configuration")
	ErrInvalidInput         = errors.New("invalid input")
	ErrNoAgents             = errors.New("no agents configured")
	ErrConsensusNotReached  = errors.New("consensus threshold not reached")
	ErrMaxRoundsExceeded    = errors.New("maximum refinement rounds exceeded")
	ErrLLMUnavailable       = errors.New("LLM service unavailable")
	ErrContextEnrichmentFailed = errors.New("context enrichment failed")
	ErrCriticalIssuesFound  = errors.New("critical issues found in output")
)

// HLDError represents a structured error with context
type HLDError struct {
	Code    string
	Message string
	Cause   error
	Context map[string]interface{}
}

// Error implements error interface
func (e *HLDError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *HLDError) Unwrap() error {
	return e.Cause
}

// NewHLDError creates a new HLD error
func NewHLDError(code, message string, cause error) *HLDError {
	return &HLDError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// WithContext adds context to the error
func (e *HLDError) WithContext(key string, value interface{}) *HLDError {
	e.Context[key] = value
	return e
}

// Error codes
const (
	ErrCodeInvalidConfig    = "INVALID_CONFIG"
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeNoAgents         = "NO_AGENTS"
	ErrCodeConsensusFailure = "CONSENSUS_FAILURE"
	ErrCodeMaxRounds        = "MAX_ROUNDS"
	ErrCodeLLMFailure       = "LLM_FAILURE"
	ErrCodeContextFailure   = "CONTEXT_FAILURE"
	ErrCodeCriticalIssues   = "CRITICAL_ISSUES"
	ErrCodeValidationFailed = "VALIDATION_FAILED"
)

// ErrorRecovery provides error recovery strategies
type ErrorRecovery struct {
	maxRetries int
	backoff    func(int) int
}

// NewErrorRecovery creates a new error recovery handler
func NewErrorRecovery() *ErrorRecovery {
	return &ErrorRecovery{
		maxRetries: 3,
		backoff: func(attempt int) int {
			// Exponential backoff: 1s, 2s, 4s
			return 1 << attempt
		},
	}
}

// Retry executes a function with retries
func (er *ErrorRecovery) Retry(fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < er.maxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on validation errors
		var valErr *ValidationErrors
		if errors.As(err, &valErr) {
			return err
		}

		// Backoff before retry
		if attempt < er.maxRetries-1 {
			// In production, use time.Sleep
			// time.Sleep(time.Duration(er.backoff(attempt)) * time.Second)
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// WrapError wraps an error with additional context
func WrapError(err error, code, message string) error {
	if err == nil {
		return nil
	}
	return NewHLDError(code, message, err)
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Don't retry validation errors
	var valErr *ValidationErrors
	if errors.As(err, &valErr) {
		return false
	}

	// Don't retry invalid config
	if errors.Is(err, ErrInvalidConfig) {
		return false
	}

	// Don't retry invalid input
	if errors.Is(err, ErrInvalidInput) {
		return false
	}

	// Retry LLM errors
	if errors.Is(err, ErrLLMUnavailable) {
		return true
	}

	// Retry context enrichment errors
	if errors.Is(err, ErrContextEnrichmentFailed) {
		return true
	}

	return false
}

// ErrorReporter reports errors for monitoring
type ErrorReporter interface {
	ReportError(err error, context map[string]interface{})
}

// NoOpErrorReporter is a no-op error reporter
type NoOpErrorReporter struct{}

// ReportError does nothing
func (n *NoOpErrorReporter) ReportError(err error, context map[string]interface{}) {
	// No-op
}
