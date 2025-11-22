package pipeline

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kyivinua/docgen-tool/tools/protoctx"
	"github.com/kyivinua/docgen-tool/tools/protodocs/diagrams"
)

// Pipeline orchestrates the documentation generation pipeline.
type Pipeline struct {
	config              *PipelineConfig
	logger              *log.Logger
	notificationManager *NotificationManager
}

// NewPipeline creates a new pipeline with the given configuration.
func NewPipeline(config *PipelineConfig) *Pipeline {
	return &Pipeline{
		config:              config,
		logger:              log.New(os.Stdout, "[pipeline] ", log.LstdFlags),
		notificationManager: NewNotificationManager(&config.Notifications),
	}
}

// RunAll executes all pipeline stages in order.
//
// Stages:
//  1. Discovery - determine changed .proto files
//  2. Lint - buf lint + comment-policy check
//  3. Breaking Check - buf breaking
//  4. Descriptor Build - buf build → image.bin
//  5. Doc Model Build - build ApiDocModel
//  6. Docs Generation - generate Markdown/HTML
//  7. OpenAPI Generation - generate OpenAPI
//  8. Site Assembly - build static site
//
// Any error in any stage causes the pipeline to fail.
func (p *Pipeline) RunAll() error {
	p.logger.Println("Starting full documentation pipeline")

	// Get commit and branch info for notifications
	commit := getSourceCommit()
	branch := getCurrentBranch()

	// Notify pipeline start
	if p.notificationManager != nil {
		p.notificationManager.NotifyPipelineStart(commit, branch)
	}

	// Stage 0: Discovery
	scope, err := p.runDiscovery()
	if err != nil {
		if p.notificationManager != nil {
			p.notificationManager.NotifyPipelineComplete(false, nil, []string{err.Error()})
		}
		return fmt.Errorf("discovery: %w", err)
	}

	if scope.IsEmpty() {
		p.logger.Println("No proto files to process, skipping pipeline")
		return nil
	}

	p.logger.Printf("Scope: %s\n", scope)

	// Stage 1: Lint
	if err := p.runLint(); err != nil {
		return fmt.Errorf("lint: %w", err)
	}

	// Stage 2: Breaking Check
	if err := p.runBreaking(); err != nil {
		return fmt.Errorf("breaking: %w", err)
	}

	// Stage 3: Descriptor Build
	descPath, err := p.runDescriptorBuild()
	if err != nil {
		return fmt.Errorf("descriptor build: %w", err)
	}

	// Stage 4: Doc Model Build
	model, err := p.runDocModelBuild(descPath)
	if err != nil {
		return fmt.Errorf("doc model build: %w", err)
	}

	// Stage 4.5: Enrichment (optional)
	if p.config.Enrichment.Enabled {
		enrichedModel, err := p.runEnrichment(model)
		if err != nil {
			return fmt.Errorf("enrichment: %w", err)
		}
		model = enrichedModel

		// Notify enrichment completion
		if p.notificationManager != nil {
			p.notificationManager.NotifyEnrichmentComplete(p.config.Enrichment.ManifestPath)
		}
	}

	// Stage 4.7: Diagram Generation (optional)
	if p.config.Diagrams.Enabled {
		if err := p.runDiagramGeneration(model); err != nil {
			return fmt.Errorf("diagram generation: %w", err)
		}
	}

	// Stage 4.8: High-Level Design Generation (optional)
	if p.config.HLD.Enabled {
		if err := p.runHLDGeneration(model); err != nil {
			return fmt.Errorf("HLD generation: %w", err)
		}
	}

	// Stage 5: Docs Generation
	if err := p.runDocsGeneration(); err != nil {
		return fmt.Errorf("docs generation: %w", err)
	}

	// Stage 6: OpenAPI Generation
	if p.config.OpenAPI.Enabled {
		if err := p.runOpenAPIGeneration(); err != nil {
			return fmt.Errorf("openapi generation: %w", err)
		}
	}

	// Stage 7: Site Assembly
	if err := p.runSiteAssembly(); err != nil {
		return fmt.Errorf("site assembly: %w", err)
	}

	p.logger.Printf("Pipeline completed successfully. Model: %d modules, %d services, %d messages\n",
		len(model.Modules),
		model.Statistics["total_services"],
		model.Statistics["total_messages"])

	// Notify successful completion
	if p.notificationManager != nil {
		p.notificationManager.NotifyPipelineComplete(true, model, nil)

		// Send release notes if configured
		manifestPath := ""
		if p.config.Enrichment.Enabled {
			manifestPath = p.config.Enrichment.ManifestPath
		}
		p.notificationManager.NotifyReleaseNotes(model, manifestPath)
	}

	return nil
}

