// Package generator provides the main documentation generation functionality.
package generator

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kyivinua/docgen-tool/internal/config"
	"github.com/kyivinua/docgen-tool/internal/diagrams"
	"github.com/kyivinua/docgen-tool/internal/docgen"
	"github.com/kyivinua/docgen-tool/internal/enricher"
	"github.com/kyivinua/docgen-tool/internal/fileutil"
	"github.com/kyivinua/docgen-tool/internal/validator"
	"github.com/rs/zerolog"
)

// Generator is the main documentation generator.
type Generator struct {
	config       *config.Config
	validator    *validator.Validator
	enricher     *enricher.Enricher
	diagramGen   *diagrams.Generator
	logger       zerolog.Logger
}

// NewGenerator creates a new documentation generator.
func NewGenerator(cfg *config.Config, logger zerolog.Logger) *Generator {
	validatorConfig := validator.Config{
		CoverageTarget:           cfg.Quality.CoverageTarget,
		DescriptionQualityTarget: cfg.Quality.DescriptionQualityTarget,
		MinMethodCoverage:        cfg.Quality.MinMethodCoverage,
		MinFieldCoverage:         cfg.Quality.MinFieldCoverage,
		AutoFix:                  cfg.Quality.AutoFix,
		FailOnQualityGate:        cfg.Quality.FailOnQualityGate,
	}

	enricherConfig := enricher.Config{
		Provider:         cfg.Enricher.Provider,
		APIKey:           cfg.Enricher.APIKey,
		Model:            cfg.Enricher.Model,
		MaxTokens:        cfg.Enricher.MaxTokens,
		Temperature:      cfg.Enricher.Temperature,
		CacheEnabled:     cfg.Enricher.CacheEnabled,
		RateLimitPerMin:  cfg.Enricher.RateLimitPerMin,
		ParallelRequests: cfg.Enricher.ParallelRequests,
	}

	diagramConfig := diagrams.Config{
		ServiceGraph:  cfg.Diagrams.ServiceGraph,
		MessageGraph:  cfg.Diagrams.MessageGraph,
		SequenceGraph: cfg.Diagrams.SequenceGraph,
		OverviewGraph: cfg.Diagrams.OverviewGraph,
	}

	return &Generator{
		config:     cfg,
		validator:  validator.NewValidator(validatorConfig),
		enricher:   enricher.NewEnricher(enricherConfig),
		diagramGen: diagrams.NewGenerator(diagramConfig),
		logger:     logger,
	}
}

// Generate generates documentation for all enabled services.
func (g *Generator) Generate() (*docgen.Documentation, error) {
	g.logger.Info().Msg("Starting documentation generation")

	enabledServices := g.config.GetEnabledServices()
	if len(enabledServices) == 0 {
		return nil, fmt.Errorf("no enabled services found")
	}

	// Ensure output directory exists
	if err := fileutil.EnsureDir(g.config.OutputDir); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate documentation for each service
	var services []docgen.Service
	var mu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, len(enabledServices))

	for _, svcConfig := range enabledServices {
		wg.Add(1)
		go func(svcCfg config.ServiceConfig) {
			defer wg.Done()

			g.logger.Info().Str("service", svcCfg.Name).Msg("Generating documentation for service")

			service, err := g.generateServiceDoc(svcCfg)
			if err != nil {
				errChan <- fmt.Errorf("failed to generate docs for %s: %w", svcCfg.Name, err)
				return
			}

			mu.Lock()
			services = append(services, *service)
			mu.Unlock()

			g.logger.Info().
				Str("service", svcCfg.Name).
				Int("methods", len(service.Methods)).
				Int("messages", len(service.Messages)).
				Msg("Service documentation generated")
		}(svcConfig)
	}

	wg.Wait()
	close(errChan)

	// Collect errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("generation errors: %v", errs)
	}

	// Validate quality
	g.logger.Info().Msg("Validating documentation quality")
	qualityReport, err := g.validator.ValidateAll(services)
	if err != nil {
		return nil, fmt.Errorf("quality validation failed: %w", err)
	}

	g.logger.Info().
		Float64("coverage", qualityReport.CoverageScore).
		Float64("quality", qualityReport.DescriptionQuality).
		Bool("passed", qualityReport.Passed).
		Msg("Quality validation complete")

	// Generate overview diagram
	var allDiagrams map[string]string
	if g.config.Diagrams.OverviewGraph {
		overview, err := g.diagramGen.GenerateOverviewDiagram(services)
		if err != nil {
			g.logger.Warn().Err(err).Msg("Failed to generate overview diagram")
		} else {
			allDiagrams = make(map[string]string)
			allDiagrams["overview"] = overview
		}
	}

	// Create documentation
	doc := &docgen.Documentation{
		Services: services,
		Metadata: docgen.DocumentMetadata{
			ProjectName: g.config.ProjectName,
			Version:     g.config.Version,
			Generated:   time.Now(),
			Generator:   "docgen-tool v1.0.0",
			Config: map[string]string{
				"output_dir": g.config.OutputDir,
			},
		},
		Quality:     *qualityReport,
		Diagrams:    allDiagrams,
		GeneratedAt: time.Now(),
	}

	// Write documentation files
	if err := g.writeDocumentation(doc); err != nil {
		return nil, fmt.Errorf("failed to write documentation: %w", err)
	}

	g.logger.Info().Msg("Documentation generation complete")
	return doc, nil
}

