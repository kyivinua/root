// Package main is the entry point for docgen-tool CLI.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/kyivinua/docgen-tool/internal/config"
	"github.com/kyivinua/docgen-tool/internal/generator"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	version = "1.0.0"
	cfgFile string
	aiEnabled bool
	autoFix bool
	verbose bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "docgen",
	Short: "Protocol Buffer Documentation Generator",
	Long: `docgen-tool is an enterprise-grade documentation generator for Protocol Buffer services.

Features:
  - AI-powered enrichment with Claude
  - Quality validation and auto-fix
  - Mermaid diagram generation
  - Configurable templates
  - Parallel processing

For more information, visit: https://github.com/kyivinua/docgen-tool`,
	Version: version,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new configuration file",
	Long:  "Creates a default configuration file with example settings.",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputPath := cfgFile
		if outputPath == "" {
			outputPath = "./docgen.yaml"
		}

		if err := config.InitConfig(outputPath); err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}

		fmt.Printf("✅ Configuration file created: %s\n", outputPath)
		fmt.Println("\nNext steps:")
		fmt.Println("1. Edit the configuration file to match your project")
		fmt.Println("2. Run: docgen generate --config ./docgen.yaml")
		return nil
	},
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate documentation",
	Long:  "Generates documentation for Protocol Buffer services based on configuration.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		if cfgFile == "" {
			cfgFile = "./docgen.yaml"
		}

		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Override config with command-line flags
		if aiEnabled {
			cfg.Enricher.Enabled = true
		}
		if autoFix {
			cfg.Quality.AutoFix = true
		}

		// Setup logger
		logger := setupLogger(cfg.Logging, verbose)

		logger.Info().
			Str("version", version).
			Str("project", cfg.ProjectName).
			Msg("Starting docgen-tool")

		// Create generator
		gen := generator.NewGenerator(cfg, logger)

		// Generate documentation
		startTime := time.Now()
		doc, err := gen.Generate()
		if err != nil {
			logger.Error().Err(err).Msg("Documentation generation failed")
			return err
		}

		duration := time.Since(startTime)

		// Print summary
		fmt.Println("\n✅ Documentation generated successfully!")
		fmt.Printf("\n📊 Summary:\n")
		fmt.Printf("  Services:           %d\n", len(doc.Services))
		fmt.Printf("  Coverage Score:     %.1f%%\n", doc.Quality.CoverageScore)
		fmt.Printf("  Quality Score:      %.1f%%\n", doc.Quality.DescriptionQuality)
		fmt.Printf("  Quality Gate:       %s\n", map[bool]string{true: "✅ Passed", false: "❌ Failed"}[doc.Quality.Passed])
		fmt.Printf("  Generation Time:    %s\n", duration.Round(time.Millisecond))
		fmt.Printf("  Output Directory:   %s\n", cfg.OutputDir)

		if len(doc.Quality.Issues) > 0 {
			fmt.Printf("\n⚠️  Quality Issues: %d\n", len(doc.Quality.Issues))
			for i, issue := range doc.Quality.Issues {
				if i >= 5 {
					fmt.Printf("  ... and %d more\n", len(doc.Quality.Issues)-5)
					break
				}
				fmt.Printf("  - [%s] %s\n", issue.Severity, issue.Message)
			}
		}

		logger.Info().
			Dur("duration", duration).
			Float64("coverage", doc.Quality.CoverageScore).
			Bool("passed", doc.Quality.Passed).
			Msg("Documentation generation complete")

		return nil
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration file",
	Long:  "Validates the configuration file without generating documentation.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgFile == "" {
			cfgFile = "./docgen.yaml"
		}

		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		fmt.Printf("✅ Configuration is valid\n\n")
		fmt.Printf("Project:         %s\n", cfg.ProjectName)
		fmt.Printf("Version:         %s\n", cfg.Version)
		fmt.Printf("Output Dir:      %s\n", cfg.OutputDir)
		fmt.Printf("Services:        %d\n", len(cfg.Services))
		fmt.Printf("Enricher:        %s\n", map[bool]string{true: "Enabled", false: "Disabled"}[cfg.Enricher.Enabled])
		fmt.Printf("Diagrams:        %s\n", map[bool]string{true: "Enabled", false: "Disabled"}[cfg.Diagrams.Enabled])

		return nil
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("docgen-tool version %s\n", version)
		fmt.Printf("Build date: 2025-01-28\n")
		fmt.Printf("Go version: 1.22+\n")
	},
}

func init() {
	// Root command flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path (default: ./docgen.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")

	// Generate command flags
	generateCmd.Flags().BoolVar(&aiEnabled, "ai", false, "enable AI-powered enrichment")
	generateCmd.Flags().BoolVar(&autoFix, "auto-fix", false, "enable auto-fix for quality issues")

	// Init command flags
	initCmd.Flags().StringVarP(&cfgFile, "output", "o", "./docgen.yaml", "output path for config file")

	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(versionCmd)
}

func setupLogger(cfg config.LoggingConfig, verbose bool) zerolog.Logger {
	// Set log level
	level := zerolog.InfoLevel
	if verbose {
		level = zerolog.DebugLevel
	} else {
		switch cfg.Level {
		case "debug":
			level = zerolog.DebugLevel
		case "info":
			level = zerolog.InfoLevel
		case "warn":
			level = zerolog.WarnLevel
		case "error":
			level = zerolog.ErrorLevel
		}
	}

	zerolog.SetGlobalLevel(level)

	// Configure output
	var output *os.File
	if cfg.OutputFile != "" {
		f, err := os.OpenFile(cfg.OutputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			output = f
		} else {
			output = os.Stderr
		}
	} else {
		output = os.Stderr
	}

	// Configure format
	if cfg.Format == "console" {
		return zerolog.New(zerolog.ConsoleWriter{Out: output}).
			With().
			Timestamp().
			Logger()
	}

	return zerolog.New(output).
		With().
		Timestamp().
		Logger()
}
