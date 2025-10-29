// Package enricher provides advanced AI enrichment capabilities.
package enricher

import (
	"context"
	"fmt"
	"sync"

	"github.com/kyivinua/docgen-tool/internal/docgen"
	"github.com/rs/zerolog"
)

// AdvancedEnricher provides sophisticated AI-powered enrichment.
type AdvancedEnricher struct {
	client            AIClient
	strategies        map[string]EnrichmentStrategy
	config            AdvancedConfig
	logger            zerolog.Logger
	terminology       map[string]string
	terminologyMutex  sync.RWMutex
	batchProcessor    *BatchProcessor
	qualityScorer     *QualityScorer
}

// AdvancedConfig configures advanced enrichment features.
type AdvancedConfig struct {
	// Strategy selection
	UseContextAware      bool
	UseExampleGeneration bool
	UseTerminology       bool
	UseBestPractices     bool
	UseMultiPass         bool

	// Options
	MultiPassCount       int
	BatchSize            int
	DetailLevel          string // minimal, standard, detailed, comprehensive
	Tone                 string // technical, conversational, formal
	IncludeCodeExamples  bool
	LanguagesForExamples []string
	ProjectDomain        string

	// Quality
	MinQualityScore      float64
	EnableQualityScoring bool
}

// NewAdvancedEnricher creates a new advanced enricher.
func NewAdvancedEnricher(client AIClient, config AdvancedConfig, logger zerolog.Logger) *AdvancedEnricher {
	e := &AdvancedEnricher{
		client:         client,
		config:         config,
		logger:         logger,
		strategies:     make(map[string]EnrichmentStrategy),
		terminology:    make(map[string]string),
		batchProcessor: NewBatchProcessor(client, config.BatchSize),
		qualityScorer:  NewQualityScorer(client),
	}

	// Register strategies
	if config.UseContextAware {
		e.strategies["context-aware"] = NewContextAwareStrategy(client)
	}
	if config.UseExampleGeneration {
		e.strategies["example-generation"] = NewExampleGenerationStrategy(client)
	}
	if config.UseTerminology {
		e.strategies["terminology"] = NewTerminologyStrategy(client)
	}
	if config.UseBestPractices {
		e.strategies["best-practice"] = NewBestPracticeStrategy(client)
	}
	if config.UseMultiPass {
		e.strategies["multi-pass"] = NewMultiPassStrategy(client, config.MultiPassCount)
	}

	return e
}

// EnrichService enriches an entire service with advanced strategies.
func (e *AdvancedEnricher) EnrichService(ctx context.Context, service *docgen.Service) (*EnrichmentSummary, error) {
	e.logger.Info().Str("service", service.Name).Msg("Starting advanced enrichment")

	summary := &EnrichmentSummary{
		ServiceName:       service.Name,
		StrategiesApplied: []string{},
		Examples:          []CodeExample{},
		BestPractices:     []BestPractice{},
		Terminology:       make(map[string]string),
		QualityScores:     make(map[string]float64),
		QualityDimensions: make(map[string]float64),
	}

	// Build rich context
	enrichCtx := e.buildContext(service)

	// Enrich service description
	if service.Description == "" || len(service.Description) < 50 {
		result, err := e.enrichWithBestStrategy(ctx, &EnrichmentRequest{
			Type:    "service",
			Name:    service.Name,
			Content: service.Description,
			Context: enrichCtx,
			Options: e.buildOptions(),
		})

		if err != nil {
			e.logger.Warn().Err(err).Msg("Service description enrichment failed")
		} else {
			service.Description = result.EnrichedContent
			summary.QualityScores["service"] = result.QualityScore
			e.mergeTerminology(result.Terminology)
		}
	}

	// Enrich methods
	if err := e.enrichMethods(ctx, service, enrichCtx, summary); err != nil {
		return nil, err
	}

	// Enrich messages
	if err := e.enrichMessages(ctx, service, enrichCtx, summary); err != nil {
		return nil, err
	}

	// Generate examples if enabled
	if e.config.UseExampleGeneration {
		examples, err := e.generateExamples(ctx, service)
		if err != nil {
			e.logger.Warn().Err(err).Msg("Example generation failed")
		} else {
			summary.Examples = examples
		}
	}

	// Analyze best practices if enabled
	if e.config.UseBestPractices {
		practices, err := e.analyzeBestPractices(ctx, service, enrichCtx)
		if err != nil {
			e.logger.Warn().Err(err).Msg("Best practice analysis failed")
		} else {
			summary.BestPractices = practices
		}
	}

	// Overall quality score
	if e.config.EnableQualityScoring {
		assessment, err := e.qualityScorer.ScoreService(ctx, service)
		if err != nil {
			e.logger.Warn().Err(err).Msg("Quality scoring failed")
		} else {
			summary.OverallQuality = assessment.OverallScore
			summary.QualityDimensions = assessment.Dimensions
			summary.Suggestions = assessment.Suggestions
		}
	}

	// Copy terminology
	e.terminologyMutex.RLock()
	for k, v := range e.terminology {
		summary.Terminology[k] = v
	}
	e.terminologyMutex.RUnlock()

	e.logger.Info().
		Str("service", service.Name).
		Float64("quality", summary.OverallQuality).
		Int("examples", len(summary.Examples)).
		Msg("Advanced enrichment complete")

	return summary, nil
}

