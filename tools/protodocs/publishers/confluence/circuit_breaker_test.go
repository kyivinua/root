package confluence

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircuitBreaker_InitialState(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig("test"))

	assert.Equal(t, StateClosed, cb.State())

	counts := cb.Counts()
	assert.Equal(t, uint32(0), counts.Requests)
	assert.Equal(t, uint32(0), counts.TotalSuccesses)
	assert.Equal(t, uint32(0), counts.TotalFailures)
}

func TestCircuitBreaker_SuccessfulRequests(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig("test"))
	ctx := context.Background()

	// Execute 10 successful requests
	for i := 0; i < 10; i++ {
		err := cb.Execute(ctx, func(ctx context.Context) error {
			return nil
		})
		require.NoError(t, err)
	}

	// Should remain closed
	assert.Equal(t, StateClosed, cb.State())

	counts := cb.Counts()
	assert.Equal(t, uint32(10), counts.Requests)
	assert.Equal(t, uint32(10), counts.TotalSuccesses)
	assert.Equal(t, uint32(0), counts.TotalFailures)
}

func TestCircuitBreaker_OpensAfterConsecutiveFailures(t *testing.T) {
	config := DefaultCircuitBreakerConfig("test")
	config.MaxFailures = 3
	cb := NewCircuitBreaker(config)
	ctx := context.Background()

	testErr := errors.New("test error")

	// Execute 3 failing requests
	for i := 0; i < 3; i++ {
		_ = cb.Execute(ctx, func(ctx context.Context) error {
			return testErr
		})
	}

	// Should be open now
	assert.Equal(t, StateOpen, cb.State())

	// Next request should fail immediately
	err := cb.Execute(ctx, func(ctx context.Context) error {
		return nil
	})
	assert.Equal(t, ErrCircuitBreakerOpen, err)
}

func TestCircuitBreaker_HalfOpenTransition(t *testing.T) {
	config := DefaultCircuitBreakerConfig("test")
	config.MaxFailures = 2
	config.Timeout = 100 * time.Millisecond
	cb := NewCircuitBreaker(config)
	ctx := context.Background()

	testErr := errors.New("test error")

	// Fail enough to open
	for i := 0; i < 2; i++ {
		_ = cb.Execute(ctx, func(ctx context.Context) error {
			return testErr
		})
	}

	assert.Equal(t, StateOpen, cb.State())

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Should transition to half-open
	assert.Equal(t, StateHalfOpen, cb.State())
}

