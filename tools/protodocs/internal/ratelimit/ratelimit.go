package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Limiter is a rate limiter that controls the rate of operations.
type Limiter struct {
	rate     int           // requests per interval
	interval time.Duration // time interval
	tokens   chan struct{} // token bucket
	mu       sync.Mutex
}

// NewLimiter creates a new rate limiter.
// rate is the number of requests allowed per interval.
func NewLimiter(rate int, interval time.Duration) *Limiter {
	if rate <= 0 {
		rate = 1
	}
	if interval <= 0 {
		interval = time.Second
	}

	limiter := &Limiter{
		rate:     rate,
		interval: interval,
		tokens:   make(chan struct{}, rate),
	}

	// Fill the token bucket initially
	for i := 0; i < rate; i++ {
		limiter.tokens <- struct{}{}
	}

	// Start the token refill goroutine
	go limiter.refill()

	return limiter
}

// refill refills tokens at the specified rate.
func (l *Limiter) refill() {
	ticker := time.NewTicker(l.interval / time.Duration(l.rate))
	defer ticker.Stop()

	for range ticker.C {
		select {
		case l.tokens <- struct{}{}:
			// Token added
		default:
			// Bucket is full, skip
		}
	}
}

// Wait blocks until a token is available.
func (l *Limiter) Wait() {
	<-l.tokens
}

// WaitWithContext waits for a token with context support.
func (l *Limiter) WaitWithContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-l.tokens:
		return nil
	}
}

// TryAcquire attempts to acquire a token without blocking.
// Returns true if a token was acquired, false otherwise.
func (l *Limiter) TryAcquire() bool {
	select {
	case <-l.tokens:
		return true
	default:
		return false
	}
}

// AcquireWithTimeout attempts to acquire a token with a timeout.
func (l *Limiter) AcquireWithTimeout(timeout time.Duration) error {
	select {
	case <-l.tokens:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("rate limit timeout after %v", timeout)
	}
}

// MultiLimiter manages multiple rate limiters for different operations.
type MultiLimiter struct {
	limiters map[string]*Limiter
	mu       sync.RWMutex
}

// NewMultiLimiter creates a new multi-limiter.
func NewMultiLimiter() *MultiLimiter {
	return &MultiLimiter{
		limiters: make(map[string]*Limiter),
	}
}

// Add adds a new rate limiter for a specific operation.
func (m *MultiLimiter) Add(name string, rate int, interval time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.limiters[name] = NewLimiter(rate, interval)
}

// Wait waits for a token from the specified limiter.
func (m *MultiLimiter) Wait(name string) error {
	m.mu.RLock()
	limiter, exists := m.limiters[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("limiter %s not found", name)
	}

	limiter.Wait()
	return nil
}

// WaitWithContext waits with context support.
func (m *MultiLimiter) WaitWithContext(ctx context.Context, name string) error {
	m.mu.RLock()
	limiter, exists := m.limiters[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("limiter %s not found", name)
	}

	return limiter.WaitWithContext(ctx)
}

// Adaptive limiter adjusts the rate based on success/failure.
type AdaptiveLimiter struct {
	limiter      *Limiter
	minRate      int
	maxRate      int
	currentRate  int
	interval     time.Duration
	mu           sync.Mutex
	successCount int
	failureCount int
}

// NewAdaptiveLimiter creates an adaptive rate limiter.
func NewAdaptiveLimiter(minRate, maxRate int, interval time.Duration) *AdaptiveLimiter {
	if minRate <= 0 {
		minRate = 1
	}
	if maxRate < minRate {
		maxRate = minRate
	}

	currentRate := (minRate + maxRate) / 2
	return &AdaptiveLimiter{
		limiter:     NewLimiter(currentRate, interval),
		minRate:     minRate,
		maxRate:     maxRate,
		currentRate: currentRate,
		interval:    interval,
	}
}

// Wait waits for a token.
func (a *AdaptiveLimiter) Wait() {
	a.limiter.Wait()
}

// RecordSuccess records a successful operation and may increase the rate.
func (a *AdaptiveLimiter) RecordSuccess() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.successCount++
	a.failureCount = 0

	// Increase rate after 10 consecutive successes
	if a.successCount >= 10 && a.currentRate < a.maxRate {
		a.increaseRate()
	}
}

// RecordFailure records a failed operation and decreases the rate.
func (a *AdaptiveLimiter) RecordFailure() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.failureCount++
	a.successCount = 0

	// Decrease rate after 3 failures
	if a.failureCount >= 3 && a.currentRate > a.minRate {
		a.decreaseRate()
	}
}

// increaseRate increases the rate limit.
func (a *AdaptiveLimiter) increaseRate() {
	newRate := int(float64(a.currentRate) * 1.5)
	if newRate > a.maxRate {
		newRate = a.maxRate
	}

	if newRate != a.currentRate {
		a.currentRate = newRate
		a.limiter = NewLimiter(newRate, a.interval)
		a.successCount = 0
	}
}

// decreaseRate decreases the rate limit.
func (a *AdaptiveLimiter) decreaseRate() {
	newRate := a.currentRate / 2
	if newRate < a.minRate {
		newRate = a.minRate
	}

	if newRate != a.currentRate {
		a.currentRate = newRate
		a.limiter = NewLimiter(newRate, a.interval)
		a.failureCount = 0
	}
}

// GetCurrentRate returns the current rate.
func (a *AdaptiveLimiter) GetCurrentRate() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.currentRate
}
