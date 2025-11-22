package enricher

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// EnrichmentManifest tracks enrichment execution for site assembly
type EnrichmentManifest struct {
	mu sync.RWMutex

	Version     string                   `json:"version"`
	StartTime   time.Time                `json:"start_time"`
	EndTime     time.Time                `json:"end_time"`
	Duration    time.Duration            `json:"duration"`
	ConfigHash  string                   `json:"config_hash"`
	ModelUsed   string                   `json:"model_used"`
	ProviderUsed string                  `json:"provider_used"`

	Targets     map[string]*TargetStatus `json:"targets"`
	Statistics  ManifestStatistics       `json:"statistics"`
}

// TargetStatus represents the enrichment status of a single target
type TargetStatus struct {
	TargetID      string    `json:"target_id"`
	Enriched      bool      `json:"enriched"`
	CacheHit      bool      `json:"cache_hit"`
	TokensUsed    int       `json:"tokens_used"`
	Timestamp     time.Time `json:"timestamp"`
	ErrorMessage  string    `json:"error_message,omitempty"`
}

// ManifestStatistics holds aggregate statistics
type ManifestStatistics struct {
	TotalTargets     int     `json:"total_targets"`
	EnrichedTargets  int     `json:"enriched_targets"`
	FailedTargets    int     `json:"failed_targets"`
	CacheHits        int     `json:"cache_hits"`
	TotalTokensUsed  int     `json:"total_tokens_used"`
	SuccessRate      float64 `json:"success_rate"`
	CacheHitRate     float64 `json:"cache_hit_rate"`
}

// NewEnrichmentManifest creates a new manifest
func NewEnrichmentManifest() *EnrichmentManifest {
	return &EnrichmentManifest{
		Version:    "1.0",
		Targets:    make(map[string]*TargetStatus),
		Statistics: ManifestStatistics{},
	}
}

// RecordEnrichment records the result of enriching a target
func (m *EnrichmentManifest) RecordEnrichment(targetID string, success bool, cacheHit bool, tokensUsed int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	status := &TargetStatus{
		TargetID:   targetID,
		Enriched:   success,
		CacheHit:   cacheHit,
		TokensUsed: tokensUsed,
		Timestamp:  time.Now(),
	}

	if !success {
		status.ErrorMessage = "enrichment failed"
	}

	m.Targets[targetID] = status

	// Update statistics
	m.Statistics.TotalTargets = len(m.Targets)

	enrichedCount := 0
	failedCount := 0
	cacheHitCount := 0
	totalTokens := 0

	for _, ts := range m.Targets {
		if ts.Enriched {
			enrichedCount++
		} else {
			failedCount++
		}
		if ts.CacheHit {
			cacheHitCount++
		}
		totalTokens += ts.TokensUsed
	}

	m.Statistics.EnrichedTargets = enrichedCount
	m.Statistics.FailedTargets = failedCount
	m.Statistics.CacheHits = cacheHitCount
	m.Statistics.TotalTokensUsed = totalTokens

	if m.Statistics.TotalTargets > 0 {
		m.Statistics.SuccessRate = float64(enrichedCount) / float64(m.Statistics.TotalTargets)
		m.Statistics.CacheHitRate = float64(cacheHitCount) / float64(m.Statistics.TotalTargets)
	}
}

// RecordError records an error for a target
func (m *EnrichmentManifest) RecordError(targetID string, errMsg string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	status := &TargetStatus{
		TargetID:     targetID,
		Enriched:     false,
		CacheHit:     false,
		TokensUsed:   0,
		Timestamp:    time.Now(),
		ErrorMessage: errMsg,
	}

	m.Targets[targetID] = status

	// Update statistics
	m.Statistics.TotalTargets = len(m.Targets)

	failedCount := 0
	for _, ts := range m.Targets {
		if !ts.Enriched {
			failedCount++
		}
	}
	m.Statistics.FailedTargets = failedCount

	if m.Statistics.TotalTargets > 0 {
		successCount := m.Statistics.TotalTargets - failedCount
		m.Statistics.SuccessRate = float64(successCount) / float64(m.Statistics.TotalTargets)
	}
}

// SetModelInfo sets the model and provider information
func (m *EnrichmentManifest) SetModelInfo(model, provider string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ModelUsed = model
	m.ProviderUsed = provider
}

// SetConfigHash sets the configuration hash
func (m *EnrichmentManifest) SetConfigHash(hash string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ConfigHash = hash
}

// GetStatus returns the status for a specific target
func (m *EnrichmentManifest) GetStatus(targetID string) (*TargetStatus, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	status, ok := m.Targets[targetID]
	return status, ok
}

// GetStatistics returns a copy of the current statistics
func (m *EnrichmentManifest) GetStatistics() ManifestStatistics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Statistics
}

// SaveToFile saves the manifest to a JSON file
func (m *EnrichmentManifest) SaveToFile(path string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write manifest file: %w", err)
	}

	return nil
}

// LoadFromFile loads a manifest from a JSON file
func LoadManifestFromFile(path string) (*EnrichmentManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest file: %w", err)
	}

	var manifest EnrichmentManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("unmarshal manifest: %w", err)
	}

	return &manifest, nil
}

// PrintSummary prints a human-readable summary of the manifest
func (m *EnrichmentManifest) PrintSummary() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	fmt.Println("=== Enrichment Manifest Summary ===")
	fmt.Printf("Version: %s\n", m.Version)
	fmt.Printf("Model: %s (%s)\n", m.ModelUsed, m.ProviderUsed)
	fmt.Printf("Duration: %s\n", m.Duration)
	fmt.Printf("\nStatistics:\n")
	fmt.Printf("  Total Targets: %d\n", m.Statistics.TotalTargets)
	fmt.Printf("  Enriched: %d\n", m.Statistics.EnrichedTargets)
	fmt.Printf("  Failed: %d\n", m.Statistics.FailedTargets)
	fmt.Printf("  Cache Hits: %d\n", m.Statistics.CacheHits)
	fmt.Printf("  Total Tokens: %d\n", m.Statistics.TotalTokensUsed)
	fmt.Printf("  Success Rate: %.2f%%\n", m.Statistics.SuccessRate*100)
	fmt.Printf("  Cache Hit Rate: %.2f%%\n", m.Statistics.CacheHitRate*100)
}

// HasFailures returns true if any targets failed enrichment
func (m *EnrichmentManifest) HasFailures() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Statistics.FailedTargets > 0
}

// GetFailedTargets returns a list of all failed target IDs
func (m *EnrichmentManifest) GetFailedTargets() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var failed []string
	for _, status := range m.Targets {
		if !status.Enriched {
			failed = append(failed, status.TargetID)
		}
	}
	return failed
}
