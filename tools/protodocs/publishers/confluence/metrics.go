package confluence

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for Confluence publisher
type Metrics struct {
	// API request metrics
	APIRequestsTotal     *prometheus.CounterVec
	APIRequestDuration   *prometheus.HistogramVec
	APIRequestErrors     *prometheus.CounterVec
	APIRetries           *prometheus.CounterVec

	// Cache metrics
	CacheHits            prometheus.Counter
	CacheMisses          prometheus.Counter
	CacheSize            prometheus.Gauge
	CacheEvictions       prometheus.Counter

	// Batch operation metrics
	BatchOperationsTotal *prometheus.CounterVec
	BatchOperationSize   *prometheus.HistogramVec

	// Circuit breaker metrics
	CircuitBreakerState  *prometheus.GaugeVec
	CircuitBreakerTrips  *prometheus.CounterVec

	// Publisher metrics
	PagesCreated         prometheus.Counter
	PagesUpdated         prometheus.Counter
	PagesDeleted         prometheus.Counter
	PublishDuration      *prometheus.HistogramVec

	// Attachment metrics
	AttachmentsUploaded  prometheus.Counter
	AttachmentSize       prometheus.Histogram
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics(namespace string) *Metrics {
	if namespace == "" {
		namespace = "confluence_publisher"
	}

	return &Metrics{
		// API request metrics
		APIRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "api_requests_total",
				Help:      "Total number of Confluence API requests",
			},
			[]string{"operation", "status"},
		),

		APIRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "api_request_duration_seconds",
				Help:      "Histogram of Confluence API request durations",
				Buckets:   prometheus.DefBuckets, // 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
			},
			[]string{"operation"},
		),

		APIRequestErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "api_request_errors_total",
				Help:      "Total number of Confluence API request errors by type",
			},
			[]string{"operation", "error_type"},
		),

		APIRetries: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "api_retries_total",
				Help:      "Total number of API request retries",
			},
			[]string{"operation"},
		),

		// Cache metrics
		CacheHits: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_hits_total",
				Help:      "Total number of cache hits",
			},
		),

		CacheMisses: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_misses_total",
				Help:      "Total number of cache misses",
			},
		),

		CacheSize: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "cache_size",
				Help:      "Current number of items in cache",
			},
		),

		CacheEvictions: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_evictions_total",
				Help:      "Total number of cache evictions",
			},
		),

		// Batch operation metrics
		BatchOperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "batch_operations_total",
				Help:      "Total number of batch operations",
			},
			[]string{"operation", "status"},
		),

		BatchOperationSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "batch_operation_size",
				Help:      "Size of batch operations (number of items)",
				Buckets:   []float64{1, 5, 10, 25, 50, 100, 250, 500},
			},
			[]string{"operation"},
		),

		// Circuit breaker metrics
		CircuitBreakerState: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "circuit_breaker_state",
				Help:      "Circuit breaker state (0=closed, 1=half-open, 2=open)",
			},
			[]string{"name"},
		),

		CircuitBreakerTrips: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "circuit_breaker_trips_total",
				Help:      "Total number of circuit breaker trips",
			},
			[]string{"name"},
		),

		// Publisher metrics
		PagesCreated: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "pages_created_total",
				Help:      "Total number of pages created",
			},
		),

		PagesUpdated: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "pages_updated_total",
				Help:      "Total number of pages updated",
			},
		),

		PagesDeleted: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "pages_deleted_total",
				Help:      "Total number of pages deleted",
			},
		),

		PublishDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "publish_duration_seconds",
				Help:      "Histogram of publish operation durations",
				Buckets:   []float64{0.1, 0.5, 1, 2.5, 5, 10, 30, 60, 120, 300},
			},
			[]string{"mode"}, // single, batch, consolidated
		),

		// Attachment metrics
		AttachmentsUploaded: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "attachments_uploaded_total",
				Help:      "Total number of attachments uploaded",
			},
		),

		AttachmentSize: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "attachment_size_bytes",
				Help:      "Size of uploaded attachments in bytes",
				Buckets:   prometheus.ExponentialBuckets(1024, 2, 20), // 1KB to ~1GB
			},
		),
	}
}

// RecordAPIRequest records an API request with duration and status
func (m *Metrics) RecordAPIRequest(operation string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
		// Record error type
		errType := "unknown"
		// Try to extract error type from our custom errors
		if e, ok := err.(interface{ Error() string }); ok {
			errStr := e.Error()
			if contains(errStr, "validation") {
				errType = "validation"
			} else if contains(errStr, "permission") || contains(errStr, "authentication") {
				errType = "permission"
			} else if contains(errStr, "not found") {
				errType = "not_found"
			} else if contains(errStr, "network") {
				errType = "network"
			} else if contains(errStr, "rate limit") {
				errType = "rate_limit"
			} else if contains(errStr, "timeout") {
				errType = "timeout"
			}
		}
		m.APIRequestErrors.WithLabelValues(operation, errType).Inc()
	}

	m.APIRequestsTotal.WithLabelValues(operation, status).Inc()
	m.APIRequestDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordRetry records a retry attempt for an operation
func (m *Metrics) RecordRetry(operation string) {
	m.APIRetries.WithLabelValues(operation).Inc()
}

// RecordCacheHit records a cache hit
func (m *Metrics) RecordCacheHit() {
	m.CacheHits.Inc()
}

// RecordCacheMiss records a cache miss
func (m *Metrics) RecordCacheMiss() {
	m.CacheMisses.Inc()
}

// UpdateCacheSize updates the current cache size
func (m *Metrics) UpdateCacheSize(size int) {
	m.CacheSize.Set(float64(size))
}

// RecordCacheEviction records a cache eviction
func (m *Metrics) RecordCacheEviction() {
	m.CacheEvictions.Inc()
}

// RecordBatchOperation records a batch operation
func (m *Metrics) RecordBatchOperation(operation string, size int, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	m.BatchOperationsTotal.WithLabelValues(operation, status).Inc()
	m.BatchOperationSize.WithLabelValues(operation).Observe(float64(size))
}

// UpdateCircuitBreakerState updates circuit breaker state
// state: 0=closed, 1=half-open, 2=open
func (m *Metrics) UpdateCircuitBreakerState(name string, state int) {
	m.CircuitBreakerState.WithLabelValues(name).Set(float64(state))
}

// RecordCircuitBreakerTrip records a circuit breaker trip (transition to open)
func (m *Metrics) RecordCircuitBreakerTrip(name string) {
	m.CircuitBreakerTrips.WithLabelValues(name).Inc()
}

// RecordPageCreated records a page creation
func (m *Metrics) RecordPageCreated() {
	m.PagesCreated.Inc()
}

// RecordPageUpdated records a page update
func (m *Metrics) RecordPageUpdated() {
	m.PagesUpdated.Inc()
}

// RecordPageDeleted records a page deletion
func (m *Metrics) RecordPageDeleted() {
	m.PagesDeleted.Inc()
}

// RecordPublishDuration records the duration of a publish operation
func (m *Metrics) RecordPublishDuration(mode string, duration time.Duration) {
	m.PublishDuration.WithLabelValues(mode).Observe(duration.Seconds())
}

// RecordAttachmentUpload records an attachment upload
func (m *Metrics) RecordAttachmentUpload(sizeBytes int) {
	m.AttachmentsUploaded.Inc()
	m.AttachmentSize.Observe(float64(sizeBytes))
}

// Helper function to check if string contains substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
		 len(s) > len(substr) &&
		 (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		  findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// DefaultMetrics is the default metrics instance
var DefaultMetrics = NewMetrics("confluence_publisher")
