package pipeline

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/kyivinua/docgen-tool/tools/protodocs/pkg/errors"
)

// ServiceProtoFile represents a proto file with its service information
type ServiceProtoFile struct {
	FilePath     string   // Absolute path to proto file
	RelativePath string   // Path relative to monorepo root
	PackageName  string   // Proto package name (from "package" declaration)
	Services     []string // Service names defined in this file
	ServiceOwner string   // Primary service this file belongs to
	Dependencies []string // Imported proto files
}

// ServiceGroup represents a group of proto files belonging to a service
type ServiceGroup struct {
	ServiceName  string              // Name of the service
	ProtoFiles   []*ServiceProtoFile // All proto files for this service
	PackageName  string              // Primary package name
	RootPath     string              // Root directory for this service's protos
	TotalFiles   int                 // Total number of proto files
	TotalServices int                // Total number of service definitions
}

// MonorepoDiscoveryConfig configures monorepo proto discovery
type MonorepoDiscoveryConfig struct {
	RootDir            string   // Monorepo root directory
	ProtoPatterns      []string // Glob patterns for proto files (e.g., "services/*/api/**/*.proto")
	ExcludePatterns    []string // Patterns to exclude (e.g., "vendor/**", "third_party/**")
	ServiceDetection   ServiceDetectionStrategy
	MaxConcurrency     int  // Maximum concurrent file parsing
	ParseDependencies  bool // Whether to parse import statements
	GroupByDirectory   bool // Group by directory structure in addition to service definitions
}

// ServiceDetectionStrategy defines how to detect service ownership
type ServiceDetectionStrategy string

const (
	// DetectByServiceDefinition uses "service Foo" declarations in proto files
	DetectByServiceDefinition ServiceDetectionStrategy = "service_definition"

	// DetectByDirectory uses directory structure (e.g., services/user-service/api/)
	DetectByDirectory ServiceDetectionStrategy = "directory"

	// DetectByPackage uses package declaration (e.g., package user.v1)
	DetectByPackage ServiceDetectionStrategy = "package"

	// DetectByHybrid combines all strategies
	DetectByHybrid ServiceDetectionStrategy = "hybrid"
)

// DefaultMonorepoDiscoveryConfig returns sensible defaults
func DefaultMonorepoDiscoveryConfig(rootDir string) *MonorepoDiscoveryConfig {
	return &MonorepoDiscoveryConfig{
		RootDir: rootDir,
		ProtoPatterns: []string{
			"**/api/**/*.proto",
			"**/proto/**/*.proto",
			"**/protobuf/**/*.proto",
			"services/**/api/**/*.proto",
			"pkg/**/api/**/*.proto",
		},
		ExcludePatterns: []string{
			"vendor/**",
			"third_party/**",
			"node_modules/**",
			".git/**",
			"**/*_test.proto",
		},
		ServiceDetection:  DetectByHybrid,
		MaxConcurrency:    10,
		ParseDependencies: true,
		GroupByDirectory:  true,
	}
}

// MonorepoDiscovery handles discovery and grouping of proto files in monorepo
type MonorepoDiscovery struct {
	config *MonorepoDiscoveryConfig
	mu     sync.RWMutex
	cache  map[string]*ServiceProtoFile // Cache parsed files
}

// NewMonorepoDiscovery creates a new monorepo discovery instance
func NewMonorepoDiscovery(config *MonorepoDiscoveryConfig) *MonorepoDiscovery {
	if config == nil {
		config = DefaultMonorepoDiscoveryConfig(".")
	}

	if config.MaxConcurrency <= 0 {
		config.MaxConcurrency = 10
	}

	return &MonorepoDiscovery{
		config: config,
		cache:  make(map[string]*ServiceProtoFile),
	}
}

// DiscoverAll finds and groups all proto files in the monorepo
func (md *MonorepoDiscovery) DiscoverAll(ctx context.Context) (map[string]*ServiceGroup, error) {
	// Find all proto files matching patterns
	protoFiles, err := md.findProtoFiles()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to find proto files")
	}

	if len(protoFiles) == 0 {
		return make(map[string]*ServiceGroup), nil
	}

	// Parse files concurrently
	parsedFiles, err := md.parseProtoFilesConcurrent(ctx, protoFiles)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to parse proto files")
	}

	// Group files by service
	serviceGroups := md.groupByService(parsedFiles)

	return serviceGroups, nil
}

// findProtoFiles finds all proto files matching configured patterns
func (md *MonorepoDiscovery) findProtoFiles() ([]string, error) {
	var allFiles []string
	fileSet := make(map[string]bool)

	// Search for each pattern
	for _, pattern := range md.config.ProtoPatterns {
		files, err := md.globProtoFiles(pattern)
		if err != nil {
			continue // Skip patterns that fail
		}

		for _, file := range files {
			// Check if file should be excluded
			if md.shouldExclude(file) {
				continue
			}

			// Deduplicate
			if !fileSet[file] {
				allFiles = append(allFiles, file)
				fileSet[file] = true
			}
		}
	}

	return allFiles, nil
}

