package hldgen

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/sync/semaphore"
)

// Orchestrator coordinates the multi-agent HLD generation process
type Orchestrator struct {
	cfg             Config
	agents          map[AgentRole]Agent
	critic          *CriticAgent
	refinementLoop  *RefinementLoop
	merger          *ResponseMerger
	contextEngine   *ContextEngine
	llmRouter       *LLMRouter
	enrichedContext *EnrichedContext
	logger          zerolog.Logger
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator(cfg Config, logger zerolog.Logger) (*Orchestrator, error) {
	// Validate configuration
	if err := ValidateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Initialize Context Engine
	contextEngine := NewContextEngine(cfg.Context, logger)

	// Initialize LLM Router
	llmRouter, err := NewLLMRouter(cfg.LLM)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to initialize LLM router, using mock client")
		// Use mock client as fallback
		llmRouter = &LLMRouter{
			cfg: cfg.LLM,
			providers: map[string]LLMClient{
				"mock": NewMockLLMClient("mock", "mock-model"),
			},
		}
	}

	// Initialize agents
	agentMap := make(map[AgentRole]Agent)

	for _, agentCfg := range cfg.Intelligence.MultiAgent.Agents {
		role := AgentRole(agentCfg.Role)

		switch role {
		case RoleArchitect:
			agentMap[role] = NewArchitectAgent(agentCfg)
		case RolePM:
			agentMap[role] = NewSimpleAgent(role, agentCfg)
		case RoleSecurity:
			agentMap[role] = NewSimpleAgent(role, agentCfg)
		case RoleSRE:
			agentMap[role] = NewSimpleAgent(role, agentCfg)
		case RoleQA:
			agentMap[role] = NewSimpleAgent(role, agentCfg)
		}
	}

	// Find critic config
	var criticCfg AgentConfig
	for _, agentCfg := range cfg.Intelligence.MultiAgent.Agents {
		if agentCfg.Role == string(RoleCritic) {
			criticCfg = agentCfg
			break
		}
	}

	// Initialize critic
	critic := NewCriticAgent(criticCfg, cfg.Intelligence.ConsensusThreshold)

	// Initialize merger
	merger := NewResponseMerger("priority_merge")

	// Initialize refinement loop
	refinementLoop := NewRefinementLoop(cfg.Refinement, critic, merger, logger)

	logger.Info().
		Int("agents", len(agentMap)).
		Bool("context_enabled", cfg.Context.Enabled).
		Bool("rag_enabled", cfg.Context.RAG.Enabled).
		Msg("Orchestrator initialized")

	return &Orchestrator{
		cfg:            cfg,
		agents:         agentMap,
		critic:         critic,
		refinementLoop: refinementLoop,
		merger:         merger,
		contextEngine:  contextEngine,
		llmRouter:      llmRouter,
		logger:         logger,
	}, nil
}

// Run executes the HLD generation process
func (o *Orchestrator) Run(ctx context.Context, docs *ConsolidatedDocs) (*HLDOutput, error) {
	// Validate input
	if err := ValidateConsolidatedDocs(docs); err != nil {
		return nil, fmt.Errorf("invalid input docs: %w", err)
	}

	o.logger.Info().
		Str("module", docs.ModuleName).
		Int("services", len(docs.Services)).
		Int("messages", len(docs.Messages)).
		Msg("Starting HLD generation")

	startTime := time.Now()

	// Enrich documentation with context
	enrichedCtx, err := o.contextEngine.Enrich(ctx, docs)
	if err != nil {
		o.logger.Warn().Err(err).Msg("Context enrichment failed, continuing without enrichment")
		// Create minimal enriched context
		enrichedCtx = &EnrichedContext{
			Docs:        docs,
			RAGContext:  []RAGDocument{},
			Sources:     make(map[string]ContextSource),
			GeneratedAt: time.Now(),
		}
	}

	// Store enriched context for agents
	o.enrichedContext = enrichedCtx

	// Run refinement loop with agent thinking function
	output, err := o.refinementLoop.Run(ctx, docs, o.ParallelThink)
	if err != nil {
		return nil, fmt.Errorf("refinement loop failed: %w", err)
	}

	// Set metadata
	output.Metadata.ModuleName = docs.ModuleName
	output.Metadata.SourceCommit = docs.SourceCommit
	output.Metadata.Author = "ProtoDocs HLD Generator"
	output.Metadata.Version = o.cfg.Version
	output.Metadata.GenerationMode = string(o.cfg.DefaultMode)

	// Validate output
	if err := ValidateHLDOutput(output); err != nil {
		o.logger.Warn().Err(err).Msg("Output validation failed")
		output.Warnings = append(output.Warnings, fmt.Sprintf("Validation warnings: %v", err))
	}

	duration := time.Since(startTime)
	o.logger.Info().
		Float64("final_score", output.FinalScore).
		Int("rounds", output.RoundsCompleted).
		Int("rag_docs", len(enrichedCtx.RAGContext)).
		Dur("duration", duration).
		Msg("HLD generation completed")

	return output, nil
}

// ParallelThink executes all agents in parallel
func (o *Orchestrator) ParallelThink(
	ctx context.Context,
	docs *ConsolidatedDocs,
	previousDraft *HLDOutput,
) (map[string]*AgentResponse, error) {

	responses := make(map[string]*AgentResponse)
	var mu sync.Mutex
	var wg sync.WaitGroup

	sem := semaphore.NewWeighted(int64(o.cfg.Performance.Concurrency))

	// Build agent input
	input := &AgentInput{
		Docs:          docs,
		EnrichedCtx:   o.enrichedContext,
		PreviousDraft: previousDraft,
		Round:         o.refinementLoop.CurrentRound(),
		Criticism:     o.refinementLoop.LastCriticism(),
	}

	for role, agent := range o.agents {
		wg.Add(1)
		go func(r AgentRole, a Agent) {
			defer wg.Done()

			if err := sem.Acquire(ctx, 1); err != nil {
				o.logger.Error().Err(err).Str("agent", string(r)).Msg("Semaphore acquire failed")
				return
			}
			defer sem.Release(1)

			agentCtx, cancel := context.WithTimeout(ctx, time.Duration(o.cfg.Performance.TimeoutSeconds)*time.Second)
			defer cancel()

			o.logger.Info().
				Str("agent", string(r)).
				Int("round", input.Round).
				Msg("Agent thinking")

			resp, err := a.Think(agentCtx, input)
			if err != nil {
				o.logger.Warn().Err(err).Str("agent", string(r)).Msg("Agent failed")
				return
			}

			mu.Lock()
			responses[string(r)] = resp
			mu.Unlock()

			o.logger.Info().
				Str("agent", string(r)).
				Float64("confidence", resp.Confidence).
				Int("tokens", resp.TokensUsed).
				Msg("Agent completed")
		}(role, agent)
	}

	wg.Wait()

	if len(responses) == 0 {
		return nil, fmt.Errorf("no agents produced responses")
	}

	return responses, nil
}

// Generate is a convenience method for generating HLD
func Generate(ctx context.Context, docs *ConsolidatedDocs, cfg Config, logger zerolog.Logger) (*HLDOutput, error) {
	orchestrator, err := NewOrchestrator(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("create orchestrator: %w", err)
	}

	return orchestrator.Run(ctx, docs)
}
