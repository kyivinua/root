package pipeline

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kyivinua/docgen-tool/tools/protodocs/pkg/errors"
)

// ConsolidationConfig holds configuration for proto consolidation
type ConsolidationConfig struct {
	// OutputRoot is the root directory where consolidated services will be placed
	OutputRoot string

	// CreateBufConfig determines if buf.yaml should be generated for each service
	CreateBufConfig bool

	// CopyDependencies determines if imported proto files should be copied
	CopyDependencies bool

	// PreserveDirStructure preserves the original directory structure within service dirs
	PreserveDirStructure bool

	// FlattenStructure puts all proto files in a single directory per service
	FlattenStructure bool

	// Optional callbacks and logging
	Logger           Logger           // Logger for consolidation operations
	ProgressCallback ProgressCallback // Progress callback for UI updates
	EnableMetrics    bool             // Enable detailed metrics collection
}

// ConsolidationResult contains information about consolidated services
type ConsolidationResult struct {
	ServiceName       string
	OutputPath        string
	FilesCopied       int
	DependenciesCopied int
	BufConfigCreated  bool
	Errors            []error
}

// ProtoConsolidator handles consolidation of proto files by service
type ProtoConsolidator struct {
	config  *ConsolidationConfig
	logger  Logger
	metrics *ConsolidationMetrics
}

// NewProtoConsolidator creates a new proto consolidator
func NewProtoConsolidator(config *ConsolidationConfig) *ProtoConsolidator {
	if config == nil {
		config = &ConsolidationConfig{
			OutputRoot:           "./consolidated-protos",
			CreateBufConfig:      true,
			CopyDependencies:     false,
			PreserveDirStructure: true,
			FlattenStructure:     false,
		}
	}

	// Set default logger if not provided
	logger := config.Logger
	if logger == nil {
		logger = &DefaultLogger{}
	}

	// Initialize metrics if enabled
	var metrics *ConsolidationMetrics
	if config.EnableMetrics {
		metrics = &ConsolidationMetrics{
			ServiceMetrics: make(map[string]*ServiceConsolidationMetrics),
		}
	}

	return &ProtoConsolidator{
		config:  config,
		logger:  logger,
		metrics: metrics,
	}
}

// GetMetrics returns consolidation metrics (if enabled)
func (pc *ProtoConsolidator) GetMetrics() *ConsolidationMetrics {
	return pc.metrics
}

// ConsolidateAll consolidates all service groups into separate directories
func (pc *ProtoConsolidator) ConsolidateAll(ctx context.Context, serviceGroups map[string]*ServiceGroup) (map[string]*ConsolidationResult, error) {
	startTime := time.Now()

	if pc.metrics != nil {
		defer func() {
			pc.metrics.Duration = time.Since(startTime)
		}()
	}

	pc.logger.Info("Starting consolidation of %d services to: %s", len(serviceGroups), pc.config.OutputRoot)

	results := make(map[string]*ConsolidationResult)
	current := 0

	for serviceName, group := range serviceGroups {
		current++

		select {
		case <-ctx.Done():
			pc.logger.Warn("Consolidation cancelled by context")
			return results, ctx.Err()
		default:
		}

		// Report progress
		if pc.config.ProgressCallback != nil {
			pc.config.ProgressCallback("consolidation", current, len(serviceGroups),
				fmt.Sprintf("Consolidating %s", serviceName))
		}

		pc.logger.Info("Consolidating service %d/%d: %s", current, len(serviceGroups), serviceName)

		result, err := pc.ConsolidateService(ctx, group)
		if err != nil {
			pc.logger.Error("Failed to consolidate %s: %v", serviceName, err)
			return nil, errors.Wrap(err, errors.ErrorTypeInternal, fmt.Sprintf("failed to consolidate service %s", serviceName))
		}

		results[serviceName] = result

		// Update metrics
		if pc.metrics != nil {
			pc.metrics.ServicesProcessed++
			pc.metrics.TotalFilesCopied += result.FilesCopied
			pc.metrics.TotalErrors += len(result.Errors)
			if result.BufConfigCreated {
				pc.metrics.BufConfigsCreated++
			}
			pc.metrics.ReadmesCreated++ // README is always created
		}

		pc.logger.Info("Consolidated %s: %d files copied to %s",
			serviceName, result.FilesCopied, result.OutputPath)
	}

	// Report completion
	if pc.config.ProgressCallback != nil {
		pc.config.ProgressCallback("consolidation", len(serviceGroups), len(serviceGroups), "Consolidation completed")
	}

	totalFiles := 0
	if pc.metrics != nil {
		totalFiles = pc.metrics.TotalFilesCopied
	}
	pc.logger.Info("Consolidation completed: %d services, %d files, took %v",
		len(results), totalFiles, time.Since(startTime))

	return results, nil
}

