package confluence

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/kyivinua/docgen-tool/tools/protodocs/pkg/errors"
)

// RetryConfig holds configuration for retry logic
type RetryConfig struct {
	MaxAttempts     int
	InitialBackoff  time.Duration
	MaxBackoff      time.Duration
	BackoffMultiplier float64
	RetryableErrors []errors.ErrorType
}

// DefaultRetryConfig returns sensible defaults for retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:      3,
		InitialBackoff:   time.Second,
		MaxBackoff:       30 * time.Second,
		BackoffMultiplier: 2.0,
		RetryableErrors: []errors.ErrorType{
			errors.ErrorTypeRetryable,
			errors.ErrorTypeNetwork,
		},
	}
}

// RetryableFunc is a function that can be retried
type RetryableFunc func(ctx context.Context) error

// WithRetry executes a function with exponential backoff retry logic
func WithRetry(ctx context.Context, config *RetryConfig, fn RetryableFunc) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr error
	backoff := config.InitialBackoff

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// Check context cancellation before attempting
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute the function
		err := fn(ctx)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err, config.RetryableErrors) {
			return err
		}

		// Don't sleep after last attempt
		if attempt == config.MaxAttempts {
			break
		}

		// Calculate backoff with jitter
		sleepDuration := calculateBackoff(backoff, config.MaxBackoff)

		// Log retry attempt
		fmt.Printf("[confluence] Attempt %d/%d failed: %v. Retrying in %v...\n",
			attempt, config.MaxAttempts, err, sleepDuration)

		// Wait with context cancellation support
		select {
		case <-time.After(sleepDuration):
			// Continue to next attempt
		case <-ctx.Done():
			return ctx.Err()
		}

		// Increase backoff for next attempt
		backoff = time.Duration(float64(backoff) * config.BackoffMultiplier)
	}

	return fmt.Errorf("max retry attempts (%d) exceeded: %w", config.MaxAttempts, lastErr)
}

// isRetryableError checks if an error should be retried
func isRetryableError(err error, retryableTypes []errors.ErrorType) bool {
	if err == nil {
		return false
	}

	// Check if it's our custom error type
	if appErr, ok := err.(*errors.Error); ok {
		for _, errType := range retryableTypes {
			if appErr.Type == errType {
				return true
			}
		}
	}

	return false
}

// calculateBackoff calculates the backoff duration with jitter
func calculateBackoff(current, max time.Duration) time.Duration {
	// Add jitter (±20%)
	jitter := float64(current) * 0.2 * (2*float64(time.Now().UnixNano()%100)/100 - 1)
	backoff := time.Duration(float64(current) + jitter)

	// Cap at max backoff
	if backoff > max {
		backoff = max
	}

	// Ensure minimum backoff of 100ms
	if backoff < 100*time.Millisecond {
		backoff = 100 * time.Millisecond
	}

	return backoff
}

// WithExponentialBackoff is a helper that retries with default exponential backoff
func WithExponentialBackoff(ctx context.Context, fn RetryableFunc) error {
	return WithRetry(ctx, DefaultRetryConfig(), fn)
}

// WithCustomRetry allows custom retry configuration
func WithCustomRetry(ctx context.Context, maxAttempts int, initialBackoff time.Duration, fn RetryableFunc) error {
	config := &RetryConfig{
		MaxAttempts:      maxAttempts,
		InitialBackoff:   initialBackoff,
		MaxBackoff:       30 * time.Second,
		BackoffMultiplier: 2.0,
		RetryableErrors: []errors.ErrorType{
			errors.ErrorTypeRetryable,
			errors.ErrorTypeNetwork,
		},
	}
	return WithRetry(ctx, config, fn)
}

// RetryStats tracks retry statistics for monitoring
type RetryStats struct {
	TotalAttempts   int
	SuccessfulRetries int
	FailedRetries   int
	TotalBackoffTime time.Duration
}

// RetryWithStats executes a function with retry and returns statistics
func RetryWithStats(ctx context.Context, config *RetryConfig, fn RetryableFunc) (*RetryStats, error) {
	if config == nil {
		config = DefaultRetryConfig()
	}

	stats := &RetryStats{}
	var lastErr error
	backoff := config.InitialBackoff

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		stats.TotalAttempts++

		select {
		case <-ctx.Done():
			return stats, ctx.Err()
		default:
		}

		err := fn(ctx)
		if err == nil {
			if attempt > 1 {
				stats.SuccessfulRetries++
			}
			return stats, nil
		}

		lastErr = err

		if !isRetryableError(err, config.RetryableErrors) {
			return stats, err
		}

		if attempt == config.MaxAttempts {
			stats.FailedRetries++
			break
		}

		sleepDuration := calculateBackoff(backoff, config.MaxBackoff)
		stats.TotalBackoffTime += sleepDuration

		select {
		case <-time.After(sleepDuration):
		case <-ctx.Done():
			return stats, ctx.Err()
		}

		backoff = time.Duration(float64(backoff) * config.BackoffMultiplier)
	}

	return stats, fmt.Errorf("max retry attempts (%d) exceeded: %w", config.MaxAttempts, lastErr)
}

// calculateJitteredBackoff calculates backoff with full jitter (0-100%)
func calculateJitteredBackoff(attempt int, base time.Duration, max time.Duration) time.Duration {
	// Exponential: base * 2^attempt
	exponential := float64(base) * math.Pow(2, float64(attempt-1))

	// Cap at max
	if exponential > float64(max) {
		exponential = float64(max)
	}

	// Full jitter: random value between 0 and exponential
	jittered := float64(time.Now().UnixNano()%int64(exponential))

	backoff := time.Duration(jittered)
	if backoff < 100*time.Millisecond {
		backoff = 100 * time.Millisecond
	}

	return backoff
}
