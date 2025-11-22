package enricher

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/semaphore"
)

// Enricher is the main orchestrator for documentation enrichment
type Enricher struct {
	config   *EnrichmentConfig
	llm      LLMClient
	rag      RAGRetriever
	cache    EnrichmentCache
	safety   SafetyGuard
	policy   PolicyEngine
	metrics  EnrichmentMetrics
	trace    TraceSink
	templates PromptTemplateEngine
	strategy SmartStrategy

	// Concurrency control
	sem *semaphore.Weighted
	mu  sync.RWMutex

	// Manifest tracking
	manifest *EnrichmentManifest
}

// NewEnricher creates a new enricher instance
func NewEnricher(
	config *EnrichmentConfig,
	llm LLMClient,
	rag RAGRetriever,
	cache EnrichmentCache,
	safety SafetyGuard,
	policy PolicyEngine,
	metrics EnrichmentMetrics,
	trace TraceSink,
	templates PromptTemplateEngine,
	strategy SmartStrategy,
) (*Enricher, error) {
	if config == nil {
		config = DefaultConfig()
	}

	e := &Enricher{
		config:    config,
		llm:       llm,
		rag:       rag,
		cache:     cache,
		safety:    safety,
		policy:    policy,
		metrics:   metrics,
		trace:     trace,
		templates: templates,
		strategy:  strategy,
		sem:       semaphore.NewWeighted(int64(config.Concurrency.MaxConcurrent)),
		manifest:  NewEnrichmentManifest(),
	}

	return e, nil
}

// EnrichModel enriches an entire API documentation model
func (e *Enricher) EnrichModel(ctx context.Context, model *ApiDocModel, tenant string) error {
	e.mu.Lock()
	e.manifest.StartTime = time.Now()
	e.mu.Unlock()

	var wg sync.WaitGroup
	errChan := make(chan error, 100)

	for i := range model.Modules {
		module := &model.Modules[i]

		// Enrich services
		for j := range module.Services {
			service := &module.Services[j]

			// Enrich methods
			for k := range service.Methods {
				method := &service.Methods[k]

				wg.Add(1)
				go func(mod *ModuleDoc, svc *ServiceDoc, mth *MethodDoc) {
					defer wg.Done()

					if err := e.sem.Acquire(ctx, 1); err != nil {
						errChan <- fmt.Errorf("acquire semaphore: %w", err)
						return
					}
					defer e.sem.Release(1)

					target := &MethodTarget{
						Method:  mth,
						Service: svc,
						Module:  mod,
					}

					if err := e.EnrichTarget(ctx, target, tenant); err != nil {
						errChan <- fmt.Errorf("enrich method %s: %w", mth.FullName, err)
					}
				}(module, service, method)
			}
		}

		// Enrich messages
		for j := range module.Messages {
			message := &module.Messages[j]

			wg.Add(1)
			go func(mod *ModuleDoc, msg *MessageDoc) {
				defer wg.Done()

				if err := e.sem.Acquire(ctx, 1); err != nil {
					errChan <- fmt.Errorf("acquire semaphore: %w", err)
					return
				}
				defer e.sem.Release(1)

				target := &MessageTarget{
					Message: msg,
					Module:  mod,
				}

				if err := e.EnrichTarget(ctx, target, tenant); err != nil {
					errChan <- fmt.Errorf("enrich message %s: %w", msg.FullName, err)
				}
			}(module, message)
		}
	}

	// Wait for all enrichments to complete
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Collect errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	e.mu.Lock()
	e.manifest.EndTime = time.Now()
	e.manifest.Duration = e.manifest.EndTime.Sub(e.manifest.StartTime)
	e.mu.Unlock()

	if len(errs) > 0 {
		return fmt.Errorf("enrichment errors: %v", errs)
	}

	return nil
}

