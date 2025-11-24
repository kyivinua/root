package hldgen

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	serviceName = "hldgen"
	tracerName  = "github.com/kyivinua/docgen-tool/tools/protodocs/hldgen"
)

// InitTracing initializes OpenTelemetry tracing
// Uses TracingConfig from config.go
func InitTracing(cfg TracingConfig) (func(context.Context) error, error) {
	if !cfg.Enabled {
		// Return a no-op shutdown function
		return func(ctx context.Context) error { return nil }, nil
	}

	// Set default endpoint
	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = "localhost:4318" // Default OTLP HTTP endpoint
	}

	// Parse sampling rate from string (e.g., "1.0", "0.1", "always", "never")
	samplingRate := parseSamplingRate(cfg.Sampling)

	// Create OTLP HTTP exporter
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(), // Use HTTP (not HTTPS) for local development
	)

	// Use timeout context for initialization (10 seconds should be enough)
	initCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exporter, err := otlptrace.New(initCtx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.New(
		initCtx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider with sampling
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(samplingRate)),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Return shutdown function
	return tp.Shutdown, nil
}

// parseSamplingRate parses sampling configuration string to float64
func parseSamplingRate(sampling string) float64 {
	switch sampling {
	case "always", "":
		return 1.0 // Trace everything
	case "never":
		return 0.0 // Trace nothing
	default:
		// Try to parse as float (e.g., "0.1", "0.5")
		var rate float64
		if _, err := fmt.Sscanf(sampling, "%f", &rate); err == nil {
			if rate >= 0.0 && rate <= 1.0 {
				return rate
			}
		}
		return 1.0 // Default to always if invalid
	}
}

// GetTracer returns the global tracer
func GetTracer() trace.Tracer {
	return otel.Tracer(tracerName)
}

// TracedLLMClient wraps an LLMClient with OpenTelemetry tracing
type TracedLLMClient struct {
	client LLMClient
	tracer trace.Tracer
}

// NewTracedLLMClient creates a new traced LLM client
func NewTracedLLMClient(client LLMClient) *TracedLLMClient {
	return &TracedLLMClient{
		client: client,
		tracer: GetTracer(),
	}
}

// Generate implements LLMClient.Generate with tracing
func (t *TracedLLMClient) Generate(ctx context.Context, prompt string, config LLMConfig) (*LLMResponse, error) {
	// Start a span
	ctx, span := t.tracer.Start(ctx, "llm.generate",
		trace.WithAttributes(
			attribute.String("llm.provider", t.client.GetProviderName()),
			attribute.String("llm.model", t.client.GetModelName()),
			attribute.Int("llm.prompt_length", len(prompt)),
		),
	)
	defer span.End()

	// Record start time
	startTime := time.Now()

	// Call the underlying client
	resp, err := t.client.Generate(ctx, prompt, config)

	// Record duration
	duration := time.Since(startTime)
	span.SetAttributes(attribute.Float64("llm.duration_ms", float64(duration.Milliseconds())))

	if err != nil {
		// Record error
		span.RecordError(err)
		span.SetAttributes(attribute.Bool("llm.error", true))
		return nil, err
	}

	// Check if this was a cache hit
	cacheHit := resp.FinishReason == "cache_hit"

	// Record success metrics
	span.SetAttributes(
		attribute.Bool("llm.error", false),
		attribute.Int("llm.tokens_used", resp.TokensUsed),
		attribute.Int("llm.response_length", len(resp.Content)),
		attribute.Float64("llm.confidence", resp.Confidence),
		attribute.String("llm.finish_reason", resp.FinishReason),
		attribute.Bool("llm.cache_hit", cacheHit),
	)

	// Calculate and record cost (simplified cost model)
	// Cache hits have zero cost
	cost := 0.0
	if !cacheHit {
		cost = calculateCost(t.client.GetProviderName(), resp.TokensUsed)
	}
	span.SetAttributes(attribute.Float64("llm.cost_usd", cost))

	return resp, nil
}

// GetModelName implements LLMClient.GetModelName
func (t *TracedLLMClient) GetModelName() string {
	return t.client.GetModelName()
}

// GetProviderName implements LLMClient.GetProviderName
func (t *TracedLLMClient) GetProviderName() string {
	return t.client.GetProviderName()
}

// calculateCost calculates the cost of an LLM call in USD
func calculateCost(provider string, tokensUsed int) float64 {
	// Simplified cost model (input + output combined)
	// Real implementation should track input/output separately
	costPer1MTokens := 0.0

	switch provider {
	case "anthropic":
		costPer1MTokens = 3.0 // Claude 3.5 Sonnet average
	case "openai":
		costPer1MTokens = 10.0 // GPT-4 average
	case "google", "gemini":
		costPer1MTokens = 1.25 // Gemini 1.5 Pro
	case "ollama":
		costPer1MTokens = 0.0 // Free (local)
	default:
		costPer1MTokens = 5.0 // Default estimate
	}

	return (float64(tokensUsed) / 1_000_000.0) * costPer1MTokens
}

// WrapLLMClient wraps an LLM client with tracing if tracing is enabled
func WrapLLMClient(client LLMClient, tracingEnabled bool) LLMClient {
	if !tracingEnabled {
		return client
	}
	return NewTracedLLMClient(client)
}

// TracedAgentResponse adds tracing to agent responses
func TraceAgentThink(ctx context.Context, role AgentRole, fn func(context.Context) (*AgentResponse, error)) (*AgentResponse, error) {
	tracer := GetTracer()

	ctx, span := tracer.Start(ctx, fmt.Sprintf("agent.%s.think", role),
		trace.WithAttributes(
			attribute.String("agent.role", string(role)),
		),
	)
	defer span.End()

	startTime := time.Now()
	resp, err := fn(ctx)
	duration := time.Since(startTime)

	span.SetAttributes(attribute.Float64("agent.duration_ms", float64(duration.Milliseconds())))

	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Bool("agent.error", true))
		return nil, err
	}

	span.SetAttributes(
		attribute.Bool("agent.error", false),
		attribute.Int("agent.tokens_used", resp.TokensUsed),
		attribute.Int("agent.content_length", len(resp.Content)),
		attribute.Float64("agent.confidence", resp.Confidence),
		attribute.Int("agent.diagrams_count", len(resp.Diagrams)),
	)

	return resp, nil
}
