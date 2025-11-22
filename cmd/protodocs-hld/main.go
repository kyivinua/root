package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kyivinua/root/tools/protodocs/enricher"
	"github.com/kyivinua/root/tools/protodocs/hldgen"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	configPath   string
	inputModel   string
	outputPath   string
	moduleName   string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "protodocs-hld",
		Short: "Generate High-Level Design documentation from ProtoDocs",
		Long: `ProtoDocs HLD Generator v7.0

Enterprise-grade High-Level Design documentation generator using multi-agent AI system.

Features:
- Multi-agent reasoning (Architect, PM, Security, SRE, QA, Critic)
- Iterative refinement with consensus-based validation
- Structured output in Markdown, HTML, Confluence, PDF
- Full observability and audit trail`,
		RunE: runHLDGeneration,
	}

	rootCmd.Flags().StringVarP(&configPath, "config", "c", "configs/hld_generator.yaml", "Configuration file path")
	rootCmd.Flags().StringVarP(&inputModel, "input", "i", "api-docs/model/api-doc-model-enriched.json", "Input API doc model JSON")
	rootCmd.Flags().StringVarP(&outputPath, "output", "o", "api-docs/hld/", "Output directory")
	rootCmd.Flags().StringVarP(&moduleName, "module", "m", "", "Module name (optional, overrides model)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runHLDGeneration(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Setup logger
	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("component", "hld-generator").
		Logger()

	logger.Info().Msg("Starting HLD generation")

	// Load configuration
	logger.Info().Str("config", configPath).Msg("Loading configuration")
	cfg, err := hldgen.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if !cfg.Enabled {
		logger.Info().Msg("HLD Generator is disabled in configuration")
		return nil
	}

	// Load input model
	logger.Info().Str("input", inputModel).Msg("Loading input model")
	docs, err := loadConsolidatedDocs(inputModel)
	if err != nil {
		return fmt.Errorf("load input model: %w", err)
	}

	if moduleName != "" {
		docs.ModuleName = moduleName
	}

	logger.Info().
		Str("module", docs.ModuleName).
		Int("services", len(docs.Services)).
		Int("messages", len(docs.Messages)).
		Msg("Input model loaded")

	// Generate HLD
	logger.Info().Msg("Generating HLD with multi-agent system")
	output, err := hldgen.Generate(ctx, docs, *cfg, logger)
	if err != nil {
		return fmt.Errorf("HLD generation failed: %w", err)
	}

	logger.Info().
		Float64("consensus_score", output.ConsensusScore).
		Float64("final_score", output.FinalScore).
		Int("rounds", output.RoundsCompleted).
		Msg("HLD generation completed")

	// Save output
	if err := saveHLDOutput(output, outputPath, docs.ModuleName); err != nil {
		return fmt.Errorf("save output: %w", err)
	}

	logger.Info().
		Str("output_dir", outputPath).
		Msg("HLD saved successfully")

	// Print summary
	printSummary(output)

	return nil
}

// loadConsolidatedDocs loads the API doc model from enricher output
func loadConsolidatedDocs(path string) (*hldgen.ConsolidatedDocs, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var model enricher.ApiDocModel
	if err := json.Unmarshal(data, &model); err != nil {
		return nil, fmt.Errorf("unmarshal JSON: %w", err)
	}

	// Extract services and messages from modules
	docs := &hldgen.ConsolidatedDocs{
		Statistics:   model.Statistics,
		SourceCommit: model.SourceCommit,
		Services:     make([]hldgen.ServiceDoc, 0),
		Messages:     make([]hldgen.MessageDoc, 0),
		Enums:        make([]hldgen.EnumDoc, 0),
	}

	// Flatten modules
	for _, module := range model.Modules {
		if docs.ModuleName == "" {
			docs.ModuleName = module.PackageName
		}
		for _, svc := range module.Services {
			docs.Services = append(docs.Services, convertServiceDoc(svc))
		}
		for _, msg := range module.Messages {
			docs.Messages = append(docs.Messages, convertMessageDoc(msg))
		}
		for _, enum := range module.Enums {
			docs.Enums = append(docs.Enums, convertEnumDoc(enum))
		}
	}

	return docs, nil
}

