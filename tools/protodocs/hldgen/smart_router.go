package hldgen

import (
	"context"
	"fmt"
	"strings"
)

// SmartRouter intelligently routes LLM requests based on complexity and cost
type SmartRouter struct {
	providers   map[string]LLMClient
	strategy    string
	gpuMonitor  *GPUMonitor
	fallbackChain []string
}

// TaskComplexity represents the complexity level of a task
type TaskComplexity string

const (
	ComplexitySimple  TaskComplexity = "simple"
	ComplexityMedium  TaskComplexity = "medium"
	ComplexityComplex TaskComplexity = "complex"
)

// RouterStrategy represents routing strategy
type RouterStrategy string

const (
	StrategyCostFirst    RouterStrategy = "cost_first"
	StrategyQualityFirst RouterStrategy = "quality_first"
	StrategyBalanced     RouterStrategy = "balanced"
)

// NewSmartRouter creates a new smart router
func NewSmartRouter(providers map[string]LLMClient, strategy string, gpuMonitor *GPUMonitor, fallbackChain []string) *SmartRouter {
	if strategy == "" {
		strategy = "balanced"
	}

	return &SmartRouter{
		providers:     providers,
		strategy:      strategy,
		gpuMonitor:    gpuMonitor,
		fallbackChain: fallbackChain,
	}
}

// Route selects the best provider for the given task
func (sr *SmartRouter) Route(ctx context.Context, prompt string, config LLMConfig) (LLMClient, error) {
	// Classify task complexity
	complexity := sr.classifyComplexity(prompt)

	// Select provider based on strategy and complexity
	providerName := sr.selectProvider(complexity)

	// Get the client
	client, ok := sr.providers[providerName]
	if !ok {
		// Fallback to first available provider
		if len(sr.fallbackChain) > 0 {
			for _, name := range sr.fallbackChain {
				if client, ok = sr.providers[name]; ok {
					return client, nil
				}
			}
		}
		return nil, fmt.Errorf("no provider available")
	}

	return client, nil
}

// classifyComplexity determines the complexity level of a task
func (sr *SmartRouter) classifyComplexity(prompt string) TaskComplexity {
	promptLower := strings.ToLower(prompt)

	// Complex indicators
	complexIndicators := []string{
		"architecture",
		"distributed system",
		"microservices",
		"scalability",
		"multi-region",
		"critical path",
		"security analysis",
		"compliance",
		"threat model",
	}

	// Simple indicators
	simpleIndicators := []string{
		"summary",
		"list",
		"what is",
		"define",
		"explain briefly",
	}

	// Check for complex indicators
	complexCount := 0
	for _, indicator := range complexIndicators {
		if strings.Contains(promptLower, indicator) {
			complexCount++
		}
	}

	// Check for simple indicators
	simpleCount := 0
	for _, indicator := range simpleIndicators {
		if strings.Contains(promptLower, indicator) {
			simpleCount++
		}
	}

	// Determine complexity
	if complexCount >= 2 {
		return ComplexityComplex
	}
	if simpleCount >= 1 && complexCount == 0 {
		return ComplexitySimple
	}

	// Also consider prompt length
	if len(prompt) > 10000 {
		return ComplexityComplex
	}
	if len(prompt) < 1000 {
		return ComplexitySimple
	}

	return ComplexityMedium
}

// selectProvider chooses the best provider based on strategy and complexity
func (sr *SmartRouter) selectProvider(complexity TaskComplexity) string {
	strategy := RouterStrategy(sr.strategy)

	switch strategy {
	case StrategyCostFirst:
		return sr.selectCostFirst(complexity)
	case StrategyQualityFirst:
		return sr.selectQualityFirst(complexity)
	case StrategyBalanced:
		return sr.selectBalanced(complexity)
	default:
		return sr.selectBalanced(complexity)
	}
}

// selectCostFirst prioritizes cost efficiency
func (sr *SmartRouter) selectCostFirst(complexity TaskComplexity) string {
	// Always try Ollama first if GPU is available
	if sr.gpuMonitor != nil && sr.gpuMonitor.ShouldUseOllama() {
		if _, ok := sr.providers["ollama"]; ok {
			return "ollama"
		}
	}

	// Cost ranking: Ollama (free) > Gemini ($1.25/1M) > Anthropic ($3/1M) > OpenAI ($10/1M)
	switch complexity {
	case ComplexitySimple:
		// Simple tasks: Use cheapest available
		if _, ok := sr.providers["gemini"]; ok {
			return "gemini"
		}
		if _, ok := sr.providers["google"]; ok {
			return "google"
		}
		return "anthropic" // Fallback

	case ComplexityMedium:
		// Medium tasks: Gemini is good enough
		if _, ok := sr.providers["gemini"]; ok {
			return "gemini"
		}
		if _, ok := sr.providers["google"]; ok {
			return "google"
		}
		return "anthropic"

	case ComplexityComplex:
		// Complex tasks: Use Gemini unless quality is critical
		if _, ok := sr.providers["gemini"]; ok {
			return "gemini"
		}
		if _, ok := sr.providers["google"]; ok {
			return "google"
		}
		return "anthropic"
	}

	return "anthropic"
}

