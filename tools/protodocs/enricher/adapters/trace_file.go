package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/kyivinua/root/tools/protodocs/enricher"
)

// FileTraceSink implements TraceSink by writing to JSON files
type FileTraceSink struct {
	mu         sync.Mutex
	outputPath string
	traces     []*enricher.EnrichmentTrace
}

// NewFileTraceSink creates a new file-based trace sink
func NewFileTraceSink(outputPath string) (*FileTraceSink, error) {
	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create trace directory: %w", err)
	}

	return &FileTraceSink{
		outputPath: outputPath,
		traces:     make([]*enricher.EnrichmentTrace, 0),
	}, nil
}

// RecordTrace records a trace to memory (will be flushed to file later)
func (s *FileTraceSink) RecordTrace(ctx context.Context, trace *enricher.EnrichmentTrace) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.traces = append(s.traces, trace)
	return nil
}

// QueryTraces queries traces based on filter
func (s *FileTraceSink) QueryTraces(ctx context.Context, filter enricher.TraceFilter) ([]*enricher.EnrichmentTrace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	results := make([]*enricher.EnrichmentTrace, 0)

	for _, trace := range s.traces {
		// Apply filters
		if filter.TenantID != "" && trace.Tenant != filter.TenantID {
			continue
		}
		if filter.TargetID != "" && trace.TargetID != filter.TargetID {
			continue
		}
		if !filter.StartTime.IsZero() && trace.Timestamp.Before(filter.StartTime) {
			continue
		}
		if !filter.EndTime.IsZero() && trace.Timestamp.After(filter.EndTime) {
			continue
		}
		if filter.Success != nil && trace.Success != *filter.Success {
			continue
		}

		results = append(results, trace)

		// Apply limit
		if filter.Limit > 0 && len(results) >= filter.Limit {
			break
		}
	}

	// Apply offset
	if filter.Offset > 0 {
		if filter.Offset >= len(results) {
			return []*enricher.EnrichmentTrace{}, nil
		}
		results = results[filter.Offset:]
	}

	return results, nil
}

// Flush writes all traces to file
func (s *FileTraceSink) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.traces) == 0 {
		return nil
	}

	data, err := json.MarshalIndent(s.traces, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal traces: %w", err)
	}

	if err := os.WriteFile(s.outputPath, data, 0644); err != nil {
		return fmt.Errorf("write traces file: %w", err)
	}

	return nil
}

// Close flushes and closes the trace sink
func (s *FileTraceSink) Close() error {
	return s.Flush()
}

// NoOpTraceSink is a no-op implementation for when tracing is disabled
type NoOpTraceSink struct{}

// NewNoOpTraceSink creates a no-op trace sink
func NewNoOpTraceSink() *NoOpTraceSink {
	return &NoOpTraceSink{}
}

// RecordTrace does nothing
func (s *NoOpTraceSink) RecordTrace(ctx context.Context, trace *enricher.EnrichmentTrace) error {
	return nil
}

// QueryTraces returns empty results
func (s *NoOpTraceSink) QueryTraces(ctx context.Context, filter enricher.TraceFilter) ([]*enricher.EnrichmentTrace, error) {
	return []*enricher.EnrichmentTrace{}, nil
}
