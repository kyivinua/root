package pipeline

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kyivinua/docgen-tool/tools/protoctx"
	"github.com/kyivinua/docgen-tool/tools/protodocs/diagrams"
	"github.com/kyivinua/docgen-tool/tools/protodocs/publishers/confluence"
)

// Pipeline orchestrates the documentation generation pipeline.
type Pipeline struct {
	config              *PipelineConfig
	logger              *log.Logger
	notificationManager *NotificationManager
	diagramManifest     *diagrams.DiagramManifest

	// Monorepo-specific fields
	serviceGroups       map[string]*ServiceGroup
	consolidationResults map[string]*ConsolidationResult
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

	// Stage 8: Publishers (Confluence, etc.)
	if p.config.Publishers.Confluence.Enabled {
		if err := p.runConfluencePublishing(); err != nil {
			p.logger.Printf("Warning: Confluence publishing failed: %v", err)
			// Don't fail the pipeline for publishing errors
		}
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

	var scope *Scope
	var err error

	// Check if monorepo mode is enabled
	if p.config.Discovery.MonorepoMode {
		return p.runMonorepoDiscovery()
	}

	// Use incremental discovery if enabled
	if p.config.Discovery.Incremental {
		baseRef := p.config.Discovery.BaseRef
		if baseRef == "" {
			baseRef = "main"
		}

		headRef := p.config.Discovery.HeadRef
		if headRef == "" {
			headRef = "HEAD"
		}

		p.logger.Printf("Using incremental discovery: %s..%s", baseRef, headRef)
		scope, err = DiscoverChangedProto(p.config.ProtoRoot, baseRef, headRef)
		if err != nil {
			p.logger.Printf("Warning: Incremental discovery failed, falling back to full discovery: %v", err)
			scope, err = DiscoverAllProto(p.config.ProtoRoot)
		}
	} else {
		// Full discovery
		p.logger.Println("Using full discovery")
		scope, err = DiscoverAllProto(p.config.ProtoRoot)
	}

	if err != nil {
		return nil, err
	}

	p.logger.Printf("Discovered %d proto files in %d packages", len(scope.ProtoFiles), len(scope.ProtoPackages))

	return scope, nil
}

// runMonorepoDiscovery executes monorepo-specific discovery with service grouping and optional consolidation.
func (p *Pipeline) runMonorepoDiscovery() (*Scope, error) {
	p.logger.Println("Using monorepo discovery mode")

	// Create monorepo discovery config
	discoveryConfig := &MonorepoDiscoveryConfig{
		RootDir:         p.config.ProtoRoot,
		ProtoPatterns:   p.config.Discovery.Patterns,
		ExcludePatterns: p.config.Discovery.ExcludePatterns,
		MaxConcurrency:  p.config.Discovery.MaxConcurrency,
	}

	// Set detection strategy
	switch p.config.Discovery.ServiceDetectionStrategy {
	case "directory":
		discoveryConfig.ServiceDetection = DetectByDirectory
	case "package":
		discoveryConfig.ServiceDetection = DetectByPackage
	case "service_definition":
		discoveryConfig.ServiceDetection = DetectByServiceDefinition
	case "hybrid", "":
		discoveryConfig.ServiceDetection = DetectByHybrid
	default:
		p.logger.Printf("Warning: Unknown detection strategy %s, using hybrid", p.config.Discovery.ServiceDetectionStrategy)
		discoveryConfig.ServiceDetection = DetectByHybrid
	}

	// Apply defaults
	if len(discoveryConfig.ProtoPatterns) == 0 {
		discoveryConfig.ProtoPatterns = []string{"**/*.proto"}
	}
	if discoveryConfig.MaxConcurrency == 0 {
		discoveryConfig.MaxConcurrency = 10
	}

	// Create discovery instance
	discovery := NewMonorepoDiscovery(discoveryConfig)

	// Discover all services
	ctx := p.getContext()
	serviceGroups, err := discovery.DiscoverAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("monorepo discovery failed: %w", err)
	}

	p.serviceGroups = serviceGroups
	p.logger.Printf("Discovered %d services in monorepo", len(serviceGroups))

	// Log service details
	for serviceName, group := range serviceGroups {
		p.logger.Printf("  - %s: %d proto files, %d service definitions", serviceName, group.TotalFiles, group.TotalServices)
	}

	// Run consolidation if enabled
	if p.config.Discovery.EnableConsolidation {
		if err := p.runConsolidation(); err != nil {
			return nil, fmt.Errorf("consolidation failed: %w", err)
		}
	}

	// Convert service groups to scope for backward compatibility
	scope := p.convertServiceGroupsToScope()

	p.logger.Printf("Discovered %d proto files in %d packages across %d services",
		len(scope.ProtoFiles), len(scope.ProtoPackages), len(serviceGroups))

	return scope, nil
}