// generateServiceDoc generates documentation for a single service.
func (g *Generator) generateServiceDoc(svcConfig config.ServiceConfig) (*docgen.Service, error) {
	// Create service structure (in a real implementation, this would parse proto files)
	service := &docgen.Service{
		Name:        svcConfig.Name,
		Description: fmt.Sprintf("%s service", svcConfig.Name),
		Package:     svcConfig.Package,
		Version:     svcConfig.Version,
		ProtoFiles:  svcConfig.ProtoFiles,
		Generated:   time.Now(),
		Methods:     g.generateSampleMethods(svcConfig.Name),
		Messages:    g.generateSampleMessages(svcConfig.Name),
	}

	// Enrich with AI if enabled
	if g.config.Enricher.Enabled {
		g.logger.Info().Str("service", service.Name).Msg("Enriching documentation with AI")
		if err := g.enricher.Enrich(service); err != nil {
			g.logger.Warn().Err(err).Msg("AI enrichment failed, continuing with basic descriptions")
		}
	}

	// Validate and auto-fix
	quality, issues, err := g.validator.Validate(service)
	if err != nil {
		return nil, err
	}

	if g.config.Quality.AutoFix && len(issues) > 0 {
		g.logger.Info().Str("service", service.Name).Msg("Auto-fixing quality issues")
		if err := g.validator.AutoFix(service, issues); err != nil {
			g.logger.Warn().Err(err).Msg("Auto-fix failed")
		}
	}

	g.logger.Info().
		Str("service", service.Name).
		Float64("coverage", quality.Coverage).
		Float64("quality", quality.DescriptionQuality).
		Msg("Service validation complete")

	return service, nil
}

// generateSampleMethods generates sample methods for demonstration.
func (g *Generator) generateSampleMethods(serviceName string) []docgen.Method {
	return []docgen.Method{
		{
			Name:            "Create",
			Description:     fmt.Sprintf("Creates a new %s resource", serviceName),
			InputType:       fmt.Sprintf("Create%sRequest", serviceName),
			OutputType:      fmt.Sprintf("Create%sResponse", serviceName),
			ClientStreaming: false,
			ServerStreaming: false,
		},
		{
			Name:            "Get",
			Description:     fmt.Sprintf("Retrieves a %s resource by ID", serviceName),
			InputType:       fmt.Sprintf("Get%sRequest", serviceName),
			OutputType:      fmt.Sprintf("Get%sResponse", serviceName),
			ClientStreaming: false,
			ServerStreaming: false,
		},
		{
			Name:            "List",
			Description:     fmt.Sprintf("Lists %s resources with pagination", serviceName),
			InputType:       fmt.Sprintf("List%sRequest", serviceName),
			OutputType:      fmt.Sprintf("List%sResponse", serviceName),
			ClientStreaming: false,
			ServerStreaming: true,
		},
	}
}

// generateSampleMessages generates sample messages for demonstration.
func (g *Generator) generateSampleMessages(serviceName string) []docgen.Message {
	return []docgen.Message{
		{
			Name:        fmt.Sprintf("Create%sRequest", serviceName),
			Description: fmt.Sprintf("Request message for creating a %s", serviceName),
			Fields: []docgen.Field{
				{Name: "name", Type: "string", Description: "Resource name", Number: 1, Label: "optional"},
				{Name: "description", Type: "string", Description: "Resource description", Number: 2, Label: "optional"},
			},
		},
		{
			Name:        fmt.Sprintf("Create%sResponse", serviceName),
			Description: fmt.Sprintf("Response message for creating a %s", serviceName),
			Fields: []docgen.Field{
				{Name: "id", Type: "string", Description: "Created resource ID", Number: 1, Label: "optional"},
				{Name: "status", Type: "string", Description: "Creation status", Number: 2, Label: "optional"},
			},
		},
	}
}