// runDiscovery executes the discovery stage.
func (p *Pipeline) runDiscovery() (*Scope, error) {
	p.logger.Println("Stage 0: Discovery")

	// For now, always discover all proto files
	// TODO: Implement incremental discovery based on git diff
	scope, err := DiscoverAllProto(p.config.ProtoRoot)
	if err != nil {
		return nil, err
	}

	return scope, nil
}

// runLint executes the lint stage.
func (p *Pipeline) runLint() error {
	p.logger.Println("Stage 1: Lint")

	if !p.config.Lint.EnableBufLint {
		p.logger.Println("Buf lint disabled, skipping")
		return nil
	}

	// Run buf lint
	cmd := exec.Command("buf", "lint", "--path", p.config.ProtoRoot)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("buf lint failed:\n%s", string(output))
	}

	p.logger.Println("Lint passed")
	return nil
}

// runBreaking executes the breaking check stage.
func (p *Pipeline) runBreaking() error {
	p.logger.Println("Stage 2: Breaking Check")

	if !p.config.Breaking.Enable {
		p.logger.Println("Breaking check disabled, skipping")
		return nil
	}

	// Run buf breaking
	cmd := exec.Command("buf", "breaking", "--against", p.config.Breaking.Target)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("buf breaking failed:\n%s", string(output))
	}

	p.logger.Println("Breaking check passed")
	return nil
}

// runDescriptorBuild executes the descriptor build stage.
func (p *Pipeline) runDescriptorBuild() (string, error) {
	p.logger.Println("Stage 3: Descriptor Build")

	descPath := p.config.Descriptors.OutputPath

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(descPath), 0755); err != nil {
		return "", fmt.Errorf("create output dir: %w", err)
	}

	// Run buf build
	cmd := exec.Command("buf", "build", "-o", descPath)
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("buf build failed:\n%s", string(output))
	}

	// Check that descriptor file was created
	if _, err := os.Stat(descPath); err != nil {
		return "", fmt.Errorf("descriptor file not created: %w", err)
	}

	p.logger.Printf("Descriptor built: %s\n", descPath)
	return descPath, nil
}

// runDocModelBuild executes the doc model build stage.
func (p *Pipeline) runDocModelBuild(descPath string) (*ApiDocModel, error) {
	p.logger.Println("Stage 4: Doc Model Build")

	// Load context from descriptor set
	ctx, err := protoctx.LoadFromDescriptorSetFile(descPath)
	if err != nil {
		return nil, fmt.Errorf("load context: %w", err)
	}

	// Get source commit
	sourceCommit := getSourceCommit()

	// Build model
	model, err := BuildApiDocModel(ctx, sourceCommit)
	if err != nil {
		return nil, fmt.Errorf("build model: %w", err)
	}

	// Save model to file
	modelPath := "api-docs/model/api-doc-model.json"
	if err := os.MkdirAll(filepath.Dir(modelPath), 0755); err != nil {
		return nil, fmt.Errorf("create model dir: %w", err)
	}

	if err := model.SaveToFile(modelPath); err != nil {
		return nil, fmt.Errorf("save model: %w", err)
	}

	p.logger.Printf("Doc model built: %d modules, %d services, %d messages\n",
		len(model.Modules),
		model.Statistics["total_services"],
		model.Statistics["total_messages"])

	return model, nil
}

// runEnrichment executes the enrichment stage using protodocs-enricher.
func (p *Pipeline) runEnrichment(model *ApiDocModel) (*ApiDocModel, error) {
	p.logger.Println("Stage 4.5: Enrichment")

	// Build command to run protodocs-enricher
	modelPath := "api-docs/model/api-doc-model.json"

	cmd := exec.Command(
		"go", "run", "./cmd/protodocs-enricher",
		"--config", p.config.Enrichment.ConfigPath,
		"--input", modelPath,
		"--output", p.config.Enrichment.OutputModelPath,
		"--manifest", p.config.Enrichment.ManifestPath,
		"--tenant", p.config.Enrichment.Tenant,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		p.logger.Printf("Warning: Enrichment failed: %v\n", err)
		p.logger.Println("Continuing with non-enriched model...")
		return model, nil // Non-fatal: continue with original model
	}

	// Load enriched model
	enrichedModel, err := LoadApiDocModelFromFile(p.config.Enrichment.OutputModelPath)
	if err != nil {
		p.logger.Printf("Warning: Failed to load enriched model: %v\n", err)
		p.logger.Println("Continuing with non-enriched model...")
		return model, nil // Non-fatal
	}

	p.logger.Println("Enrichment completed successfully")
	return enrichedModel, nil
}