// globProtoFiles uses filepath.Glob to match pattern
func (md *MonorepoDiscovery) globProtoFiles(pattern string) ([]string, error) {
	// Convert pattern to absolute if it's relative
	if !filepath.IsAbs(pattern) {
		pattern = filepath.Join(md.config.RootDir, pattern)
	}

	// Use filepath.Walk for ** patterns
	if strings.Contains(pattern, "**") {
		return md.walkPattern(pattern)
	}

	// Use standard glob for simple patterns
	return filepath.Glob(pattern)
}

// walkPattern implements ** glob pattern matching
func (md *MonorepoDiscovery) walkPattern(pattern string) ([]string, error) {
	var matches []string

	// Split pattern at ** to get base and suffix
	parts := strings.Split(pattern, "**")
	if len(parts) == 0 {
		return matches, nil
	}

	baseDir := parts[0]
	if baseDir == "" {
		baseDir = md.config.RootDir
	}

	suffix := ""
	if len(parts) > 1 {
		suffix = parts[1]
	}

	// Walk directory tree
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() {
			return nil
		}

		// Check if path matches suffix
		if suffix != "" {
			matched, _ := filepath.Match(suffix, filepath.Base(path))
			if !matched && !strings.HasSuffix(path, suffix) {
				return nil
			}
		}

		// Must be .proto file
		if filepath.Ext(path) == ".proto" {
			matches = append(matches, path)
		}

		return nil
	})

	return matches, err
}

// shouldExclude checks if file should be excluded based on patterns
func (md *MonorepoDiscovery) shouldExclude(filePath string) bool {
	// Convert to relative path for pattern matching
	relativePath := strings.TrimPrefix(filePath, md.config.RootDir+"/")
	relativePath = strings.TrimPrefix(relativePath, md.config.RootDir)
	relativePath = strings.TrimPrefix(relativePath, "/")

	for _, pattern := range md.config.ExcludePatterns {
		// Convert ** pattern to regex
		regexPattern := strings.ReplaceAll(pattern, "**", ".*")
		regexPattern = strings.ReplaceAll(regexPattern, "*", "[^/]*")
		regexPattern = "^" + regexPattern + "$"

		if matched, _ := regexp.MatchString(regexPattern, relativePath); matched {
			return true
		}
	}
	return false
}

// parseProtoFilesConcurrent parses multiple proto files concurrently
func (md *MonorepoDiscovery) parseProtoFilesConcurrent(ctx context.Context, files []string) ([]*ServiceProtoFile, error) {
	// Create buffered channel for results
	results := make(chan *ServiceProtoFile, len(files))
	errChan := make(chan error, len(files))

	// Create semaphore for concurrency control
	sem := make(chan struct{}, md.config.MaxConcurrency)

	var wg sync.WaitGroup

	// Parse files concurrently
	for _, file := range files {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Check context
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			default:
			}

			// Check cache first
			md.mu.RLock()
			if cached, ok := md.cache[filePath]; ok {
				md.mu.RUnlock()
				results <- cached
				return
			}
			md.mu.RUnlock()

			// Parse file
			parsed, err := md.parseProtoFile(filePath)
			if err != nil {
				errChan <- errors.Wrap(err, errors.ErrorTypeInternal, "failed to parse "+filePath)
				return
			}

			// Cache result
			md.mu.Lock()
			md.cache[filePath] = parsed
			md.mu.Unlock()

			results <- parsed
		}(file)
	}

	// Wait for all goroutines
	go func() {
		wg.Wait()
		close(results)
		close(errChan)
	}()

	// Collect results
	var parsedFiles []*ServiceProtoFile
	var errs []error

	for result := range results {
		parsedFiles = append(parsedFiles, result)
	}

	for err := range errChan {
		errs = append(errs, err)
	}

	// Return first error if any
	if len(errs) > 0 {
		return parsedFiles, errs[0]
	}

	return parsedFiles, nil
}