// runConsolidation consolidates proto files by service.
func (p *Pipeline) runConsolidation() error {
	p.logger.Println("Stage 0.5: Proto Consolidation")

	config := &ConsolidationConfig{
		OutputRoot:           p.config.Discovery.ConsolidatedOutputDir,
		CreateBufConfig:      p.config.Discovery.CreateBufConfig,
		PreserveDirStructure: p.config.Discovery.PreserveDirStructure,
	}

	consolidator := NewProtoConsolidator(config)

	ctx := p.getContext()
	results, err := consolidator.ConsolidateAll(ctx, p.serviceGroups)
	if err != nil {
		return err
	}

	p.consolidationResults = results

	// Log consolidation results
	totalFiles := 0
	totalErrors := 0
	for serviceName, result := range results {
		totalFiles += result.FilesCopied
		totalErrors += len(result.Errors)

		if len(result.Errors) > 0 {
			p.logger.Printf("  - %s: %d files copied, %d errors", serviceName, result.FilesCopied, len(result.Errors))
			for _, err := range result.Errors {
				p.logger.Printf("    Error: %v", err)
			}
		} else {
			p.logger.Printf("  - %s: %d files copied to %s", serviceName, result.FilesCopied, result.OutputPath)
		}
	}

	p.logger.Printf("Consolidated %d proto files across %d services (%d errors)",
		totalFiles, len(results), totalErrors)

	return nil
}

// convertServiceGroupsToScope converts service groups to a Scope for backward compatibility.
func (p *Pipeline) convertServiceGroupsToScope() *Scope {
	scope := &Scope{
		ProtoFiles:    []string{},
		ProtoPackages: []string{},
	}

	// Track unique packages
	packageSet := make(map[string]bool)

	for _, group := range p.serviceGroups {
		for _, protoFile := range group.ProtoFiles {
			scope.ProtoFiles = append(scope.ProtoFiles, protoFile.FilePath)

			if protoFile.PackageName != "" && !packageSet[protoFile.PackageName] {
				scope.ProtoPackages = append(scope.ProtoPackages, protoFile.PackageName)
				packageSet[protoFile.PackageName] = true
			}
		}
	}

	return scope
}