// runDiagramGeneration executes the diagram generation stage.
func (p *Pipeline) runDiagramGeneration(model *ApiDocModel) error {
	p.logger.Println("Stage 4.7: Diagram Generation")

	// Convert pipeline DiagramsConfig to diagrams.DiagramConfig
	diagramConfig := &diagrams.DiagramConfig{
		OutputDir:              p.config.Diagrams.OutputDir,
		EnablePipeline:         p.config.Diagrams.EnablePipeline,
		EnableEnricher:         p.config.Diagrams.EnableEnricher,
		EnableComponent:        p.config.Diagrams.EnableComponent,
		EnableTransform:        p.config.Diagrams.EnableTransform,
		EnableDeploy:           p.config.Diagrams.EnableDeploy,
		EnableDataModel:        p.config.Diagrams.EnableDataModel,
		EnableServiceMap:       p.config.Diagrams.EnableServiceMap,
		EnableMessageHierarchy: p.config.Diagrams.EnableMessageHierarchy,
		GenerateIndex:          p.config.Diagrams.GenerateIndex,
		IncludeTimestamp:       true,
		Theme:                  p.config.Diagrams.Theme,
		MaxServicesPerDiagram:  p.config.Diagrams.MaxServicesPerDiagram,
		MaxMessagesPerDiagram:  p.config.Diagrams.MaxMessagesPerDiagram,
		IncludePrivateTypes:    false,
	}

	// Create diagram generator
	generator := diagrams.NewDiagramGenerator(diagramConfig)

	// Convert ApiDocModel to diagrams.ApiDocModel
	diagramModel := convertTodiagramsModel(model)

	// Generate all diagrams
	results, err := generator.GenerateAll(diagramModel)
	if err != nil {
		p.logger.Printf("Warning: Diagram generation encountered errors: %v\n", err)
	}

	// Count successful diagrams
	successCount := 0
	for _, result := range results {
		if result.Error == nil {
			successCount++
		}
	}

	p.logger.Printf("Diagram generation complete: %d/%d diagrams generated successfully\n",
		successCount, len(results))

	return nil
}

// convertTodiagramsModel converts pipeline.ApiDocModel to diagrams.ApiDocModel
func convertTodiagramsModel(model *ApiDocModel) *diagrams.ApiDocModel {
	if model == nil {
		return nil
	}

	dm := &diagrams.ApiDocModel{
		GeneratedAt:  model.GeneratedAt,
		SourceCommit: model.SourceCommit,
		Statistics:   model.Statistics,
		Tools:        model.Tools,
		Modules:      make([]diagrams.DocModule, len(model.Modules)),
	}

	for i, module := range model.Modules {
		dm.Modules[i] = diagrams.DocModule{
			Name:        module.Name,
			Package:     module.Package,
			Description: module.Description,
			FilePath:    module.FilePath,
			Services:    convertServices(module.Services),
			Messages:    convertMessages(module.Messages),
			Enums:       convertEnums(module.Enums),
		}
	}

	return dm
}

func convertServices(services []DocService) []diagrams.DocService {
	result := make([]diagrams.DocService, len(services))
	for i, svc := range services {
		result[i] = diagrams.DocService{
			Name:        svc.Name,
			FullName:    svc.FullName(),
			Description: svc.Description,
			Methods:     convertMethods(svc.Methods),
			Visibility:  svc.Visibility,
		}
	}
	return result
}

func convertMethods(methods []DocMethod) []diagrams.DocMethod {
	result := make([]diagrams.DocMethod, len(methods))
	for i, method := range methods {
		result[i] = diagrams.DocMethod{
			Name:            method.Name,
			FullName:        method.FullName(),
			Description:     method.Description,
			InputType:       method.InputType,
			OutputType:      method.OutputType,
			ClientStreaming: method.ClientStreaming(),
			ServerStreaming: method.ServerStreaming(),
			HTTPMethods:     convertHTTPMethods(method.HTTPMethods, method.HTTPPaths),
			Visibility:      method.Visibility,
		}
	}
	return result
}

func convertHTTPMethods(httpMethods, httpPaths []string) []diagrams.HTTPMethodInfo {
	maxLen := len(httpMethods)
	if len(httpPaths) > maxLen {
		maxLen = len(httpPaths)
	}

	result := make([]diagrams.HTTPMethodInfo, 0, maxLen)
	for i := 0; i < maxLen; i++ {
		method := ""
		path := ""
		if i < len(httpMethods) {
			method = httpMethods[i]
		}
		if i < len(httpPaths) {
			path = httpPaths[i]
		}
		if method != "" || path != "" {
			result = append(result, diagrams.HTTPMethodInfo{
				Method: method,
				Path:   path,
			})
		}
	}
	return result
}