// parseProtoFile parses a single proto file to extract service information
func (md *MonorepoDiscovery) parseProtoFile(filePath string) (*ServiceProtoFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	spf := &ServiceProtoFile{
		FilePath:     filePath,
		RelativePath: strings.TrimPrefix(filePath, md.config.RootDir+"/"),
		Services:     []string{},
		Dependencies: []string{},
	}

	scanner := bufio.NewScanner(file)

	// Regex patterns
	packageRegex := regexp.MustCompile(`^\s*package\s+([a-zA-Z0-9_.]+)\s*;`)
	serviceRegex := regexp.MustCompile(`^\s*service\s+([a-zA-Z0-9_]+)\s*\{`)
	importRegex := regexp.MustCompile(`^\s*import\s+["']([^"']+)["']\s*;`)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments
		if strings.HasPrefix(strings.TrimSpace(line), "//") {
			continue
		}

		// Extract package name
		if matches := packageRegex.FindStringSubmatch(line); len(matches) > 1 {
			spf.PackageName = matches[1]
		}

		// Extract service names
		if matches := serviceRegex.FindStringSubmatch(line); len(matches) > 1 {
			spf.Services = append(spf.Services, matches[1])
		}

		// Extract dependencies if configured
		if md.config.ParseDependencies {
			if matches := importRegex.FindStringSubmatch(line); len(matches) > 1 {
				spf.Dependencies = append(spf.Dependencies, matches[1])
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Determine service owner based on strategy
	spf.ServiceOwner = md.determineServiceOwner(spf)

	return spf, nil
}

// determineServiceOwner determines which service owns this proto file
func (md *MonorepoDiscovery) determineServiceOwner(spf *ServiceProtoFile) string {
	switch md.config.ServiceDetection {
	case DetectByServiceDefinition:
		return md.detectByServiceDefinition(spf)
	case DetectByDirectory:
		return md.detectByDirectory(spf)
	case DetectByPackage:
		return md.detectByPackage(spf)
	case DetectByHybrid:
		return md.detectByHybrid(spf)
	default:
		return md.detectByHybrid(spf)
	}
}

// detectByServiceDefinition uses the first service definition found
func (md *MonorepoDiscovery) detectByServiceDefinition(spf *ServiceProtoFile) string {
	if len(spf.Services) > 0 {
		return spf.Services[0]
	}
	return "common" // Default for files without services (messages, enums, etc.)
}

// detectByDirectory extracts service name from directory structure
// Examples:
//   - services/user-service/api/v1/user.proto -> user-service
//   - pkg/auth-service/proto/auth.proto -> auth-service
//   - apps/billing/api/billing.proto -> billing
func (md *MonorepoDiscovery) detectByDirectory(spf *ServiceProtoFile) string {
	parts := strings.Split(spf.RelativePath, string(filepath.Separator))

	// Look for common service directory patterns
	for i, part := range parts {
		// Check for "services/service-name" pattern
		if part == "services" && i+1 < len(parts) {
			return parts[i+1]
		}

		// Check for "pkg/service-name" pattern
		if part == "pkg" && i+1 < len(parts) {
			return parts[i+1]
		}

		// Check for "apps/service-name" pattern
		if part == "apps" && i+1 < len(parts) {
			return parts[i+1]
		}

		// Check for directories ending with "-service" or "-api"
		if strings.HasSuffix(part, "-service") || strings.HasSuffix(part, "-api") {
			return part
		}
	}

	// Fallback to first directory component
	if len(parts) > 0 {
		return parts[0]
	}

	return "unknown"
}

// detectByPackage uses the package declaration
// Examples:
//   - package user.v1 -> user
//   - package com.company.billing.v1 -> billing
func (md *MonorepoDiscovery) detectByPackage(spf *ServiceProtoFile) string {
	if spf.PackageName == "" {
		return "unknown"
	}

	// Split package by dots
	parts := strings.Split(spf.PackageName, ".")

	// Skip common prefixes (com, org, io, etc.)
	startIdx := 0
	if len(parts) > 0 {
		first := parts[0]
		if first == "com" || first == "org" || first == "io" || first == "net" {
			startIdx = 1
		}
	}

	// Skip company name after domain prefix (e.g., com.company.billing -> billing)
	if len(parts) > startIdx+1 && startIdx > 0 {
		startIdx++
	}

	// Common suffixes to skip (internal, common, shared, api, v1, v2, etc.)
	commonSuffixes := map[string]bool{
		"internal": true,
		"common":   true,
		"shared":   true,
		"api":      true,
	}

	// Find first non-version, non-suffix part
	for i := startIdx; i < len(parts); i++ {
		part := parts[i]
		// Skip version parts (v1, v2, etc.)
		if matched, _ := regexp.MatchString(`^v\d+$`, part); matched {
			continue
		}
		// Skip common suffixes if not the only part left
		if commonSuffixes[part] && i < len(parts)-1 {
			continue
		}
		// If this is a suffix but it's the only non-version part, use the previous part
		if commonSuffixes[part] && i > startIdx {
			return parts[i-1]
		}
		return part
	}

	// Fallback to first non-version part
	for i := startIdx; i < len(parts); i++ {
		part := parts[i]
		if matched, _ := regexp.MatchString(`^v\d+$`, part); !matched {
			return part
		}
	}

	return parts[0]
}

// detectByHybrid uses multiple strategies and picks the best match
func (md *MonorepoDiscovery) detectByHybrid(spf *ServiceProtoFile) string {
	// Priority:
	// 1. Service definition (most explicit)
	// 2. Directory structure (very reliable)
	// 3. Package name (less reliable but useful)

	// If file defines services, use that
	if len(spf.Services) > 0 {
		return spf.Services[0]
	}

	// Try directory-based detection
	dirService := md.detectByDirectory(spf)
	if dirService != "unknown" && dirService != "" {
		return dirService
	}

	// Fallback to package-based detection
	pkgService := md.detectByPackage(spf)
	if pkgService != "unknown" && pkgService != "" {
		return pkgService
	}

	return "common"
}

// groupByService groups parsed proto files by service
func (md *MonorepoDiscovery) groupByService(files []*ServiceProtoFile) map[string]*ServiceGroup {
	groups := make(map[string]*ServiceGroup)

	for _, file := range files {
		serviceName := file.ServiceOwner
		if serviceName == "" {
			serviceName = "common"
		}

		// Get or create service group
		group, exists := groups[serviceName]
		if !exists {
			group = &ServiceGroup{
				ServiceName: serviceName,
				ProtoFiles:  []*ServiceProtoFile{},
				PackageName: file.PackageName,
				RootPath:    filepath.Dir(file.FilePath),
			}
			groups[serviceName] = group
		}

		// Add file to group
		group.ProtoFiles = append(group.ProtoFiles, file)
		group.TotalFiles++
		group.TotalServices += len(file.Services)

		// Update root path to common ancestor if different
		if group.RootPath != filepath.Dir(file.FilePath) {
			group.RootPath = md.findCommonAncestor(group.RootPath, filepath.Dir(file.FilePath))
		}
	}

	return groups
}

// findCommonAncestor finds the common ancestor directory of two paths
func (md *MonorepoDiscovery) findCommonAncestor(path1, path2 string) string {
	parts1 := strings.Split(filepath.Clean(path1), string(filepath.Separator))
	parts2 := strings.Split(filepath.Clean(path2), string(filepath.Separator))

	commonParts := []string{}
	minLen := len(parts1)
	if len(parts2) < minLen {
		minLen = len(parts2)
	}

	for i := 0; i < minLen; i++ {
		if parts1[i] == parts2[i] {
			commonParts = append(commonParts, parts1[i])
		} else {
			break
		}
	}

	if len(commonParts) == 0 {
		return md.config.RootDir
	}

	return strings.Join(commonParts, string(filepath.Separator))
}

// GetServiceGroup returns a specific service group by name
func (md *MonorepoDiscovery) GetServiceGroup(ctx context.Context, serviceName string) (*ServiceGroup, error) {
	groups, err := md.DiscoverAll(ctx)
	if err != nil {
		return nil, err
	}

	group, exists := groups[serviceName]
	if !exists {
		return nil, errors.New(errors.ErrorTypeNotFound, "service not found: "+serviceName)
	}

	return group, nil
}

// ListServices returns a list of all discovered service names
func (md *MonorepoDiscovery) ListServices(ctx context.Context) ([]string, error) {
	groups, err := md.DiscoverAll(ctx)
	if err != nil {
		return nil, err
	}

	services := make([]string, 0, len(groups))
	for serviceName := range groups {
		services = append(services, serviceName)
	}

	return services, nil
}

// GetServiceProtoFiles returns all proto files for a specific service
func (group *ServiceGroup) GetServiceProtoFiles() []string {
	paths := make([]string, len(group.ProtoFiles))
	for i, file := range group.ProtoFiles {
		paths[i] = file.FilePath
	}
	return paths
}

// GetPackageNames returns all unique package names in this service group
func (group *ServiceGroup) GetPackageNames() []string {
	packageSet := make(map[string]bool)
	for _, file := range group.ProtoFiles {
		if file.PackageName != "" {
			packageSet[file.PackageName] = true
		}
	}

	packages := make([]string, 0, len(packageSet))
	for pkg := range packageSet {
		packages = append(packages, pkg)
	}
	return packages
}

// GetServiceNames returns all unique service names defined in this group
func (group *ServiceGroup) GetServiceNames() []string {
	serviceSet := make(map[string]bool)
	for _, file := range group.ProtoFiles {
		for _, service := range file.Services {
			serviceSet[service] = true
		}
	}

	services := make([]string, 0, len(serviceSet))
	for svc := range serviceSet {
		services = append(services, svc)
	}
	return services
}

// Summary returns a human-readable summary of the service group
func (group *ServiceGroup) Summary() string {
	return fmt.Sprintf("Service: %s | Files: %d | Services: %d | Packages: %v",
		group.ServiceName,
		group.TotalFiles,
		group.TotalServices,
		group.GetPackageNames())
}
