package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/kyivinua/docgen-tool/tools/protodocs/enricher"
	"github.com/kyivinua/docgen-tool/tools/protodocs/enricher/adapters"
	"github.com/spf13/cobra"
)

var (
	configPath   string
	inputModel   string
	outputModel  string
	manifestPath string
	tenant       string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "protodocs-enricher",
		Short: "Enrich Protocol Buffer documentation with LLM",
		Long: `Enterprise-grade LLM-based documentation enrichment for Protocol Buffer APIs.

Features:
- Multi-provider LLM support (Anthropic Claude, OpenAI, Ollama)
- RAG integration with Weaviate for context retrieval
- Semantic entropy + LLM-as-judge for hallucination detection
- PII/PCI-DSS safety checks
- Tenant isolation and policy enforcement
- Full audit trail and metrics`,
		RunE: runEnrichment,
	}

	rootCmd.Flags().StringVarP(&configPath, "config", "c", "configs/enricher.config.yaml", "Configuration file path")
	rootCmd.Flags().StringVarP(&inputModel, "input", "i", "api-docs/model/api-doc-model.json", "Input API doc model JSON")
	rootCmd.Flags().StringVarP(&outputModel, "output", "o", "api-docs/model/api-doc-model-enriched.json", "Output enriched model JSON")
	rootCmd.Flags().StringVarP(&manifestPath, "manifest", "m", "api-docs/enrichment-manifest.json", "Enrichment manifest output path")
	rootCmd.Flags().StringVarP(&tenant, "tenant", "t", "default", "Tenant ID for policy enforcement")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runEnrichment(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Load configuration
	fmt.Printf("Loading configuration from %s...\n", configPath)
	config, err := enricher.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Initialize components
	fmt.Println("Initializing enrichment components...")

	// LLM client
	llmClient, err := adapters.NewGollmLLMClient(
		config.Provider,
		config.Model,
		config.APIKey,
		config.BaseURL,
		config.Temperature,
		config.MaxTokens,
	)
	if err != nil {
		return fmt.Errorf("create LLM client: %w", err)
	}

	// RAG retriever
	var ragRetriever enricher.RAGRetriever
	if config.RAG.Enabled {
		ragRetriever, err = adapters.NewWeaviateRAGRetriever(
			config.RAG.Weaviate.Host,
			config.RAG.Weaviate.Scheme,
			config.RAG.Weaviate.ClassName,
			config.RAG.Weaviate.APIKey,
			config.RAG.EmbedModel,
			config.RAG.MinScore,
		)
		if err != nil {
			fmt.Printf("Warning: Failed to create RAG retriever: %v\n", err)
			fmt.Println("Continuing without RAG...")
			ragRetriever = adapters.NewNoOpRAGRetriever()
		}
	} else {
		ragRetriever = adapters.NewNoOpRAGRetriever()
	}

	// Cache
	var cache enricher.EnrichmentCache
	if config.Cache.Enabled {
		cache, err = adapters.NewRistrettoCache(config.Cache.MaxSize, config.Cache.NumCounters)
		if err != nil {
			fmt.Printf("Warning: Failed to create cache: %v\n", err)
			cache = adapters.NewNoOpCache()
		}
	} else {
		cache = adapters.NewNoOpCache()
	}

	// Safety guard
	var safetyGuard enricher.SafetyGuard
	if config.Safety.Enabled {
		safetyGuard = adapters.NewEntropyJudgeSafetyGuard(
			llmClient,
			config.Safety.JudgeModel,
			config.Safety.EntropyThreshold,
			config.Safety.ConfidenceThreshold,
			config.Safety.PIIChecks,
			config.Safety.PCIDSSChecks,
		)
	} else {
		safetyGuard = adapters.NewNoOpSafetyGuard()
	}

	// Policy engine
	var policyEngine enricher.PolicyEngine
	if config.Policy.Enabled {
		policyEngine, err = enricher.NewStaticPolicyEngine(
			config.Policy.ConfigPath,
			config.Policy.DefaultTenant,
		)
		if err != nil {
			fmt.Printf("Warning: Failed to create policy engine: %v\n", err)
			policyEngine = enricher.NewNoOpPolicyEngine()
		}
	} else {
		policyEngine = enricher.NewNoOpPolicyEngine()
	}

	// Metrics
	var metrics enricher.EnrichmentMetrics
	if config.Metrics.Enabled {
		metrics = adapters.NewPrometheusMetrics(
			config.Metrics.Namespace,
			config.Metrics.Subsystem,
		)
	} else {
		metrics = adapters.NewNoOpMetrics()
	}

	// Trace sink
	var traceSink enricher.TraceSink
	if config.Trace.Enabled && config.Trace.Type == "file" {
		traceSink, err = adapters.NewFileTraceSink(config.Trace.OutputPath)
		if err != nil {
			fmt.Printf("Warning: Failed to create trace sink: %v\n", err)
			traceSink = adapters.NewNoOpTraceSink()
		}
	} else {
		traceSink = adapters.NewNoOpTraceSink()
	}

	// Template engine
	templateEngine, err := adapters.NewCoTTemplateEngine(
		config.Templates.UseCoT,
		config.Templates.DefaultFormat,
	)
	if err != nil {
		return fmt.Errorf("create template engine: %w", err)
	}

	// Strategy
	var strategy enricher.SmartStrategy
	if config.RAG.Enabled && config.RAG.UseAdaptive {
		strategy = adapters.NewAdaptiveRAGStrategy(config.RAG.TopK)
	} else if config.RAG.Enabled {
		strategy = adapters.NewAlwaysRAGStrategy(config.RAG.TopK)
	} else {
		strategy = adapters.NewNeverRAGStrategy()
	}

	// Create enricher
	enricherInstance, err := enricher.NewEnricher(
		config,
		llmClient,
		ragRetriever,
		cache,
		safetyGuard,
		policyEngine,
		metrics,
		traceSink,
		templateEngine,
		strategy,
	)
	if err != nil {
		return fmt.Errorf("create enricher: %w", err)
	}
	defer func() {
		if err := enricherInstance.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close enricher: %v\n", err)
		}
	}()

	// Load input model
	fmt.Printf("Loading input model from %s...\n", inputModel)
	modelData, err := os.ReadFile(inputModel)
	if err != nil {
		return fmt.Errorf("read input model: %w", err)
	}

	var apiModel enricher.ApiDocModel
	if err := json.Unmarshal(modelData, &apiModel); err != nil {
		return fmt.Errorf("unmarshal model: %w", err)
	}

	fmt.Printf("Loaded model with %d modules\n", len(apiModel.Modules))

	// Enrich model
	fmt.Println("\nStarting enrichment...")
	startTime := time.Now()

	if err := enricherInstance.EnrichModel(ctx, &apiModel, tenant); err != nil {
		return fmt.Errorf("enrich model: %w", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("\nEnrichment completed in %s\n", duration)

	// Get and save manifest
	manifest := enricherInstance.GetManifest()
	manifest.SetModelInfo(config.Model, config.Provider)
	manifest.PrintSummary()

	if err := manifest.SaveToFile(manifestPath); err != nil {
		fmt.Printf("Warning: Failed to save manifest: %v\n", err)
	} else {
		fmt.Printf("\nManifest saved to %s\n", manifestPath)
	}

	// Save enriched model
	fmt.Printf("Saving enriched model to %s...\n", outputModel)
	apiModel.GeneratedAt = time.Now()
	if apiModel.Tools == nil {
		apiModel.Tools = make(map[string]string)
	}
	apiModel.Tools["enricher"] = fmt.Sprintf("%s/%s", config.Provider, config.Model)

	enrichedData, err := json.MarshalIndent(apiModel, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal enriched model: %w", err)
	}

	if err := os.WriteFile(outputModel, enrichedData, 0644); err != nil {
		return fmt.Errorf("write enriched model: %w", err)
	}

	fmt.Println("\n✓ Enrichment complete!")

	// Check for failures
	if manifest.HasFailures() {
		fmt.Printf("\nWarning: %d targets failed enrichment\n", manifest.Statistics.FailedTargets)
		failedTargets := manifest.GetFailedTargets()
		if len(failedTargets) <= 10 {
			fmt.Println("Failed targets:")
			for _, target := range failedTargets {
				fmt.Printf("  - %s\n", target)
			}
		}
		return fmt.Errorf("enrichment completed with failures")
	}

	return nil
}
