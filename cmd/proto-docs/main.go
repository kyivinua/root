package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kyivinua/docgen-tool/tools/protodocs/pipeline"
	"github.com/spf13/cobra"
)

var (
	configPath string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "proto-docs",
		Short: "Protocol Buffer Documentation Generator",
		Long: `proto-docs is a CLI tool for generating documentation from Protocol Buffer schemas.

It supports:
  - Automatic extraction of comments and structure from .proto files
  - Generation of Markdown/HTML documentation
  - OpenAPI specification generation
  - Schema compatibility checking
  - Static documentation site assembly`,
	}

	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "configs/proto-docs.config.yaml", "Path to config file")

	// Command: all
	allCmd := &cobra.Command{
		Use:   "all",
		Short: "Run the complete documentation pipeline",
		RunE:  runAll,
	}

	// Command: lint
	lintCmd := &cobra.Command{
		Use:   "lint",
		Short: "Run proto linting",
		RunE:  runLint,
	}

	// Command: breaking
	breakingCmd := &cobra.Command{
		Use:   "breaking",
		Short: "Check for breaking changes",
		RunE:  runBreaking,
	}

	// Command: build-desc
	buildDescCmd := &cobra.Command{
		Use:   "build-desc",
		Short: "Build descriptor set (image.bin)",
		RunE:  runBuildDesc,
	}

	// Command: model
	modelCmd := &cobra.Command{
		Use:   "model",
		Short: "Build documentation model",
		RunE:  runModel,
	}

	rootCmd.AddCommand(allCmd, lintCmd, breakingCmd, buildDescCmd, modelCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runAll(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	p := pipeline.NewPipeline(cfg)
	return p.RunAll()
}

func runLint(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	p := pipeline.NewPipeline(cfg)
	return p.RunLint()
}

func runBreaking(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	p := pipeline.NewPipeline(cfg)
	return p.RunBreaking()
}

func runBuildDesc(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	p := pipeline.NewPipeline(cfg)
	descPath, err := p.RunDescriptorBuild()
	if err != nil {
		return err
	}

	log.Printf("✓ Descriptor built successfully: %s\n", descPath)
	return nil
}

func runModel(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	p := pipeline.NewPipeline(cfg)
	model, err := p.RunDocModelBuild()
	if err != nil {
		return err
	}

	log.Printf("✓ Doc model built successfully: %d modules, %d services, %d messages\n",
		len(model.Modules),
		model.Statistics["total_services"],
		model.Statistics["total_messages"])
	return nil
}

func loadConfig() (*pipeline.PipelineConfig, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("Config file not found: %s, using defaults\n", configPath)
		return pipeline.DefaultConfig(), nil
	}

	cfg, err := pipeline.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return cfg, nil
}