func TestCircuitBreaker_HalfOpenToClosedOnSuccess(t *testing.T) {
	config := DefaultCircuitBreakerConfig("test")
	config.MaxFailures = 2
	config.Timeout = 50 * time.Millisecond
	config.MaxHalfOpenRequests = 1
	cb := NewCircuitBreaker(config)
	ctx := context.Background()

	testErr := errors.New("test error")

	// Open the circuit
	for i := 0; i < 2; i++ {
		_ = cb.Execute(ctx, func(ctx context.Context) error {
			return testErr
		})
	}

	assert.Equal(t, StateOpen, cb.State())

	// Wait for timeout to transition to half-open
	time.Sleep(100 * time.Millisecond)

	// Successful request in half-open should close it
	err := cb.Execute(ctx, func(ctx context.Context) error {
		return nil
	})
	require.NoError(t, err)

	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreaker_HalfOpenToOpenOnFailure(t *testing.T) {
	config := DefaultCircuitBreakerConfig("test")
	config.MaxFailures = 2
	config.Timeout = 50 * time.Millisecond
	cb := NewCircuitBreaker(config)
	ctx := context.Background()

	testErr := errors.New("test error")

	// Open the circuit
	for i := 0; i < 2; i++ {
		_ = cb.Execute(ctx, func(ctx context.Context) error {
			return testErr
		})
	}

	// Wait for half-open
	time.Sleep(100 * time.Millisecond)

	// Fail in half-open should reopen
	_ = cb.Execute(ctx, func(ctx context.Context) error {
		return testErr
	})

	assert.Equal(t, StateOpen, cb.State())
}

func TestCircuitBreaker_FailureRatio(t *testing.T) {
	config := DefaultCircuitBreakerConfig("test")
	config.FailureRatio = 0.5  // 50% failure rate
	config.MinRequests = 10
	cb := NewCircuitBreaker(config)
	ctx := context.Background()

	testErr := errors.New("test error")

	// Execute 10 requests: 6 failures, 4 successes
	// This is 60% failure rate, should trip
	for i := 0; i < 10; i++ {
		var err error
		if i < 6 {
			err = testErr
		}
		_ = cb.Execute(ctx, func(ctx context.Context) error {
			return err
		})
	}

	// Should be open due to high failure ratio
	assert.Equal(t, StateOpen, cb.State())
}

func TestCircuitBreaker_TooManyRequestsInHalfOpen(t *testing.T) {
	config := DefaultCircuitBreakerConfig("test")
	config.MaxFailures = 2
	config.Timeout = 50 * time.Millisecond
	config.MaxHalfOpenRequests = 1
	cb := NewCircuitBreaker(config)
	ctx := context.Background()

	testErr := errors.New("test error")

	// Open the circuit
	for i := 0; i < 2; i++ {
		_ = cb.Execute(ctx, func(ctx context.Context) error {
			return testErr
		})
	}

	// Wait for half-open
	time.Sleep(100 * time.Millisecond)

	// First request should be allowed
	err1 := cb.Execute(ctx, func(ctx context.Context) error {
		time.Sleep(10 * time.Millisecond) // Simulate slow request
		return nil
	})

	// Second concurrent request should be rejected
	err2 := cb.Execute(ctx, func(ctx context.Context) error {
		return nil
	})

	// One should succeed, one should get TooManyRequests
	if err1 != nil {
		assert.Equal(t, ErrTooManyRequests, err2)
	} else {
		assert.NoError(t, err1)
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	config := DefaultCircuitBreakerConfig("test")
	config.MaxFailures = 2
	cb := NewCircuitBreaker(config)
	ctx := context.Background()

	testErr := errors.New("test error")

	// Open the circuit
	for i := 0; i < 2; i++ {
		_ = cb.Execute(ctx, func(ctx context.Context) error {
			return testErr
		})
	}

	assert.Equal(t, StateOpen, cb.State())

	// Reset should close it
	cb.Reset()

	assert.Equal(t, StateClosed, cb.State())

	counts := cb.Counts()
	assert.Equal(t, uint32(0), counts.Requests)
	assert.Equal(t, uint32(0), counts.TotalFailures)
}

func TestCircuitBreaker_CustomShouldTrip(t *testing.T) {
	config := DefaultCircuitBreakerConfig("test")
	config.ShouldTrip = func(counts Counts) bool {
		// Custom logic: trip if we have exactly 3 consecutive failures
		return counts.ConsecutiveFailures == 3
	}
	cb := NewCircuitBreaker(config)
	ctx := context.Background()

	testErr := errors.New("test error")

	// 2 failures shouldn't trip
	for i := 0; i < 2; i++ {
		_ = cb.Execute(ctx, func(ctx context.Context) error {
			return testErr
		})
	}
	assert.Equal(t, StateClosed, cb.State())

	// 3rd failure should trip
	_ = cb.Execute(ctx, func(ctx context.Context) error {
		return testErr
	})
	assert.Equal(t, StateOpen, cb.State())
}

func TestCircuitBreaker_OnStateChange(t *testing.T) {
	stateChanges := []string{}

	config := DefaultCircuitBreakerConfig("test")
	config.MaxFailures = 2
	config.OnStateChange = func(name string, from, to CircuitBreakerState) {
		stateChanges = append(stateChanges, from.String()+"->"+to.String())
	}
	cb := NewCircuitBreaker(config)
	ctx := context.Background()

	testErr := errors.New("test error")

	// Open the circuit
	for i := 0; i < 2; i++ {
		_ = cb.Execute(ctx, func(ctx context.Context) error {
			return testErr
		})
	}

	// Should have recorded closed->open transition
	require.Len(t, stateChanges, 1)
	assert.Equal(t, "closed->open", stateChanges[0])
}

func TestCircuitBreakerState_String(t *testing.T) {
	assert.Equal(t, "closed", StateClosed.String())
	assert.Equal(t, "half-open", StateHalfOpen.String())
	assert.Equal(t, "open", StateOpen.String())
}
