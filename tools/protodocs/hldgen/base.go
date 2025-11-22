package hldgen

import (
	"context"

)

// Agent represents a specialized agent in the multi-agent system
type Agent interface {
	// Role returns the agent's role
	Role() AgentRole

	// Think generates a response based on the input
	Think(ctx context.Context, input *AgentInput) (*AgentResponse, error)

	// Weight returns the agent's weight in consensus calculation
	Weight() float64

	// SetLLMClient sets the LLM client for the agent
	SetLLMClient(client LLMClient)
}

// BaseAgent provides common functionality for all agents
type BaseAgent struct {
	role        AgentRole
	weight      float64
	model       string
	temperature float64
	maxTokens   int
	llmClient   LLMClient
}

// Role returns the agent's role
func (b *BaseAgent) Role() AgentRole {
	return b.role
}

// Weight returns the agent's weight
func (b *BaseAgent) Weight() float64 {
	return b.weight
}

// SetLLMClient sets the LLM client for the agent
func (b *BaseAgent) SetLLMClient(client LLMClient) {
	b.llmClient = client
}

// GetLLMClient returns the LLM client
func (b *BaseAgent) GetLLMClient() LLMClient {
	return b.llmClient
}
