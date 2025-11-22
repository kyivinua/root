package hldgen

import (
	"fmt"
	"strings"

	"github.com/kyivinua/docgen-tool/tools/protodocs/internal/validation"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface
func (v *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", v.Field, v.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError
}

// Error implements the error interface
func (ve *ValidationErrors) Error() string {
	var sb strings.Builder
	sb.WriteString("validation errors:\n")
	for _, err := range ve.Errors {
		sb.WriteString(fmt.Sprintf("  - %s\n", err.Error()))
	}
	return sb.String()
}

// HasErrors returns true if there are validation errors
func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.Errors) > 0
}

// Add adds a validation error
func (ve *ValidationErrors) Add(field, message string) {
	ve.Errors = append(ve.Errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// ValidateConfig validates the HLD Generator configuration
func ValidateConfig(cfg *Config) error {
	ve := &ValidationErrors{}

	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	// Validate refinement config
	if cfg.Refinement.MaxRounds < 1 || cfg.Refinement.MaxRounds > 10 {
		ve.Add("refinement.max_rounds", "must be between 1 and 10")
	}
	if cfg.Refinement.ConsensusThreshold < 0.0 || cfg.Refinement.ConsensusThreshold > 1.0 {
		ve.Add("refinement.consensus_threshold", "must be between 0.0 and 1.0")
	}
	if cfg.Refinement.MinImprovementPerRound < 0.0 || cfg.Refinement.MinImprovementPerRound > 1.0 {
		ve.Add("refinement.min_improvement_per_round", "must be between 0.0 and 1.0")
	}

	// Validate intelligence config
	if cfg.Intelligence.ConsensusThreshold < 0.0 || cfg.Intelligence.ConsensusThreshold > 1.0 {
		ve.Add("intelligence.consensus_threshold", "must be between 0.0 and 1.0")
	}

	// Validate multi-agent config
	if cfg.Intelligence.MultiAgent.Enabled {
		if len(cfg.Intelligence.MultiAgent.Agents) == 0 {
			ve.Add("intelligence.multi_agent.agents", "at least one agent must be configured")
		}

		totalWeight := 0.0
		for _, agent := range cfg.Intelligence.MultiAgent.Agents {
			if agent.Role == "" {
				ve.Add("intelligence.multi_agent.agents", "agent role cannot be empty")
			}
			if agent.Weight < 0.0 || agent.Weight > 1.0 {
				ve.Add(fmt.Sprintf("intelligence.multi_agent.agents[%s].weight", agent.Role), "must be between 0.0 and 1.0")
			}
			if agent.Role != "critic" {
				totalWeight += agent.Weight
			}
		}

		// Check that weights sum to approximately 1.0 (with some tolerance)
		if totalWeight < 0.95 || totalWeight > 1.05 {
			ve.Add("intelligence.multi_agent.agents", fmt.Sprintf("agent weights should sum to ~1.0, got %.2f", totalWeight))
		}
	}

	// Validate LLM config
	if len(cfg.LLM.Providers) == 0 {
		ve.Add("llm.providers", "at least one provider must be configured")
	}

	for _, provider := range cfg.LLM.Providers {
		if provider.Name == "" {
			ve.Add("llm.providers", "provider name cannot be empty")
		}
		if provider.Model == "" {
			ve.Add(fmt.Sprintf("llm.providers[%s].model", provider.Name), "model cannot be empty")
		}
	}

	// Validate performance config
	if cfg.Performance.Concurrency < 1 || cfg.Performance.Concurrency > 100 {
		ve.Add("performance.concurrency", "must be between 1 and 100")
	}
	if cfg.Performance.TimeoutSeconds < 10 || cfg.Performance.TimeoutSeconds > 3600 {
		ve.Add("performance.timeout_seconds", "must be between 10 and 3600")
	}

	if ve.HasErrors() {
		return ve
	}

	return nil
}

// ValidateConsolidatedDocs validates and sanitizes the consolidated docs
func ValidateConsolidatedDocs(docs *ConsolidatedDocs) error {
	ve := &ValidationErrors{}

	if docs == nil {
		return fmt.Errorf("consolidated docs is nil")
	}

	// Sanitize module name to prevent injection attacks
	docs.ModuleName = validation.SanitizeString(docs.ModuleName)
	if docs.ModuleName == "" {
		ve.Add("module_name", "cannot be empty")
	}

	// Sanitize source commit
	docs.SourceCommit = validation.SanitizeString(docs.SourceCommit)

	if len(docs.Services) == 0 && len(docs.Messages) == 0 {
		ve.Add("services/messages", "at least one service or message must be present")
	}

	// Validate and sanitize services
	for i := range docs.Services {
		// Sanitize service names to prevent injection
		docs.Services[i].Name = validation.SanitizeString(docs.Services[i].Name)
		docs.Services[i].Description = validation.SanitizeString(docs.Services[i].Description)

		if docs.Services[i].Name == "" {
			ve.Add(fmt.Sprintf("services[%d].name", i), "cannot be empty")
		}
		if len(docs.Services[i].Methods) == 0 {
			ve.Add(fmt.Sprintf("services[%d].methods", i), "service must have at least one method")
		}

		// Validate and sanitize methods
		for j := range docs.Services[i].Methods {
			// Sanitize method fields
			docs.Services[i].Methods[j].Name = validation.SanitizeString(docs.Services[i].Methods[j].Name)
			docs.Services[i].Methods[j].Description = validation.SanitizeString(docs.Services[i].Methods[j].Description)
			docs.Services[i].Methods[j].InputType = validation.SanitizeString(docs.Services[i].Methods[j].InputType)
			docs.Services[i].Methods[j].OutputType = validation.SanitizeString(docs.Services[i].Methods[j].OutputType)

			if docs.Services[i].Methods[j].Name == "" {
				ve.Add(fmt.Sprintf("services[%d].methods[%d].name", i, j), "cannot be empty")
			}
			if docs.Services[i].Methods[j].InputType == "" {
				ve.Add(fmt.Sprintf("services[%d].methods[%d].input_type", i, j), "cannot be empty")
			}
			if docs.Services[i].Methods[j].OutputType == "" {
				ve.Add(fmt.Sprintf("services[%d].methods[%d].output_type", i, j), "cannot be empty")
			}
		}
	}

	// Validate and sanitize messages
	for i := range docs.Messages {
		// Sanitize message fields
		docs.Messages[i].Name = validation.SanitizeString(docs.Messages[i].Name)
		docs.Messages[i].Description = validation.SanitizeString(docs.Messages[i].Description)

		if docs.Messages[i].Name == "" {
			ve.Add(fmt.Sprintf("messages[%d].name", i), "cannot be empty")
		}
		// Fields can be empty for empty messages
	}

	if ve.HasErrors() {
		return ve
	}

	return nil
}

// ValidateHLDOutput validates the HLD output
func ValidateHLDOutput(output *HLDOutput) error {
	ve := &ValidationErrors{}

	if output == nil {
		return fmt.Errorf("HLD output is nil")
	}

	if output.Architecture == "" {
		ve.Add("architecture", "cannot be empty")
	}

	if output.Markdown == "" {
		ve.Add("markdown", "cannot be empty")
	}

	if output.ConsensusScore < 0.0 || output.ConsensusScore > 1.0 {
		ve.Add("consensus_score", "must be between 0.0 and 1.0")
	}

	if output.FinalScore < 0.0 || output.FinalScore > 1.0 {
		ve.Add("final_score", "must be between 0.0 and 1.0")
	}

	if output.RoundsCompleted < 0 {
		ve.Add("rounds_completed", "cannot be negative")
	}

	if output.Metadata.ModuleName == "" {
		ve.Add("metadata.module_name", "cannot be empty")
	}

	if ve.HasErrors() {
		return ve
	}

	return nil
}

// ValidateAgentResponse validates an agent response
func ValidateAgentResponse(resp *AgentResponse) error {
	ve := &ValidationErrors{}

	if resp == nil {
		return fmt.Errorf("agent response is nil")
	}

	if resp.Role == "" {
		ve.Add("role", "cannot be empty")
	}

	if resp.Content == "" {
		ve.Add("content", "cannot be empty")
	}

	if resp.Confidence < 0.0 || resp.Confidence > 1.0 {
		ve.Add("confidence", "must be between 0.0 and 1.0")
	}

	if resp.TokensUsed < 0 {
		ve.Add("tokens_used", "cannot be negative")
	}

	if ve.HasErrors() {
		return ve
	}

	return nil
}

// ValidateCriticism validates a criticism object
func ValidateCriticism(crit *Criticism) error {
	ve := &ValidationErrors{}

	if crit.Agent == "" {
		ve.Add("agent", "cannot be empty")
	}

	if crit.Score < 0.0 || crit.Score > 1.0 {
		ve.Add("score", "must be between 0.0 and 1.0")
	}

	if crit.Confidence < 0.0 || crit.Confidence > 1.0 {
		ve.Add("confidence", "must be between 0.0 and 1.0")
	}

	// Validate issues
	for i, issue := range crit.Issues {
		if issue.Type == "" {
			ve.Add(fmt.Sprintf("issues[%d].type", i), "cannot be empty")
		}
		if issue.Severity == "" {
			ve.Add(fmt.Sprintf("issues[%d].severity", i), "cannot be empty")
		}
		if issue.Message == "" {
			ve.Add(fmt.Sprintf("issues[%d].message", i), "cannot be empty")
		}
	}

	if ve.HasErrors() {
		return ve
	}

	return nil
}

// SanitizeInput sanitizes user input to prevent injection attacks
func SanitizeInput(input string) string {
	// Remove potentially dangerous characters
	input = strings.ReplaceAll(input, "\x00", "")
	input = strings.ReplaceAll(input, "\r\n", "\n")

	// Limit length
	maxLen := 10000
	if len(input) > maxLen {
		input = input[:maxLen]
	}

	return input
}

// ValidateMode validates the generation mode
func ValidateMode(mode Mode) error {
	validModes := map[Mode]bool{
		ModeBasic:         true,
		ModeAdvanced:      true,
		ModeBusiness:      true,
		ModeCompliance:    true,
		ModeUltraAdvanced: true,
		ModeCustom:        true,
	}

	if !validModes[mode] {
		return fmt.Errorf("invalid mode: %s", mode)
	}

	return nil
}
