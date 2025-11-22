package hldgen

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// RefinementLoop manages the iterative improvement process
type RefinementLoop struct {
	cfg            RefinementConfig
	critic         CriticInterface
	merger         MergerInterface
	logger         zerolog.Logger

	currentRound   int
	bestDraft      *HLDOutput
	bestScore      float64
	lastCriticism  []Criticism
}

// CriticInterface defines the interface for critique operations
type CriticInterface interface {
	CritiqueAgent(ctx context.Context, role string, response *AgentResponse, all map[string]*AgentResponse, docs *ConsolidatedDocs) (Criticism, error)
	EvaluateConsensus(criticisms []Criticism) float64
}

// MergerInterface defines the interface for response merging
type MergerInterface interface {
	Merge(responses map[string]*AgentResponse) *HLDOutput
}

// AgentThinkFunc is a function that executes agent thinking
type AgentThinkFunc func(ctx context.Context, docs *ConsolidatedDocs, previousDraft *HLDOutput) (map[string]*AgentResponse, error)

// NewRefinementLoop creates a new refinement loop
func NewRefinementLoop(cfg RefinementConfig, critic CriticInterface, merger MergerInterface, logger zerolog.Logger) *RefinementLoop {
	return &RefinementLoop{
		cfg:     cfg,
		critic:  critic,
		merger:  merger,
		logger:  logger,
	}
}

// Run executes the refinement loop
func (r *RefinementLoop) Run(
	ctx context.Context,
	docs *ConsolidatedDocs,
	thinkFunc AgentThinkFunc,
) (*HLDOutput, error) {

	startTime := time.Now()
	r.currentRound = 0
	r.bestDraft = nil
	r.bestScore = 0.0
	r.lastCriticism = make([]Criticism, 0)

	var finalDraft *HLDOutput
	refinementLog := strings.Builder{}

	for r.currentRound < r.cfg.MaxRounds {
		r.currentRound++
		roundStart := time.Now()

		r.logger.Info().
			Int("round", r.currentRound).
			Msg("Starting refinement round")

		refinementLog.WriteString(fmt.Sprintf("## Round %d\n\n", r.currentRound))

		// 1. Execute parallel agent thinking
		responses, err := thinkFunc(ctx, docs, r.bestDraft)
		if err != nil {
			return nil, fmt.Errorf("agent think failed at round %d: %w", r.currentRound, err)
		}

		// 2. Critic evaluates each agent
		criticisms := make([]Criticism, 0)
		hasCritical := false

		for role, resp := range responses {
			crit, err := r.critic.CritiqueAgent(ctx, role, resp, responses, docs)
			if err != nil {
				r.logger.Warn().Err(err).Str("agent", role).Msg("Critic failed")
				continue
			}
			criticisms = append(criticisms, crit)
			if crit.Fatal {
				hasCritical = true
			}

			// Log criticism
			refinementLog.WriteString(fmt.Sprintf("**%s** (Score: %.2f)\n", role, crit.Score))
			if len(crit.Issues) > 0 {
				refinementLog.WriteString("Issues:\n")
				for _, issue := range crit.Issues {
					refinementLog.WriteString(fmt.Sprintf("- [%s] %s: %s\n", issue.Severity, issue.Type, issue.Message))
				}
			}
			refinementLog.WriteString("\n")
		}

		r.lastCriticism = criticisms

		// 3. Calculate consensus
		score := r.critic.EvaluateConsensus(criticisms)
		draft := r.merger.Merge(responses)
		draft.RefinementLog = refinementLog.String()

		// 4. Update best draft
		improvement := score - r.bestScore
		if score > r.bestScore || r.bestDraft == nil {
			r.bestScore = score
			r.bestDraft = draft
			r.logger.Info().
				Int("round", r.currentRound).
				Float64("score", score).
				Float64("improvement", improvement).
				Msg("New best draft achieved")
		}

		roundDuration := time.Since(roundStart)
		r.logger.Info().
			Int("round", r.currentRound).
			Float64("score", score).
			Dur("duration", roundDuration).
			Msg("Round completed")

		// 5. Check stopping conditions
		if score >= r.cfg.ConsensusThreshold && !hasCritical {
			r.logger.Info().
				Int("rounds", r.currentRound).
				Float64("final_score", score).
				Msg("Consensus threshold reached")
			finalDraft = draft
			finalDraft.ConsensusScore = score
			finalDraft.RoundsCompleted = r.currentRound
			break
		}

		// 6. Critical issues force continuation
		if hasCritical && r.cfg.CriticalIssueOverride {
			criticalCount := r.countCritical(criticisms)
			r.logger.Warn().
				Int("critical_issues", criticalCount).
				Msg("Critical issues found - forcing refinement")
		}

		// 7. Check stagnation
		if r.currentRound > 1 && improvement < r.cfg.MinImprovementPerRound {
			r.logger.Warn().
				Float64("improvement", improvement).
				Msg("Stagnation detected")
			if r.cfg.FallbackOnStagnation {
				finalDraft = r.bestDraft
				finalDraft.ConsensusScore = r.bestScore
				finalDraft.RoundsCompleted = r.currentRound
				break
			}
		}
	}

	// 8. Fallback if max rounds exceeded
	if finalDraft == nil {
		r.logger.Warn().
			Int("max_rounds", r.cfg.MaxRounds).
			Msg("Max refinement rounds exceeded - using best available draft")
		finalDraft = r.bestDraft
		finalDraft.ConsensusScore = r.bestScore
		finalDraft.RoundsCompleted = r.currentRound
	}

	// 9. Final validation
	if r.cfg.FinalValidation && finalDraft != nil {
		finalScore := finalDraft.ConsensusScore
		if finalScore < 0.85 {
			r.logger.Error().
				Float64("final_score", finalScore).
				Msg("Final validation below threshold")
		}
	}

	duration := time.Since(startTime)
	r.logger.Info().
		Float64("final_score", finalDraft.ConsensusScore).
		Int("rounds", finalDraft.RoundsCompleted).
		Dur("total_duration", duration).
		Msg("Refinement loop completed")

	finalDraft.FinalScore = finalDraft.ConsensusScore
	finalDraft.GeneratedAt = time.Now()

	return finalDraft, nil
}

// CurrentRound returns the current round number
func (r *RefinementLoop) CurrentRound() int {
	return r.currentRound
}

// LastCriticism returns the criticism from the last round
func (r *RefinementLoop) LastCriticism() []Criticism {
	return r.lastCriticism
}

// countCritical counts critical issues
func (r *RefinementLoop) countCritical(criticisms []Criticism) int {
	count := 0
	for _, crit := range criticisms {
		for _, issue := range crit.Issues {
			if issue.Severity == SeverityCritical {
				count++
			}
		}
	}
	return count
}