// saveHLDOutput saves the HLD output
func saveHLDOutput(output *hldgen.HLDOutput, outputDir, moduleName string) error {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	// Save markdown
	markdownPath := filepath.Join(outputDir, fmt.Sprintf("%s-hld.md", moduleName))
	if err := os.WriteFile(markdownPath, []byte(output.Markdown), 0644); err != nil {
		return fmt.Errorf("write markdown: %w", err)
	}

	// Save JSON
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	jsonPath := filepath.Join(outputDir, fmt.Sprintf("%s-hld.json", moduleName))
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return fmt.Errorf("write JSON: %w", err)
	}

	return nil
}

// printSummary prints a summary of the generation
func printSummary(output *hldgen.HLDOutput) {
	fmt.Println("\n=== HLD Generation Summary ===")
	fmt.Printf("Module: %s\n", output.Metadata.ModuleName)
	fmt.Printf("Version: %s\n", output.Metadata.Version)
	fmt.Printf("Mode: %s\n", output.Metadata.GenerationMode)
	fmt.Printf("\nQuality Scores:\n")
	fmt.Printf("  Consensus Score: %.2f\n", output.ConsensusScore)
	fmt.Printf("  Final Score: %.2f\n", output.FinalScore)
	fmt.Printf("  Refinement Rounds: %d\n", output.RoundsCompleted)
	fmt.Printf("\nContent Sections:\n")
	fmt.Printf("  Business Context: %s\n", checkContent(output.BusinessContext))
	fmt.Printf("  Architecture: %s\n", checkContent(output.Architecture))
	fmt.Printf("  Security: %s\n", checkContent(output.Security))
	fmt.Printf("  Observability: %s\n", checkContent(output.Observability))
	fmt.Printf("  Requirements: %s\n", checkContent(output.Requirements))
	fmt.Printf("\nDiagrams: %d\n", len(output.Diagrams))
	fmt.Printf("SLO: %d\n", len(output.SLO))
	if len(output.Warnings) > 0 {
		fmt.Printf("\nWarnings: %d\n", len(output.Warnings))
		for _, warning := range output.Warnings {
			fmt.Printf("  - %s\n", warning)
		}
	}
	fmt.Println("============================\n")
}

func checkContent(content string) string {
	if content == "" {
		return "❌ Not Generated"
	}
	return "✓ Generated"
}

// Helper conversion functions to map enricher types to hldgen types
func convertServiceDoc(svc enricher.ServiceDoc) hldgen.ServiceDoc {
	methods := make([]hldgen.MethodDoc, len(svc.Methods))
	for i, m := range svc.Methods {
		methods[i] = hldgen.MethodDoc{
			Name:            m.Name,
			FullName:        m.FullName,
			Description:     m.Summary,
			InputType:       m.InputType,
			OutputType:      m.OutputType,
			ClientStreaming: m.IsStreaming, // Simplified - enricher doesn't distinguish client/server
			ServerStreaming: m.IsStreaming,
			HTTPMethods:     []string{},
		}
	}

	return hldgen.ServiceDoc{
		Name:        svc.Name,
		FullName:    svc.FullName,
		Description: svc.Summary,
		Methods:     methods,
		Visibility:  "public", // Default visibility
	}
}

func convertMessageDoc(msg enricher.MessageDoc) hldgen.MessageDoc {
	fields := make([]hldgen.FieldDoc, len(msg.Fields))
	for i, f := range msg.Fields {
		label := "optional"
		if f.Repeated {
			label = "repeated"
		}
		fields[i] = hldgen.FieldDoc{
			Name:        f.Name,
			Type:        f.Type,
			TypeName:    "", // Not available in enricher format
			Label:       label,
			Description: f.Summary,
		}
	}

	return hldgen.MessageDoc{
		Name:        msg.Name,
		FullName:    msg.FullName,
		Description: msg.Summary,
		Fields:      fields,
	}
}

func convertEnumDoc(enum enricher.EnumDoc) hldgen.EnumDoc {
	values := make([]hldgen.EnumValueDoc, len(enum.Values))
	for i, v := range enum.Values {
		values[i] = hldgen.EnumValueDoc{
			Name:        v.Name,
			Number:      int32(v.Number),
			Description: v.Summary,
		}
	}

	return hldgen.EnumDoc{
		Name:        enum.Name,
		FullName:    enum.FullName,
		Description: enum.Summary,
		Values:      values,
	}
}