// ConsolidateService consolidates proto files for a single service
func (pc *ProtoConsolidator) ConsolidateService(ctx context.Context, group *ServiceGroup) (*ConsolidationResult, error) {
	if group == nil {
		return nil, errors.New(errors.ErrorTypeValidation, "service group cannot be nil")
	}

	result := &ConsolidationResult{
		ServiceName: group.ServiceName,
		OutputPath:  filepath.Join(pc.config.OutputRoot, group.ServiceName),
		Errors:      []error{},
	}

	// Create service output directory
	if err := os.MkdirAll(result.OutputPath, 0755); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to create output directory")
	}

	// Copy proto files
	for _, protoFile := range group.ProtoFiles {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		if err := pc.copyProtoFile(protoFile, group, result); err != nil {
			result.Errors = append(result.Errors, err)
			continue
		}
		result.FilesCopied++
	}

	// Create buf.yaml if requested
	if pc.config.CreateBufConfig {
		if err := pc.createBufConfig(group, result.OutputPath); err != nil {
			result.Errors = append(result.Errors, err)
		} else {
			result.BufConfigCreated = true
		}
	}

	// Create README for the service
	if err := pc.createServiceReadme(group, result.OutputPath); err != nil {
		result.Errors = append(result.Errors, err)
	}

	return result, nil
}

// copyProtoFile copies a proto file to the consolidated location
func (pc *ProtoConsolidator) copyProtoFile(protoFile *ServiceProtoFile, group *ServiceGroup, result *ConsolidationResult) error {
	var destPath string

	if pc.config.FlattenStructure {
		// Flatten: put all files in root of service directory
		destPath = filepath.Join(result.OutputPath, filepath.Base(protoFile.FilePath))
	} else if pc.config.PreserveDirStructure {
		// Preserve: maintain relative path structure
		destPath = filepath.Join(result.OutputPath, protoFile.RelativePath)
	} else {
		// Default: organize by package
		packagePath := strings.ReplaceAll(protoFile.PackageName, ".", "/")
		destPath = filepath.Join(result.OutputPath, packagePath, filepath.Base(protoFile.FilePath))
	}

	// Create destination directory
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to create destination directory")
	}

	// Copy file
	srcFile, err := os.Open(protoFile.FilePath)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to open source file")
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to create destination file")
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to copy file content")
	}

	return nil
}

// createBufConfig creates a buf.yaml configuration for the service
func (pc *ProtoConsolidator) createBufConfig(group *ServiceGroup, outputPath string) error {
	bufConfigPath := filepath.Join(outputPath, "buf.yaml")

	// Collect unique package names
	packages := group.GetPackageNames()

	// Create buf.yaml content
	content := fmt.Sprintf(`version: v1
name: buf.build/%s
deps: []
breaking:
  use:
    - FILE
lint:
  use:
    - DEFAULT
`, group.ServiceName)

	// Add package information as comments
	if len(packages) > 0 {
		content += "\n# Packages in this service:\n"
		for _, pkg := range packages {
			content += fmt.Sprintf("# - %s\n", pkg)
		}
	}

	if err := os.WriteFile(bufConfigPath, []byte(content), 0644); err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to write buf.yaml")
	}

	return nil
}

// createServiceReadme creates a README.md for the consolidated service
func (pc *ProtoConsolidator) createServiceReadme(group *ServiceGroup, outputPath string) error {
	readmePath := filepath.Join(outputPath, "README.md")

	packages := group.GetPackageNames()
	services := group.GetServiceNames()

	content := fmt.Sprintf(`# %s

Auto-generated consolidated proto files for service: **%s**

## Statistics

- **Total Proto Files**: %d
- **Total Services**: %d
- **Total Packages**: %d

## Services

`, group.ServiceName, group.ServiceName, group.TotalFiles, group.TotalServices, len(packages))

	if len(services) > 0 {
		for _, service := range services {
			content += fmt.Sprintf("- `%s`\n", service)
		}
	} else {
		content += "_No service definitions (common/shared proto files)_\n"
	}

	content += "\n## Packages\n\n"
	if len(packages) > 0 {
		for _, pkg := range packages {
			content += fmt.Sprintf("- `%s`\n", pkg)
		}
	}

	content += "\n## Proto Files\n\n"
	for _, protoFile := range group.ProtoFiles {
		content += fmt.Sprintf("- `%s`\n", protoFile.RelativePath)
		if len(protoFile.Services) > 0 {
			content += fmt.Sprintf("  - Services: %s\n", strings.Join(protoFile.Services, ", "))
		}
		if protoFile.PackageName != "" {
			content += fmt.Sprintf("  - Package: `%s`\n", protoFile.PackageName)
		}
	}

	content += fmt.Sprintf("\n---\n_Generated: %s_\n", group.ServiceName)

	if err := os.WriteFile(readmePath, []byte(content), 0644); err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to write README.md")
	}

	return nil
}

// GetConsolidatedPath returns the consolidated path for a service
func (pc *ProtoConsolidator) GetConsolidatedPath(serviceName string) string {
	return filepath.Join(pc.config.OutputRoot, serviceName)
}

// CleanOutputDirectory removes all consolidated files
func (pc *ProtoConsolidator) CleanOutputDirectory() error {
	if err := os.RemoveAll(pc.config.OutputRoot); err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to clean output directory")
	}
	return nil
}

// DefaultConsolidationConfig returns a default configuration
func DefaultConsolidationConfig(outputRoot string) *ConsolidationConfig {
	return &ConsolidationConfig{
		OutputRoot:           outputRoot,
		CreateBufConfig:      true,
		CopyDependencies:     false,
		PreserveDirStructure: true,
		FlattenStructure:     false,
	}
}