// enrichMethods enriches all methods in a service.
func (e *AdvancedEnricher) enrichMethods(ctx context.Context, service *docgen.Service, enrichCtx *EnrichmentContext, summary *EnrichmentSummary) error {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 3) // Limit concurrency
	errChan := make(chan error, len(service.Methods))

	for i := range service.Methods {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			method := &service.Methods[idx]
			methodCtx := enrichCtx
			methodCtx.Method = method

			result, err := e.enrichWithBestStrategy(ctx, &EnrichmentRequest{
				Type:    "method",
				Name:    method.Name,
				Content: method.Description,
				Context: methodCtx,
				Options: e.buildOptions(),
			})

			if err != nil {
				errChan <- fmt.Errorf("method %s: %w", method.Name, err)
				return
			}

			method.Description = result.EnrichedContent
			e.mergeTerminology(result.Terminology)
		}(i)
	}

	wg.Wait()
	close(errChan)

	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("method enrichment errors: %v", errs)
	}

	return nil
}

// enrichMessages enriches all messages in a service.
func (e *AdvancedEnricher) enrichMessages(ctx context.Context, service *docgen.Service, enrichCtx *EnrichmentContext, summary *EnrichmentSummary) error {
	for i := range service.Messages {
		message := &service.Messages[i]
		msgCtx := enrichCtx
		msgCtx.Message = message

		// Enrich message description
		if message.Description == "" || len(message.Description) < 30 {
			result, err := e.enrichWithBestStrategy(ctx, &EnrichmentRequest{
				Type:    "message",
				Name:    message.Name,
				Content: message.Description,
				Context: msgCtx,
				Options: e.buildOptions(),
			})

			if err != nil {
				e.logger.Warn().Err(err).Str("message", message.Name).Msg("Message enrichment failed")
			} else {
				message.Description = result.EnrichedContent
				e.mergeTerminology(result.Terminology)
			}
		}

		// Enrich fields
		for j := range message.Fields {
			field := &message.Fields[j]
			if field.Description == "" {
				fieldCtx := msgCtx
				fieldCtx.Field = field

				result, err := e.enrichWithBestStrategy(ctx, &EnrichmentRequest{
					Type:    "field",
					Name:    field.Name,
					Content: field.Description,
					Context: fieldCtx,
					Options: e.buildOptions(),
				})

				if err != nil {
					e.logger.Warn().Err(err).Str("field", field.Name).Msg("Field enrichment failed")
				} else {
					field.Description = result.EnrichedContent
				}
			}
		}
	}

	return nil
}

