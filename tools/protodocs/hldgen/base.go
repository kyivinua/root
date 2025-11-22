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
}

// BaseAgent provides common functionality for all agents
type BaseAgent struct {
	role        AgentRole
	weight      float64
	model       string
	temperature float64
	maxTokens   int
}

// Role returns the agent's role
func (b *BaseAgent) Role() AgentRole {
	return b.role
}

// Weight returns the agent's weight
func (b *BaseAgent) Weight() float64 {
	return b.weight
}