func convertMessages(messages []DocMessage) []diagrams.DocMessage {
	result := make([]diagrams.DocMessage, len(messages))
	for i, msg := range messages {
		result[i] = diagrams.DocMessage{
			Name:        msg.Name,
			FullName:    msg.FullName(),
			Description: msg.Description,
			Fields:      convertFields(msg.Fields),
			Visibility:  "", // DocMessage in pipeline doesn't have Visibility field
		}
	}
	return result
}

func convertFields(fields []DocField) []diagrams.DocField {
	result := make([]diagrams.DocField, len(fields))
	for i, field := range fields {
		result[i] = diagrams.DocField{
			Name:        field.Name,
			Number:      field.Number,
			Type:        field.Type,
			TypeName:    field.TypeName,
			Label:       field.Label,
			Description: field.Description,
			OneofGroup:  field.OneofGroup,
		}
	}
	return result
}

func convertEnums(enums []DocEnum) []diagrams.DocEnum {
	result := make([]diagrams.DocEnum, len(enums))
	for i, enum := range enums {
		result[i] = diagrams.DocEnum{
			Name:        enum.Name,
			FullName:    enum.FullName(),
			Description: enum.Description,
			Values:      convertEnumValues(enum.Values),
			Visibility:  enum.Visibility,
		}
	}
	return result
}

func convertEnumValues(values []DocEnumValue) []diagrams.DocEnumValue {
	result := make([]diagrams.DocEnumValue, len(values))
	for i, val := range values {
		result[i] = diagrams.DocEnumValue{
			Name:        val.Name,
			Number:      val.Number,
			Description: val.Description,
		}
	}
	return result
}

// runHLDGeneration executes the High-Level Design generation stage.
func (p *Pipeline) runHLDGeneration(model *ApiDocModel) error {
	p.logger.Println("Stage 4.8: High-Level Design Generation")

	// Determine input model path
	inputPath := p.config.HLD.InputModelPath
	if inputPath == "" {
		// Use enriched model if enrichment was enabled
		if p.config.Enrichment.Enabled && p.config.Enrichment.OutputModelPath != "" {
			inputPath = p.config.Enrichment.OutputModelPath
		} else {
			return fmt.Errorf("HLD input model path not configured")
		}
	}

	// Determine config path
	configPath := p.config.HLD.ConfigPath
	if configPath == "" {
		configPath = "configs/hld_generator.yaml"
	}

	// Determine output directory
	outputDir := p.config.HLD.OutputDir
	if outputDir == "" {
		outputDir = "api-docs/hld/"
	}

	p.logger.Printf("HLD Generation: input=%s, config=%s, output=%s\n", inputPath, configPath, outputDir)

	// Build command to run protodocs-hld
	args := []string{
		"--config", configPath,
		"--input", inputPath,
		"--output", outputDir,
	}

	// Add module name if specified
	if p.config.HLD.ModuleName != "" {
		args = append(args, "--module", p.config.HLD.ModuleName)
	}

	cmd := exec.Command("./bin/protodocs-hld", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		p.logger.Printf("HLD generation output:\n%s\n", string(output))
		return fmt.Errorf("HLD generation failed: %w", err)
	}

	p.logger.Printf("HLD generation output:\n%s\n", string(output))
	p.logger.Println("HLD generation completed successfully")

	return nil
}

// runDocsGeneration executes the docs generation stage.
func (p *Pipeline) runDocsGeneration() error {
	p.logger.Println("Stage 5: Docs Generation")

	// For now, skip docs generation as it requires buf plugins
	// TODO: Implement docs generation with buf generate
	p.logger.Println("Docs generation not yet implemented, skipping")

	return nil
}

// runOpenAPIGeneration executes the OpenAPI generation stage.
func (p *Pipeline) runOpenAPIGeneration() error {
	p.logger.Println("Stage 6: OpenAPI Generation")

	// For now, skip OpenAPI generation as it requires buf plugins
	// TODO: Implement OpenAPI generation with buf generate
	p.logger.Println("OpenAPI generation not yet implemented, skipping")

	return nil
}

// runSiteAssembly executes the site assembly stage.
func (p *Pipeline) runSiteAssembly() error {
	p.logger.Println("Stage 7: Site Assembly")

	// For now, skip site assembly as it requires mkdocs/docusaurus
	// TODO: Implement site assembly with mkdocs build
	p.logger.Println("Site assembly not yet implemented, skipping")

	return nil
}

// getSourceCommit gets the current git commit SHA.
func getSourceCommit() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	return string(output[:7]) // Short SHA
}

// getCurrentBranch gets the current git branch.
func getCurrentBranch() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	return string(output[:len(output)-1]) // Remove trailing newline
}
