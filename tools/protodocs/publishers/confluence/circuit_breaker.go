package confluence

import (
	"context"
	"errors"
	"sync"
	"time"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	// StateClosed means requests are allowed through
	StateClosed CircuitBreakerState = iota
	// StateHalfOpen means limited requests are allowed to test if service recovered
	StateHalfOpen
	// StateOpen means requests are blocked
	StateOpen
)

// String returns string representation of circuit breaker state
func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateHalfOpen:
		return "half-open"
	case StateOpen:
		return "open"
	default:
		return "unknown"
	}
}

var (
	// ErrCircuitBreakerOpen is returned when circuit breaker is open
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
	// ErrTooManyRequests is returned when too many requests are made in half-open state
	ErrTooManyRequests = errors.New("too many requests")
)

// CircuitBreakerConfig holds configuration for circuit breaker
type CircuitBreakerConfig struct {
	// Name of the circuit breaker (for metrics/logging)
	Name string

	// MaxFailures is the number of consecutive failures before opening
	MaxFailures uint32

	// Timeout is how long to wait before transitioning from open to half-open
	Timeout time.Duration

	// MaxHalfOpenRequests is max concurrent requests allowed in half-open state
	MaxHalfOpenRequests uint32

	// FailureRatio is the ratio of failures (0.0-1.0) that triggers opening
	// If set, this is used instead of MaxFailures
	FailureRatio float64

	// MinRequests is minimum requests before failure ratio is considered
	MinRequests uint32

	// ShouldTrip is a custom function to determine if breaker should trip
	// If nil, default logic (MaxFailures or FailureRatio) is used
	ShouldTrip func(counts Counts) bool

	// OnStateChange is called when state changes
	OnStateChange func(name string, from, to CircuitBreakerState)

	// Metrics instance for recording metrics
	Metrics *Metrics
}

// DefaultCircuitBreakerConfig returns default configuration
func DefaultCircuitBreakerConfig(name string) *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		Name:                name,
		MaxFailures:         5,
		Timeout:             60 * time.Second,
		MaxHalfOpenRequests: 1,
		MinRequests:         10,
		FailureRatio:        0.5,
		Metrics:             DefaultMetrics,
	}
}

// Counts holds the number of requests and their successes/failures
type Counts struct {
	Requests             uint32
	TotalSuccesses       uint32
	TotalFailures        uint32
	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
}

// CircuitBreaker prevents cascading failures by stopping requests to failing services
type CircuitBreaker struct {
	name                string
	maxFailures         uint32
	timeout             time.Duration
	maxHalfOpenRequests uint32
	failureRatio        float64
	minRequests         uint32
	shouldTrip          func(counts Counts) bool
	onStateChange       func(name string, from, to CircuitBreakerState)
	metrics             *Metrics

	mu              sync.RWMutex
	state           CircuitBreakerState
	generation      uint64
	counts          Counts
	expiry          time.Time
	halfOpenAllowed uint32
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig("default")
	}

	cb := &CircuitBreaker{
		name:                config.Name,
		maxFailures:         config.MaxFailures,
		timeout:             config.Timeout,
		maxHalfOpenRequests: config.MaxHalfOpenRequests,
		failureRatio:        config.FailureRatio,
		minRequests:         config.MinRequests,
		shouldTrip:          config.ShouldTrip,
		onStateChange:       config.OnStateChange,
		metrics:             config.Metrics,
		state:               StateClosed,
	}

	// Use default trip logic if not provided
	if cb.shouldTrip == nil {
		cb.shouldTrip = cb.defaultShouldTrip
	}

	// Update initial metrics
	if cb.metrics != nil {
		cb.metrics.UpdateCircuitBreakerState(cb.name, int(StateClosed))
	}

	return cb
}

// Execute executes the given function if circuit breaker allows it
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(context.Context) error) error {
	// Check if we can proceed
	generation, err := cb.beforeRequest()
	if err != nil {
		return err
	}

	// Execute the function
	err = fn(ctx)

	// Record the result
	cb.afterRequest(generation, err)

	return err
}

// beforeRequest checks if request is allowed and updates state if needed
func (cb *CircuitBreaker) beforeRequest() (uint64, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state := cb.currentState(now)

	if state == StateOpen {
		return 0, ErrCircuitBreakerOpen
	}

	if state == StateHalfOpen {
		// Check if we can allow another request in half-open state
		if cb.halfOpenAllowed >= cb.maxHalfOpenRequests {
			return 0, ErrTooManyRequests
		}
		cb.halfOpenAllowed++
	}

	cb.counts.Requests++
	return cb.generation, nil
}