// selectQualityFirst prioritizes output quality
func (sr *SmartRouter) selectQualityFirst(complexity TaskComplexity) string {
	// Quality ranking: Claude > GPT-4 > Gemini > Ollama

	switch complexity {
	case ComplexitySimple:
		// Simple tasks: Even Ollama/Gemini is fine
		if sr.gpuMonitor != nil && sr.gpuMonitor.ShouldUseOllama() {
			if _, ok := sr.providers["ollama"]; ok {
				return "ollama"
			}
		}
		if _, ok := sr.providers["gemini"]; ok {
			return "gemini"
		}
		if _, ok := sr.providers["google"]; ok {
			return "google"
		}
		return "anthropic"

	case ComplexityMedium:
		// Medium tasks: Prefer Claude or Gemini
		if _, ok := sr.providers["anthropic"]; ok {
			return "anthropic"
		}
		if _, ok := sr.providers["gemini"]; ok {
			return "gemini"
		}
		if _, ok := sr.providers["google"]; ok {
			return "google"
		}
		return "openai"

	case ComplexityComplex:
		// Complex tasks: Always use Claude for best quality
		if _, ok := sr.providers["anthropic"]; ok {
			return "anthropic"
		}
		if _, ok := sr.providers["openai"]; ok {
			return "openai"
		}
		if _, ok := sr.providers["gemini"]; ok {
			return "gemini"
		}
		if _, ok := sr.providers["google"]; ok {
			return "google"
		}
		// Last resort: Ollama
		return "ollama"
	}

	return "anthropic"
}

// selectBalanced balances cost and quality
func (sr *SmartRouter) selectBalanced(complexity TaskComplexity) string {
	// Balanced approach: Use appropriate model for each complexity

	switch complexity {
	case ComplexitySimple:
		// Simple tasks: Optimize for cost
		if sr.gpuMonitor != nil && sr.gpuMonitor.ShouldUseOllama() {
			if _, ok := sr.providers["ollama"]; ok {
				return "ollama"
			}
		}
		if _, ok := sr.providers["gemini"]; ok {
			return "gemini"
		}
		if _, ok := sr.providers["google"]; ok {
			return "google"
		}
		return "anthropic"

	case ComplexityMedium:
		// Medium tasks: Gemini offers good balance
		if _, ok := sr.providers["gemini"]; ok {
			return "gemini"
		}
		if _, ok := sr.providers["google"]; ok {
			return "google"
		}
		if _, ok := sr.providers["anthropic"]; ok {
			return "anthropic"
		}
		return "openai"

	case ComplexityComplex:
		// Complex tasks: Use Claude for quality
		if _, ok := sr.providers["anthropic"]; ok {
			return "anthropic"
		}
		if _, ok := sr.providers["openai"]; ok {
			return "openai"
		}
		if _, ok := sr.providers["gemini"]; ok {
			return "gemini"
		}
		if _, ok := sr.providers["google"]; ok {
			return "google"
		}
		return "ollama"
	}

	return "anthropic"
}

// SmartLLMClient wraps the router to provide LLMClient interface
type SmartLLMClient struct {
	router *SmartRouter
}

// NewSmartLLMClient creates a new smart LLM client
func NewSmartLLMClient(router *SmartRouter) *SmartLLMClient {
	return &SmartLLMClient{
		router: router,
	}
}

// Generate implements LLMClient.Generate with smart routing
func (s *SmartLLMClient) Generate(ctx context.Context, prompt string, config LLMConfig) (*LLMResponse, error) {
	// Route to appropriate provider
	client, err := s.router.Route(ctx, prompt, config)
	if err != nil {
		return nil, fmt.Errorf("routing failed: %w", err)
	}

	// Generate with selected provider
	resp, err := client.Generate(ctx, prompt, config)
	if err != nil {
		// Try fallback chain on error
		if s.router.fallbackChain != nil {
			for _, providerName := range s.router.fallbackChain {
				fallbackClient, ok := s.router.providers[providerName]
				if !ok {
					continue
				}
				resp, err = fallbackClient.Generate(ctx, prompt, config)
				if err == nil {
					return resp, nil
				}
			}
		}
		return nil, err
	}

	return resp, nil
}

// GetModelName implements LLMClient.GetModelName
func (s *SmartLLMClient) GetModelName() string {
	return "smart-router"
}

// GetProviderName implements LLMClient.GetProviderName
func (s *SmartLLMClient) GetProviderName() string {
	return "smart-router"
}
