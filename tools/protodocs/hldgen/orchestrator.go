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
	cfg            Config
	agents         map[AgentRole]Agent
	critic         *CriticAgent
	refinementLoop *RefinementLoop
	merger         *ResponseMerger
	logger         zerolog.Logger
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator(cfg Config, logger zerolog.Logger) (*Orchestrator, error) {
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

	return &Orchestrator{
		cfg:            cfg,
		agents:         agentMap,
		critic:         critic,
		refinementLoop: refinementLoop,
		merger:         merger,
		logger:         logger,
	}, nil
}

// Run executes the HLD generation process
func (o *Orchestrator) Run(ctx context.Context, docs *ConsolidatedDocs) (*HLDOutput, error) {
	o.logger.Info().
		Str("module", docs.ModuleName).
		Msg("Starting HLD generation")

	startTime := time.Now()

	// Run refinement loop with agent thinking function
	output, err := o.refinementLoop.Run(ctx, docs, o.ParallelThink)
	if err != nil {
		return nil, fmt.Errorf("refinement loop failed: %w", err)
	}

	// Set metadata
	output.Metadata.ModuleName = docs.ModuleName
	output.Metadata.SourceCommit = docs.SourceCommit
	output.Metadata.Author = "ProtoDocs HLD Generator"

	duration := time.Since(startTime)
	o.logger.Info().
		Float64("final_score", output.FinalScore).
		Int("rounds", output.RoundsCompleted).
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
