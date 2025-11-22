package adapters

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PrometheusMetrics implements EnrichmentMetrics using Prometheus
type PrometheusMetrics struct {
	enrichmentDuration  *prometheus.HistogramVec
	enrichmentTotal     *prometheus.CounterVec
	enrichmentTokens    *prometheus.CounterVec
	cacheHits           *prometheus.CounterVec
	cacheMisses         *prometheus.CounterVec
	ragRetrievals       *prometheus.HistogramVec
	ragResultsCount     *prometheus.HistogramVec
	safetyChecks        *prometheus.CounterVec
	safetyCheckDuration *prometheus.HistogramVec
	llmCost             *prometheus.CounterVec
}

// NewPrometheusMetrics creates a new Prometheus-based metrics collector
func NewPrometheusMetrics(namespace, subsystem string) *PrometheusMetrics {
	m := &PrometheusMetrics{
		enrichmentDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "enrichment_duration_seconds",
				Help:      "Duration of enrichment operations",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"target", "success"},
		),
		enrichmentTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "enrichment_total",
				Help:      "Total number of enrichment operations",
			},
			[]string{"target", "success"},
		),
		enrichmentTokens: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "enrichment_tokens_total",
				Help:      "Total tokens used in enrichment",
			},
			[]string{"target"},
		),
		cacheHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "cache_hits_total",
				Help:      "Total number of cache hits",
			},
			[]string{"target"},
		),
		cacheMisses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "cache_misses_total",
				Help:      "Total number of cache misses",
			},
			[]string{"target"},
		),
		ragRetrievals: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "rag_retrieval_duration_seconds",
				Help:      "Duration of RAG retrieval operations",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"query"},
		),
		ragResultsCount: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "rag_results_count",
				Help:      "Number of results returned by RAG",
				Buckets:   []float64{0, 1, 2, 5, 10, 20, 50},
			},
			[]string{"query"},
		),
		safetyChecks: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "safety_checks_total",
				Help:      "Total number of safety checks",
			},
			[]string{"passed"},
		),
		safetyCheckDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "safety_check_duration_seconds",
				Help:      "Duration of safety check operations",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"passed"},
		),
		llmCost: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "llm_cost_total",
				Help:      "Total cost of LLM usage",
			},
			[]string{"provider", "model"},
		),
	}

	return m
}

// RecordEnrichment records enrichment metrics
func (m *PrometheusMetrics) RecordEnrichment(target string, duration time.Duration, tokensUsed int, success bool) {
	successLabel := "false"
	if success {
		successLabel = "true"
	}

	m.enrichmentDuration.WithLabelValues(target, successLabel).Observe(duration.Seconds())
	m.enrichmentTotal.WithLabelValues(target, successLabel).Inc()

	if tokensUsed > 0 {
		m.enrichmentTokens.WithLabelValues(target).Add(float64(tokensUsed))
	}
}

// RecordCacheHit records a cache hit
func (m *PrometheusMetrics) RecordCacheHit(target string) {
	m.cacheHits.WithLabelValues(target).Inc()
}

// RecordCacheMiss records a cache miss
func (m *PrometheusMetrics) RecordCacheMiss(target string) {
	m.cacheMisses.WithLabelValues(target).Inc()
}

// RecordRAGRetrieval records RAG retrieval metrics
func (m *PrometheusMetrics) RecordRAGRetrieval(query string, resultsCount int, duration time.Duration) {
	m.ragRetrievals.WithLabelValues(query).Observe(duration.Seconds())
	m.ragResultsCount.WithLabelValues(query).Observe(float64(resultsCount))
}

// RecordSafetyCheck records safety check metrics
func (m *PrometheusMetrics) RecordSafetyCheck(passed bool, duration time.Duration) {
	passedLabel := "false"
	if passed {
		passedLabel = "true"
	}

	m.safetyChecks.WithLabelValues(passedLabel).Inc()
	m.safetyCheckDuration.WithLabelValues(passedLabel).Observe(duration.Seconds())
}

// RecordCost records LLM usage cost
func (m *PrometheusMetrics) RecordCost(provider string, model string, tokens int, cost float64) {
	m.llmCost.WithLabelValues(provider, model).Add(cost)
}

// NoOpMetrics is a no-op implementation for when metrics are disabled
type NoOpMetrics struct{}

// NewNoOpMetrics creates a no-op metrics collector
func NewNoOpMetrics() *NoOpMetrics {
	return &NoOpMetrics{}
}

// RecordEnrichment does nothing
func (m *NoOpMetrics) RecordEnrichment(target string, duration time.Duration, tokensUsed int, success bool) {}

// RecordCacheHit does nothing
func (m *NoOpMetrics) RecordCacheHit(target string) {}

// RecordCacheMiss does nothing
func (m *NoOpMetrics) RecordCacheMiss(target string) {}

// RecordRAGRetrieval does nothing
func (m *NoOpMetrics) RecordRAGRetrieval(query string, resultsCount int, duration time.Duration) {}

// RecordSafetyCheck does nothing
func (m *NoOpMetrics) RecordSafetyCheck(passed bool, duration time.Duration) {}

// RecordCost does nothing
func (m *NoOpMetrics) RecordCost(provider string, model string, tokens int, cost float64) {}
