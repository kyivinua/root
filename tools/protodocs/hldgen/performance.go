package hldgen

import (
	"context"
	"sync"
	"time"
)

// PerformanceMonitor tracks performance metrics
type PerformanceMonitor struct {
	mu      sync.RWMutex
	metrics map[string]*Metric
}

// Metric represents a performance metric
type Metric struct {
	Name     string
	Count    int64
	Total    time.Duration
	Min      time.Duration
	Max      time.Duration
	Average  time.Duration
	LastRun  time.Time
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		metrics: make(map[string]*Metric),
	}
}

// Track tracks execution time of a function
func (pm *PerformanceMonitor) Track(name string, fn func() error) error {
	start := time.Now()
	err := fn()
	duration := time.Since(start)

	pm.Record(name, duration)
	return err
}

// Record records a metric
func (pm *PerformanceMonitor) Record(name string, duration time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	metric, exists := pm.metrics[name]
	if !exists {
		metric = &Metric{
			Name: name,
			Min:  duration,
			Max:  duration,
		}
		pm.metrics[name] = metric
	}

	metric.Count++
	metric.Total += duration
	metric.LastRun = time.Now()

	if duration < metric.Min {
		metric.Min = duration
	}
	if duration > metric.Max {
		metric.Max = duration
	}

	metric.Average = metric.Total / time.Duration(metric.Count)
}

// GetMetrics returns all metrics
func (pm *PerformanceMonitor) GetMetrics() map[string]*Metric {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make(map[string]*Metric, len(pm.metrics))
	for k, v := range pm.metrics {
		result[k] = v
	}
	return result
}

// GetMetric returns a specific metric
func (pm *PerformanceMonitor) GetMetric(name string) *Metric {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return pm.metrics[name]
}

// Reset resets all metrics
func (pm *PerformanceMonitor) Reset() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.metrics = make(map[string]*Metric)
}

// CircuitBreaker implements circuit breaker pattern for LLM calls
type CircuitBreaker struct {
	maxFailures  int
	resetTimeout time.Duration

	mu            sync.RWMutex
	failures      int
	lastFailTime  time.Time
	state         CircuitState
}

// CircuitState represents circuit breaker state
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        CircuitClosed,
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()

	// Check if circuit should be half-open
	if cb.state == CircuitOpen &&
	   time.Since(cb.lastFailTime) > cb.resetTimeout {
		cb.state = CircuitHalfOpen
		cb.failures = 0
	}

	// Reject if circuit is open
	if cb.state == CircuitOpen {
		cb.mu.Unlock()
		return ErrLLMUnavailable
	}

	cb.mu.Unlock()

	// Execute function
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailTime = time.Now()

		if cb.failures >= cb.maxFailures {
			cb.state = CircuitOpen
		}
		return err
	}

	// Success - reset circuit
	if cb.state == CircuitHalfOpen {
		cb.state = CircuitClosed
	}
	cb.failures = 0

	return nil
}

// GetState returns current circuit state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.state
}

// Cache provides simple in-memory caching
type Cache struct {
	mu    sync.RWMutex
	items map[string]*CacheItem
	ttl   time.Duration
}

// CacheItem represents a cached item
type CacheItem struct {
	Value      interface{}
	Expiration time.Time
}

// NewCache creates a new cache
func NewCache(ttl time.Duration) *Cache {
	cache := &Cache{
		items: make(map[string]*CacheItem),
		ttl:   ttl,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves an item from cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check expiration
	if time.Now().After(item.Expiration) {
		return nil, false
	}

	return item.Value, true
}

// Set stores an item in cache
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &CacheItem{
		Value:      value,
		Expiration: time.Now().Add(c.ttl),
	}
}

// Delete removes an item from cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear removes all items from cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*CacheItem)
}

// cleanup removes expired items
func (c *Cache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()

		now := time.Now()
		for key, item := range c.items {
			if now.After(item.Expiration) {
				delete(c.items, key)
			}
		}

		c.mu.Unlock()
	}
}

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	tokens    int
	maxTokens int
	refillRate time.Duration

	mu        sync.Mutex
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens int, refillRate time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if an operation is allowed
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Refill tokens
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)
	tokensToAdd := int(elapsed / rl.refillRate)

	if tokensToAdd > 0 {
		rl.tokens = min(rl.tokens+tokensToAdd, rl.maxTokens)
		rl.lastRefill = now
	}

	// Check if we have tokens
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

// Wait blocks until a token is available
func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		if rl.Allow() {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(rl.refillRate):
			continue
		}
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