// afterRequest records the result and potentially changes state
func (cb *CircuitBreaker) afterRequest(generation uint64, err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state := cb.currentState(now)

	// Ignore if generation doesn't match (state has changed)
	if generation != cb.generation {
		return
	}

	if err == nil {
		cb.onSuccess(state, now)
	} else {
		cb.onFailure(state, now)
	}
}

// onSuccess handles successful request
func (cb *CircuitBreaker) onSuccess(state CircuitBreakerState, now time.Time) {
	cb.counts.TotalSuccesses++
	cb.counts.ConsecutiveSuccesses++
	cb.counts.ConsecutiveFailures = 0

	if state == StateHalfOpen {
		// If successful in half-open, consider closing
		if cb.counts.ConsecutiveSuccesses >= cb.maxHalfOpenRequests {
			cb.setState(StateClosed, now)
		}
	}
}

// onFailure handles failed request
func (cb *CircuitBreaker) onFailure(state CircuitBreakerState, now time.Time) {
	cb.counts.TotalFailures++
	cb.counts.ConsecutiveFailures++
	cb.counts.ConsecutiveSuccesses = 0

	if cb.shouldTrip(cb.counts) {
		cb.setState(StateOpen, now)
	}
}

// currentState returns current state, transitioning if needed
func (cb *CircuitBreaker) currentState(now time.Time) CircuitBreakerState {
	switch cb.state {
	case StateClosed:
		return StateClosed
	case StateOpen:
		// Check if we should transition to half-open
		if cb.expiry.Before(now) {
			cb.setState(StateHalfOpen, now)
			return StateHalfOpen
		}
		return StateOpen
	default:
		return cb.state
	}
}

// setState changes the circuit breaker state
func (cb *CircuitBreaker) setState(state CircuitBreakerState, now time.Time) {
	if cb.state == state {
		return
	}

	prev := cb.state
	cb.state = state
	cb.generation++
	cb.counts = Counts{}
	cb.halfOpenAllowed = 0

	switch state {
	case StateClosed:
		// No expiry in closed state
	case StateOpen:
		cb.expiry = now.Add(cb.timeout)
		if cb.metrics != nil {
			cb.metrics.RecordCircuitBreakerTrip(cb.name)
		}
	case StateHalfOpen:
		// No expiry in half-open, will close or open based on requests
	}

	// Notify state change
	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, prev, state)
	}

	// Update metrics
	if cb.metrics != nil {
		cb.metrics.UpdateCircuitBreakerState(cb.name, int(state))
	}
}

// defaultShouldTrip is the default logic to determine if breaker should trip
func (cb *CircuitBreaker) defaultShouldTrip(counts Counts) bool {
	// Use failure ratio if configured and we have minimum requests
	if cb.failureRatio > 0 && counts.Requests >= cb.minRequests {
		ratio := float64(counts.TotalFailures) / float64(counts.Requests)
		return ratio >= cb.failureRatio
	}

	// Otherwise use consecutive failures
	return counts.ConsecutiveFailures >= cb.maxFailures
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.currentState(time.Now())
}

// Counts returns a copy of current counts
func (cb *CircuitBreaker) Counts() Counts {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.counts
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.setState(StateClosed, time.Now())
}

// WithCircuitBreaker wraps a client with circuit breaker protection
type WithCircuitBreaker struct {
	client  *Client
	breaker *CircuitBreaker
}

// NewClientWithCircuitBreaker wraps a client with circuit breaker
func NewClientWithCircuitBreaker(client *Client, config *CircuitBreakerConfig) *WithCircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig("confluence_api")
	}

	return &WithCircuitBreaker{
		client:  client,
		breaker: NewCircuitBreaker(config),
	}
}

// Execute executes a function with circuit breaker protection
func (w *WithCircuitBreaker) Execute(ctx context.Context, fn func(*Client) error) error {
	return w.breaker.Execute(ctx, func(ctx context.Context) error {
		return fn(w.client)
	})
}

// GetClient returns the underlying client
func (w *WithCircuitBreaker) GetClient() *Client {
	return w.client
}

// GetBreaker returns the circuit breaker
func (w *WithCircuitBreaker) GetBreaker() *CircuitBreaker {
	return w.breaker
}