// getContext returns a context for pipeline operations.
func (p *Pipeline) getContext() context.Context {
	// For now, return background context
	// In the future, this could be enhanced with timeout/cancellation
	return context.Background()
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

	// Generate static architecture diagrams (infrastructure)
	generator := diagrams.NewDiagramGenerator(diagramConfig)
	diagramModel := convertTodiagramsModel(model)

	results, err := generator.GenerateAll(diagramModel)
	if err != nil {
		p.logger.Printf("Warning: Static diagram generation encountered errors: %v\n", err)
	}

	// Count successful static diagrams
	successCount := 0
	for _, result := range results {
		if result.Error == nil {
			successCount++
		}
	}

	p.logger.Printf("Static diagram generation complete: %d/%d diagrams generated successfully\n",
		successCount, len(results))

	// Generate per-service diagrams (GraphML + Enhanced Mermaid)
	exportConfig := &diagrams.ExportConfig{
		OutputDir:         filepath.Join(p.config.Diagrams.OutputDir, "services"),
		Formats:           []diagrams.DiagramExportFormat{diagrams.FormatGraphML, diagrams.FormatMermaid},
		CreateIndex:       true,
		IncludeTimestamp:  p.config.Diagrams.IncludeTimestamp,
		ServiceSubfolders: true,
	}

	exporter := diagrams.NewDiagramExporter(exportConfig)
	manifest, err := exporter.ExportAllDiagrams(diagramModel)
	if err != nil {
		p.logger.Printf("Warning: Service diagram export encountered errors: %v\n", err)
	} else {
		p.logger.Printf("Service diagram export complete: %d diagrams exported (%.2f KB total)\n",
			manifest.Statistics["total_diagrams"],
			float64(manifest.Statistics["total_size_bytes"])/1024)

		// Store manifest path for later use in Confluence publishing
		p.diagramManifest = manifest
	}

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

	// Use protoc-gen-doc or consolidated generator
	return p.runConsolidatedDocsGeneration()
}