// writeDocumentation writes documentation to files.
func (g *Generator) writeDocumentation(doc *docgen.Documentation) error {
	// Write main README
	readme := g.generateReadme(doc)
	readmePath := fmt.Sprintf("%s/README.md", g.config.OutputDir)
	if err := fileutil.WriteFile(readmePath, []byte(readme)); err != nil {
		return fmt.Errorf("failed to write README: %w", err)
	}

	// Write service documentation
	for _, service := range doc.Services {
		serviceDocs := g.generateServiceMarkdown(&service)
		servicePath := fmt.Sprintf("%s/%s.md", g.config.OutputDir, strings.ToLower(service.Name))
		if err := fileutil.WriteFile(servicePath, []byte(serviceDocs)); err != nil {
			return fmt.Errorf("failed to write service docs: %w", err)
		}

		// Generate and write diagrams
		if g.config.Diagrams.Enabled {
			diagrams, err := g.diagramGen.GenerateAllDiagrams(&service)
			if err != nil {
				g.logger.Warn().Err(err).Str("service", service.Name).Msg("Failed to generate diagrams")
			} else {
				for name, content := range diagrams {
					diagramPath := fmt.Sprintf("%s/diagrams/%s_%s.md", g.config.OutputDir, strings.ToLower(service.Name), name)
					if err := fileutil.WriteFile(diagramPath, []byte(content)); err != nil {
						g.logger.Warn().Err(err).Str("diagram", name).Msg("Failed to write diagram")
					}
				}
			}
		}
	}

	g.logger.Info().Str("output_dir", g.config.OutputDir).Msg("Documentation written successfully")
	return nil
}

// generateReadme generates the main README content.
func (g *Generator) generateReadme(doc *docgen.Documentation) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s Documentation\n\n", doc.Metadata.ProjectName))
	sb.WriteString(fmt.Sprintf("Generated: %s\n\n", doc.GeneratedAt.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("Generator: %s\n\n", doc.Metadata.Generator))

	sb.WriteString("## Services\n\n")
	for _, service := range doc.Services {
		sb.WriteString(fmt.Sprintf("- [%s](%s.md) - %s\n", service.Name, strings.ToLower(service.Name), service.Description))
	}

	sb.WriteString("\n## Quality Metrics\n\n")
	sb.WriteString(fmt.Sprintf("- Coverage Score: %.1f%%\n", doc.Quality.CoverageScore))
	sb.WriteString(fmt.Sprintf("- Description Quality: %.1f%%\n", doc.Quality.DescriptionQuality))
	sb.WriteString(fmt.Sprintf("- Method Coverage: %.1f%%\n", doc.Quality.MethodCoverage))
	sb.WriteString(fmt.Sprintf("- Field Coverage: %.1f%%\n", doc.Quality.FieldCoverage))
	sb.WriteString(fmt.Sprintf("- Quality Gate: %s\n", map[bool]string{true: "✅ Passed", false: "❌ Failed"}[doc.Quality.Passed]))

	if len(doc.Quality.Issues) > 0 {
		sb.WriteString("\n### Issues\n\n")
		for _, issue := range doc.Quality.Issues {
			sb.WriteString(fmt.Sprintf("- **%s** [%s] %s\n", strings.ToUpper(issue.Severity), issue.Location, issue.Message))
		}
	}

	return sb.String()
}

// generateServiceMarkdown generates markdown documentation for a service.
func (g *Generator) generateServiceMarkdown(service *docgen.Service) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", service.Name))
	sb.WriteString(fmt.Sprintf("%s\n\n", service.Description))
	sb.WriteString(fmt.Sprintf("**Package:** `%s`\n\n", service.Package))
	sb.WriteString(fmt.Sprintf("**Version:** %s\n\n", service.Version))

	sb.WriteString("## Methods\n\n")
	for _, method := range service.Methods {
		sb.WriteString(fmt.Sprintf("### %s\n\n", method.Name))
		sb.WriteString(fmt.Sprintf("%s\n\n", method.Description))
		sb.WriteString(fmt.Sprintf("**Input:** `%s`\n\n", method.InputType))
		sb.WriteString(fmt.Sprintf("**Output:** `%s`\n\n", method.OutputType))

		if method.ClientStreaming || method.ServerStreaming {
			sb.WriteString("**Streaming:**")
			if method.ClientStreaming {
				sb.WriteString(" Client")
			}
			if method.ServerStreaming {
				sb.WriteString(" Server")
			}
			sb.WriteString("\n\n")
		}

		if method.Deprecated {
			sb.WriteString("⚠️ **DEPRECATED**\n\n")
		}
	}

	sb.WriteString("## Messages\n\n")
	for _, message := range service.Messages {
		sb.WriteString(fmt.Sprintf("### %s\n\n", message.Name))
		sb.WriteString(fmt.Sprintf("%s\n\n", message.Description))

		if len(message.Fields) > 0 {
			sb.WriteString("| Field | Type | Description |\n")
			sb.WriteString("|-------|------|-------------|\n")
			for _, field := range message.Fields {
				fieldType := field.Type
				if field.Label == "repeated" {
					fieldType = "[]" + fieldType
				}
				sb.WriteString(fmt.Sprintf("| %s | `%s` | %s |\n", field.Name, fieldType, field.Description))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