// EnrichTarget enriches a single documentation target
func (e *Enricher) EnrichTarget(ctx context.Context, target EnrichmentTarget, tenant string) error {
	startTime := time.Now()
	traceID := uuid.New().String()

	trace := &EnrichmentTrace{
		TraceID:      traceID,
		Tenant:       tenant,
		TargetID:     target.GetIdentifier(),
		TargetType:   fmt.Sprintf("%T", target),
		Timestamp:    startTime,
		OriginalDocs: target.GetCurrentDocs(),
	}

	defer func() {
		trace.Duration = time.Since(startTime)
		if e.trace != nil {
			_ = e.trace.RecordTrace(ctx, trace)
		}
	}()

	// Step 1: Policy check
	if e.policy != nil && e.config.Policy.Enabled {
		decision, err := e.policy.Evaluate(ctx, target, tenant)
		if err != nil {
			trace.Success = false
			trace.ErrorMessage = fmt.Sprintf("policy evaluation failed: %v", err)
			return fmt.Errorf("policy evaluation: %w", err)
		}
		trace.PolicyDecision = decision

		if !decision.Allowed {
			trace.Success = false
			trace.ErrorMessage = "policy denied enrichment"
			return fmt.Errorf("policy denied: %s", decision.Reason)
		}

		if decision.RequiresApproval {
			// In real implementation, this would trigger approval workflow
			trace.Success = false
			trace.ErrorMessage = "requires approval"
			return fmt.Errorf("enrichment requires approval")
		}
	}

	// Step 2: Check cache
	if e.cache != nil && e.config.Cache.Enabled {
		cacheKey := fmt.Sprintf("enriched:%s:%s", tenant, target.GetIdentifier())
		if cached, ok := e.cache.Get(ctx, cacheKey); ok {
			target.SetEnrichedDocs(cached)
			trace.Success = true
			trace.CacheHit = true
			trace.EnrichedDocs = cached

			if e.metrics != nil {
				e.metrics.RecordCacheHit(target.GetIdentifier())
			}

			e.mu.Lock()
			e.manifest.RecordEnrichment(target.GetIdentifier(), true, true, 0)
			e.mu.Unlock()

			return nil
		}

		if e.metrics != nil {
			e.metrics.RecordCacheMiss(target.GetIdentifier())
		}
	}

	// Step 3: Decide RAG vs. base LLM
	useRAG := false
	var ragDocs []RAGDocument

	if e.rag != nil && e.config.RAG.Enabled {
		if e.strategy != nil {
			useRAG = e.strategy.ShouldUseRAG(target, target.GetContext())
		} else {
			useRAG = true // Default to using RAG if available
		}

		if useRAG {
			topK := e.config.RAG.TopK
			if e.strategy != nil {
				topK = e.strategy.GetTopKForTarget(target)
			}

			ragStart := time.Now()
			docs, err := e.rag.RetrieveContext(ctx, target.GetCurrentDocs(), topK)
			ragDuration := time.Since(ragStart)

			if err != nil {
				// Fallback to base LLM on RAG failure
				useRAG = false
			} else {
				ragDocs = docs
				trace.RAGContext = ragDocs

				if e.metrics != nil {
					e.metrics.RecordRAGRetrieval(target.GetCurrentDocs(), len(docs), ragDuration)
				}
			}
		}
	}

	// Step 4: Build prompt
	promptData := map[string]interface{}{
		"target":       target,
		"context":      target.GetContext(),
		"current_docs": target.GetCurrentDocs(),
		"use_rag":      useRAG,
	}

	if useRAG && len(ragDocs) > 0 {
		contextStr := ""
		for _, doc := range ragDocs {
			contextStr += fmt.Sprintf("- %s\n", doc.Content)
		}
		promptData["rag_context"] = contextStr
	}

	templateName := "enrich_method"
	if _, ok := target.(*MessageTarget); ok {
		templateName = "enrich_message"
	}

	prompt, err := e.templates.Render(templateName, promptData)
	if err != nil {
		trace.Success = false
		trace.ErrorMessage = fmt.Sprintf("render template: %v", err)
		return fmt.Errorf("render template: %w", err)
	}
	trace.PromptUsed = prompt
	trace.ModelUsed = e.llm.GetModelName()

	// Step 5: Generate enrichment
	llmStart := time.Now()
	enriched, err := e.llm.GenerateCompletion(ctx, prompt)
	llmDuration := time.Since(llmStart)

	if err != nil {
		trace.Success = false
		trace.ErrorMessage = fmt.Sprintf("LLM generation: %v", err)

		if e.metrics != nil {
			e.metrics.RecordEnrichment(target.GetIdentifier(), llmDuration, 0, false)
		}

		e.mu.Lock()
		e.manifest.RecordEnrichment(target.GetIdentifier(), false, false, 0)
		e.mu.Unlock()

		return fmt.Errorf("LLM generation: %w", err)
	}

	trace.EnrichedDocs = enriched
	// Estimate tokens (rough approximation: 1 token ≈ 4 chars)
	tokensUsed := (len(prompt) + len(enriched)) / 4
	trace.TokensUsed = tokensUsed

	// Step 6: Safety validation
	if e.safety != nil && e.config.Safety.Enabled {
		safetyStart := time.Now()
		safetyReport, err := e.safety.Validate(ctx, target.GetCurrentDocs(), enriched, target)
		safetyDuration := time.Since(safetyStart)

		trace.SafetyReport = safetyReport

		if e.metrics != nil {
			e.metrics.RecordSafetyCheck(safetyReport != nil && safetyReport.Passed, safetyDuration)
		}

		if err != nil || (safetyReport != nil && !safetyReport.Passed) {
			trace.Success = false
			if err != nil {
				trace.ErrorMessage = fmt.Sprintf("safety check error: %v", err)
			} else {
				trace.ErrorMessage = fmt.Sprintf("safety check failed: %v", safetyReport.FailureReasons)
			}

			if e.metrics != nil {
				e.metrics.RecordEnrichment(target.GetIdentifier(), llmDuration, tokensUsed, false)
			}

			e.mu.Lock()
			e.manifest.RecordEnrichment(target.GetIdentifier(), false, false, tokensUsed)
			e.mu.Unlock()

			return fmt.Errorf("safety validation failed")
		}
	}

	// Step 7: Apply enrichment
	target.SetEnrichedDocs(enriched)
	trace.Success = true

	// Step 8: Cache result
	if e.cache != nil && e.config.Cache.Enabled {
		cacheKey := fmt.Sprintf("enriched:%s:%s", tenant, target.GetIdentifier())
		_ = e.cache.Set(ctx, cacheKey, enriched, e.config.Cache.TTL)
	}

	// Step 9: Record metrics
	if e.metrics != nil {
		e.metrics.RecordEnrichment(target.GetIdentifier(), llmDuration, tokensUsed, true)
	}

	e.mu.Lock()
	e.manifest.RecordEnrichment(target.GetIdentifier(), true, false, tokensUsed)
	e.mu.Unlock()

	return nil
}

// GetManifest returns the current enrichment manifest
func (e *Enricher) GetManifest() *EnrichmentManifest {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.manifest
}

// Close cleans up enricher resources
func (e *Enricher) Close() error {
	// In a real implementation, this would close connections, flush buffers, etc.
	return nil
}