// runConsolidatedDocsGeneration generates consolidated documentation using our own generator.
func (p *Pipeline) runConsolidatedDocsGeneration() error {
	p.logger.Println("Generating consolidated documentation")

	// Check if consolidated-docgen exists
	consolidatedBin := filepath.Join("tools", "protodocs", "cmd", "consolidated-docgen", "consolidated-docgen")
	if _, err := os.Stat(consolidatedBin); os.IsNotExist(err) {
		// Try to build it
		p.logger.Println("Building consolidated-docgen...")
		buildCmd := exec.Command("go", "build", "-o", consolidatedBin,
			"./tools/protodocs/cmd/consolidated-docgen")
		if output, err := buildCmd.CombinedOutput(); err != nil {
			p.logger.Printf("Warning: Could not build consolidated-docgen: %v\n%s", err, string(output))
			p.logger.Println("Skipping docs generation")
			return nil
		}
	}

	// Ensure output directory exists
	if err := os.MkdirAll(p.config.Docs.OutputDir, 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	// Run consolidated-docgen
	args := []string{
		"-proto-dir=" + p.config.ProtoRoot,
		"-output-dir=" + p.config.Docs.OutputDir,
	}

	// Add theme if configured
	if p.config.Diagrams.Theme != "" {
		args = append(args, "-theme="+p.config.Diagrams.Theme)
	}

	cmd := exec.Command(consolidatedBin, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		p.logger.Printf("Docs generation output:\n%s", string(output))
		return fmt.Errorf("docs generation failed: %w", err)
	}

	p.logger.Printf("Documentation generated successfully to %s", p.config.Docs.OutputDir)
	p.logger.Printf("Output:\n%s", string(output))

	return nil
}

// runOpenAPIGeneration executes the OpenAPI generation stage.
func (p *Pipeline) runOpenAPIGeneration() error {
	p.logger.Println("Stage 6: OpenAPI Generation")

	// Check if buf or protoc-gen-openapi is available
	if p.config.UseBuf {
		return p.runBufOpenAPIGeneration()
	}

	return p.runProtocOpenAPIGeneration()
}

// runBufOpenAPIGeneration generates OpenAPI using buf.
func (p *Pipeline) runBufOpenAPIGeneration() error {
	p.logger.Println("Generating OpenAPI with buf")

	// Check if buf is installed
	if _, err := exec.LookPath("buf"); err != nil {
		p.logger.Println("buf not found, skipping OpenAPI generation")
		return nil
	}

	// Ensure output directory exists
	if err := os.MkdirAll(p.config.OpenAPI.OutputDir, 0755); err != nil {
		return fmt.Errorf("create openapi output directory: %w", err)
	}

	// Run buf generate for OpenAPI
	// Note: Requires buf.gen.yaml with protoc-gen-openapiv3 plugin configured
	args := []string{"generate", "--path", p.config.ProtoRoot}

	// Look for buf.gen.yaml
	bufGenFile := "buf.gen.yaml"
	if _, err := os.Stat(bufGenFile); os.IsNotExist(err) {
		p.logger.Printf("Warning: buf.gen.yaml not found, skipping OpenAPI generation")
		p.logger.Println("Create buf.gen.yaml with protoc-gen-openapiv3 plugin configuration")
		return nil
	}

	cmd := exec.Command("buf", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		p.logger.Printf("buf generate output:\n%s", string(output))
		p.logger.Println("Skipping OpenAPI generation (plugin may not be installed)")
		return nil
	}

	p.logger.Printf("OpenAPI generated successfully to %s", p.config.OpenAPI.OutputDir)
	return nil
}

// runProtocOpenAPIGeneration generates OpenAPI using protoc.
func (p *Pipeline) runProtocOpenAPIGeneration() error {
	p.logger.Println("Generating OpenAPI with protoc")

	// Check if protoc is installed
	if _, err := exec.LookPath("protoc"); err != nil {
		p.logger.Println("protoc not found, skipping OpenAPI generation")
		return nil
	}

	// Check if protoc-gen-openapiv3 is installed
	if _, err := exec.LookPath("protoc-gen-openapiv3"); err != nil {
		p.logger.Println("protoc-gen-openapiv3 not found, skipping OpenAPI generation")
		p.logger.Println("Install with: go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest")
		return nil
	}

	// Ensure output directory exists
	if err := os.MkdirAll(p.config.OpenAPI.OutputDir, 0755); err != nil {
		return fmt.Errorf("create openapi output directory: %w", err)
	}

	// Find all proto files
	var protoFiles []string
	err := filepath.Walk(p.config.ProtoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".proto" {
			protoFiles = append(protoFiles, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("find proto files: %w", err)
	}

	if len(protoFiles) == 0 {
		p.logger.Println("No proto files found")
		return nil
	}

	// Run protoc for each file
	for _, protoFile := range protoFiles {
		args := []string{
			"--openapiv3_out=" + p.config.OpenAPI.OutputDir,
			"--proto_path=" + p.config.ProtoRoot,
			"--proto_path=/usr/include", // For well-known types
			protoFile,
		}

		cmd := exec.Command("protoc", args...)
		if output, err := cmd.CombinedOutput(); err != nil {
			p.logger.Printf("Warning: Failed to generate OpenAPI for %s: %v\n%s",
				protoFile, err, string(output))
			continue
		}
	}

	p.logger.Printf("OpenAPI generated successfully to %s", p.config.OpenAPI.OutputDir)
	return nil
}

// runSiteAssembly executes the site assembly stage.
func (p *Pipeline) runSiteAssembly() error {
	p.logger.Println("Stage 7: Site Assembly")

	switch p.config.Site.Generator {
	case "mkdocs":
		return p.runMkDocsBuild()
	case "docusaurus":
		return p.runDocusaurusBuild()
	case "none", "":
		p.logger.Println("Site assembly disabled, skipping")
		return nil
	default:
		return fmt.Errorf("unknown site generator: %s", p.config.Site.Generator)
	}
}

// runMkDocsBuild builds the site using MkDocs.
func (p *Pipeline) runMkDocsBuild() error {
	p.logger.Println("Building site with MkDocs")

	// Check if mkdocs is installed
	if _, err := exec.LookPath("mkdocs"); err != nil {
		p.logger.Println("mkdocs not found, skipping site assembly")
		p.logger.Println("Install with: pip install mkdocs mkdocs-material")
		return nil
	}

	// Check if config file exists
	if p.config.Site.ConfigPath != "" {
		if _, err := os.Stat(p.config.Site.ConfigPath); os.IsNotExist(err) {
			p.logger.Printf("mkdocs config not found at %s, skipping", p.config.Site.ConfigPath)
			return nil
		}
	}

	// Build the site
	args := []string{"build"}

	if p.config.Site.ConfigPath != "" {
		args = append(args, "-f", p.config.Site.ConfigPath)
	}

	if p.config.Site.OutputDir != "" {
		args = append(args, "-d", p.config.Site.OutputDir)
	}

	cmd := exec.Command("mkdocs", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		p.logger.Printf("mkdocs build output:\n%s", string(output))
		return fmt.Errorf("mkdocs build failed: %w", err)
	}

	p.logger.Printf("Site built successfully to %s", p.config.Site.OutputDir)
	return nil
}

// runDocusaurusBuild builds the site using Docusaurus.
func (p *Pipeline) runDocusaurusBuild() error {
	p.logger.Println("Building site with Docusaurus")

	// Check if npm is installed
	if _, err := exec.LookPath("npm"); err != nil {
		p.logger.Println("npm not found, skipping site assembly")
		p.logger.Println("Install Node.js and npm first")
		return nil
	}

	// Check if docusaurus directory exists
	docusaurusDir := filepath.Dir(p.config.Site.ConfigPath)
	if _, err := os.Stat(docusaurusDir); os.IsNotExist(err) {
		p.logger.Printf("Docusaurus directory not found at %s, skipping", docusaurusDir)
		return nil
	}

	// Run npm build
	cmd := exec.Command("npm", "run", "build")
	cmd.Dir = docusaurusDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		p.logger.Printf("Docusaurus build output:\n%s", string(output))
		return fmt.Errorf("docusaurus build failed: %w", err)
	}

	p.logger.Printf("Site built successfully")
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

// runConfluencePublishing executes the Confluence publishing stage.
func (p *Pipeline) runConfluencePublishing() error {
	p.logger.Println("Stage 8: Confluence Publishing")

	cfg := &p.config.Publishers.Confluence

	// Validate configuration
	if cfg.BaseURL == "" {
		return fmt.Errorf("confluence base_url is required")
	}
	if cfg.SpaceKey == "" {
		return fmt.Errorf("confluence space_key is required")
	}

	// Create publisher
	publisherCfg := &confluence.PublisherConfig{
		BaseURL:              cfg.BaseURL,
		Username:             cfg.Username,
		APIToken:             cfg.APIToken,
		SpaceKey:             cfg.SpaceKey,
		ParentPageID:         cfg.ParentPageID,
		CreatePagePerService: cfg.CreatePagePerService,
		PageTitlePrefix:      cfg.PageTitlePrefix,
		IncludeTOC:           cfg.IncludeTOC,
		IncludeDiagrams:      cfg.IncludeDiagrams,
		IncludeCodeExamples:  cfg.IncludeCodeExamples,
		UpdateExisting:       cfg.UpdateExisting,
		VersionLabel:         cfg.VersionLabel,
		VisibilityFilter:     cfg.VisibilityFilter,
	}

	publisher, err := confluence.NewPublisher(publisherCfg)
	if err != nil {
		return fmt.Errorf("create confluence publisher: %w", err)
	}

	// Publish documentation
	var result *confluence.PublishResult

	if cfg.CreatePagePerService {
		// Publish separate pages for each service
		result, err = publisher.PublishFromMarkdownFiles(p.config.Docs.OutputDir)
	} else {
		// Publish as a single consolidated page
		title := "API Documentation"
		if cfg.PageTitlePrefix != "" {
			title = cfg.PageTitlePrefix + " " + title
		}
		result, err = publisher.PublishConsolidatedPage(p.config.Docs.OutputDir, title)
	}

	if err != nil {
		return fmt.Errorf("publish to confluence: %w", err)
	}

	// Log results
	p.logger.Printf("Confluence publishing complete:")
	p.logger.Printf("  Pages created: %d", result.PagesCreated)
	p.logger.Printf("  Pages updated: %d", result.PagesUpdated)
	p.logger.Printf("  Errors: %d", len(result.Errors))

	for _, url := range result.PageURLs {
		p.logger.Printf("  📄 %s", url)
	}

	if len(result.Errors) > 0 {
		p.logger.Println("Errors encountered:")
		for _, err := range result.Errors {
			p.logger.Printf("  ❌ %v", err)
		}
	}

	// Upload diagrams as attachments if available and enabled
	if cfg.IncludeDiagrams && p.diagramManifest != nil && len(result.PageIDs) > 0 {
		if err := p.uploadDiagramAttachments(result, publisher); err != nil {
			p.logger.Printf("Warning: Failed to upload diagram attachments: %v", err)
		}
	}

	return nil
}

// uploadDiagramAttachments uploads diagrams as attachments to Confluence pages
func (p *Pipeline) uploadDiagramAttachments(result *confluence.PublishResult, publisher *confluence.Publisher) error {
	if p.diagramManifest == nil || len(p.diagramManifest.Diagrams) == 0 {
		return nil
	}

	p.logger.Printf("Uploading %d diagram files as attachments...", len(p.diagramManifest.Diagrams))

	uploadCount := 0
	errorCount := 0

	// Group diagrams by service name
	diagramsByService := make(map[string][]diagrams.DiagramExport)
	for _, diagram := range p.diagramManifest.Diagrams {
		serviceName := diagram.ServiceName
		if serviceName == "" {
			serviceName = "consolidated"
		}
		diagramsByService[serviceName] = append(diagramsByService[serviceName], diagram)
	}

	// Upload diagrams to their corresponding pages
	for serviceName, serviceDiagrams := range diagramsByService {
		pageID, ok := result.PageMap[serviceName]
		if !ok {
			// If no exact match, try to upload to the first page (consolidated)
			if len(result.PageIDs) > 0 {
				pageID = result.PageIDs[0]
			} else {
				p.logger.Printf("Warning: No page found for service '%s', skipping %d diagrams", serviceName, len(serviceDiagrams))
				continue
			}
		}

		// Upload each diagram for this service
		for _, diagram := range serviceDiagrams {
			// Read diagram file
			content, err := os.ReadFile(diagram.FilePath)
			if err != nil {
				p.logger.Printf("Warning: Failed to read diagram file %s: %v", diagram.FilePath, err)
				errorCount++
				continue
			}

			// Upload as attachment
			comment := fmt.Sprintf("%s - Generated: %s", diagram.Description, diagram.GeneratedAt.Format("2006-01-02 15:04:05"))
			if err := publisher.UploadAttachment(pageID, diagram.Filename, content, comment); err != nil {
				p.logger.Printf("Warning: Failed to upload %s to page %s: %v", diagram.Filename, pageID, err)
				errorCount++
				continue
			}

			uploadCount++
			p.logger.Printf("  ✓ Uploaded %s (%s) to page %s", diagram.Filename, diagram.Format, pageID)
		}
	}

	p.logger.Printf("Diagram upload complete: %d uploaded, %d errors", uploadCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("%d diagram uploads failed", errorCount)
	}

	return nil
}

// RunLint executes only the linting stage
func (p *Pipeline) RunLint() error {
	p.logger.Println("Running lint stage only")
	return p.runLint()
}

// RunBreaking executes only the breaking change detection stage
func (p *Pipeline) RunBreaking() error {
	p.logger.Println("Running breaking check stage only")
	return p.runBreaking()
}

// RunDescriptorBuild executes only the descriptor build stage
func (p *Pipeline) RunDescriptorBuild() (string, error) {
	p.logger.Println("Running descriptor build stage only")
	return p.runDescriptorBuild()
}

// RunDocModelBuild executes descriptor build and doc model build stages
func (p *Pipeline) RunDocModelBuild() (*ApiDocModel, error) {
	p.logger.Println("Running doc model build stages")

	// Build descriptor first
	descPath, err := p.runDescriptorBuild()
	if err != nil {
		return nil, fmt.Errorf("descriptor build: %w", err)
	}

	// Then build doc model
	model, err := p.runDocModelBuild(descPath)
	if err != nil {
		return nil, fmt.Errorf("doc model build: %w", err)
	}

	return model, nil
}