// enrichWithBestStrategy selects and applies the best enrichment strategy.
func (e *AdvancedEnricher) enrichWithBestStrategy(ctx context.Context, req *EnrichmentRequest) (*EnrichmentResult, error) {
	// Try strategies in priority order
	strategyOrder := []string{"multi-pass", "context-aware", "terminology"}

	var lastResult *EnrichmentResult
	var lastErr error

	for _, strategyName := range strategyOrder {
		strategy, exists := e.strategies[strategyName]
		if !exists {
			continue
		}

		result, err := strategy.Enrich(ctx, req)
		if err != nil {
			e.logger.Warn().Err(err).Str("strategy", strategyName).Msg("Strategy failed")
			lastErr = err
			continue
		}

		// Check if result meets quality threshold
		if e.config.EnableQualityScoring && result.QualityScore < e.config.MinQualityScore {
			e.logger.Debug().
				Str("strategy", strategyName).
				Float64("score", result.QualityScore).
				Msg("Quality below threshold, trying next strategy")
			lastResult = result
			continue
		}

		// Success!
		return result, nil
	}

	// If all strategies failed, return last result or error
	if lastResult != nil {
		return lastResult, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return &EnrichmentResult{
		EnrichedContent: req.Content,
		Confidence:      0.5,
	}, nil
}

// generateExamples generates code examples for all methods.
func (e *AdvancedEnricher) generateExamples(ctx context.Context, service *docgen.Service) ([]CodeExample, error) {
	strategy := e.strategies["example-generation"]
	if strategy == nil {
		return nil, nil
	}

	var allExamples []CodeExample

	for _, method := range service.Methods {
		enrichCtx := &EnrichmentContext{
			Service: service,
			Method:  &method,
		}

		result, err := strategy.Enrich(ctx, &EnrichmentRequest{
			Type:    "method",
			Name:    method.Name,
			Context: enrichCtx,
			Options: e.buildOptions(),
		})

		if err != nil {
			e.logger.Warn().Err(err).Str("method", method.Name).Msg("Example generation failed")
			continue
		}

		allExamples = append(allExamples, result.Examples...)
	}

	return allExamples, nil
}

// analyzeBestPractices analyzes the service for best practice recommendations.
func (e *AdvancedEnricher) analyzeBestPractices(ctx context.Context, service *docgen.Service, enrichCtx *EnrichmentContext) ([]BestPractice, error) {
	strategy := e.strategies["best-practice"]
	if strategy == nil {
		return nil, nil
	}

	result, err := strategy.Enrich(ctx, &EnrichmentRequest{
		Type:    "service",
		Name:    service.Name,
		Content: service.Description,
		Context: enrichCtx,
		Options: e.buildOptions(),
	})

	if err != nil {
		return nil, err
	}

	return result.BestPractices, nil
}

// buildContext creates a rich enrichment context.
func (e *AdvancedEnricher) buildContext(service *docgen.Service) *EnrichmentContext {
	e.terminologyMutex.RLock()
	terminology := make(map[string]string)
	for k, v := range e.terminology {
		terminology[k] = v
	}
	e.terminologyMutex.RUnlock()

	return &EnrichmentContext{
		Service:       service,
		ProjectDomain: e.config.ProjectDomain,
		Terminology:   terminology,
	}
}

// buildOptions creates enrichment options from config.
func (e *AdvancedEnricher) buildOptions() *EnrichmentOptions {
	return &EnrichmentOptions{
		IncludeExamples:      e.config.IncludeCodeExamples,
		IncludeBestPractices: e.config.UseBestPractices,
		CheckConsistency:     e.config.UseTerminology,
		GenerateCodeSamples:  e.config.UseExampleGeneration,
		DetailLevel:          e.config.DetailLevel,
		Tone:                 e.config.Tone,
		MaxLength:            2048,
	}
}

// mergeTerminology merges new terminology into the global map.
func (e *AdvancedEnricher) mergeTerminology(newTerms map[string]string) {
	if len(newTerms) == 0 {
		return
	}

	e.terminologyMutex.Lock()
	defer e.terminologyMutex.Unlock()

	for term, def := range newTerms {
		if _, exists := e.terminology[term]; !exists {
			e.terminology[term] = def
		}
	}
}

// EnrichmentSummary summarizes the enrichment results.
type EnrichmentSummary struct {
	ServiceName        string
	StrategiesApplied  []string
	Examples           []CodeExample
	BestPractices      []BestPractice
	Terminology        map[string]string
	QualityScores      map[string]float64
	OverallQuality     float64
	QualityDimensions  map[string]float64
	Suggestions        []string
	Warnings           []string
}

// DefaultAdvancedConfig returns default advanced configuration.
func DefaultAdvancedConfig() AdvancedConfig {
	return AdvancedConfig{
		UseContextAware:      true,
		UseExampleGeneration: false,
		UseTerminology:       true,
		UseBestPractices:     false,
		UseMultiPass:         false,
		MultiPassCount:       2,
		BatchSize:            5,
		DetailLevel:          "standard",
		Tone:                 "technical",
		IncludeCodeExamples:  false,
		LanguagesForExamples: []string{"go", "python"},
		MinQualityScore:      70.0,
		EnableQualityScoring: false,
	}
}
