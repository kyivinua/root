// Package resolver provides shared resource resolution for proto files.
package resolver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kyivinua/docgen-tool/internal/docgen"
	"github.com/rs/zerolog"
)

// SharedResourceResolver handles shared proto file resources.
type SharedResourceResolver struct {
	config          Config
	logger          zerolog.Logger
	sharedMessages  map[string]*docgen.Message
	sharedEnums     map[string]*SharedEnum
	protoCache      map[string]*ProtoFile
	importGraph     *ImportGraph
	mu              sync.RWMutex
}

// Config configures the shared resource resolver.
type Config struct {
	// Shared proto paths to scan
	SharedProtoPaths []string

	// Common proto directory patterns
	CommonDirPatterns []string

	// Enable auto-discovery of shared resources
	AutoDiscover bool

	// Import resolution strategy
	ImportStrategy string // local, global, hybrid

	// Generate virtual services for shared protos
	GenerateVirtualServices bool

	// Deduplicate shared messages
	DeduplicateMessages bool
}

// ProtoFile represents a parsed proto file.
type ProtoFile struct {
	Path     string
	Package  string
	Imports  []string
	Messages []docgen.Message
	Services []*ProtoService
	Enums    []*SharedEnum
}

// ProtoService represents a service definition in proto.
type ProtoService struct {
	Name    string
	Methods []docgen.Method
}

// SharedEnum represents an enum definition.
type SharedEnum struct {
	Name   string
	Values []EnumValue
}

// EnumValue represents an enum value.
type EnumValue struct {
	Name   string
	Number int
}

// ImportGraph tracks proto file dependencies.
type ImportGraph struct {
	nodes map[string]*ImportNode
	mu    sync.RWMutex
}

// ImportNode represents a node in the import graph.
type ImportNode struct {
	Path         string
	Dependencies []string
	Dependents   []string
}

// NewSharedResourceResolver creates a new resolver.
func NewSharedResourceResolver(config Config, logger zerolog.Logger) *SharedResourceResolver {
	if len(config.CommonDirPatterns) == 0 {
		config.CommonDirPatterns = []string{
			"common",
			"shared",
			"proto/common",
			"proto/shared",
			"api/common",
			"api/shared",
		}
	}

	return &SharedResourceResolver{
		config:         config,
		logger:         logger,
		sharedMessages: make(map[string]*docgen.Message),
		sharedEnums:    make(map[string]*SharedEnum),
		protoCache:     make(map[string]*ProtoFile),
		importGraph:    NewImportGraph(),
	}
}

// NewImportGraph creates a new import graph.
func NewImportGraph() *ImportGraph {
	return &ImportGraph{
		nodes: make(map[string]*ImportNode),
	}
}

// ResolveService resolves shared resources for a service.
func (r *SharedResourceResolver) ResolveService(service *docgen.Service, protoFiles []string) error {
	r.logger.Info().Str("service", service.Name).Msg("Resolving shared resources")

	// Discover shared proto files if enabled
	if r.config.AutoDiscover {
		discovered, err := r.discoverSharedProtos(protoFiles)
		if err != nil {
			r.logger.Warn().Err(err).Msg("Shared proto discovery failed")
		} else {
			r.logger.Info().Int("count", len(discovered)).Msg("Discovered shared protos")
		}
	}

	// Load and parse proto files
	for _, protoPath := range protoFiles {
		if err := r.loadProtoFile(protoPath); err != nil {
			r.logger.Warn().Err(err).Str("file", protoPath).Msg("Failed to load proto file")
			continue
		}
	}

	// Resolve imports and dependencies
	if err := r.resolveImports(service); err != nil {
		return fmt.Errorf("import resolution failed: %w", err)
	}

	// Deduplicate shared messages
	if r.config.DeduplicateMessages {
		r.deduplicateMessages(service)
	}

	// Add shared messages to service
	r.addSharedMessages(service)

	return nil
}

// ResolveAllServices resolves shared resources for multiple services.
func (r *SharedResourceResolver) ResolveAllServices(services []*docgen.Service) error {
	r.logger.Info().Int("count", len(services)).Msg("Resolving shared resources for all services")

	// First pass: collect all proto files
	allProtoFiles := make(map[string]bool)
	for _, service := range services {
		for _, protoFile := range service.ProtoFiles {
			allProtoFiles[protoFile] = true
		}
	}

	// Discover shared protos
	if r.config.AutoDiscover {
		discovered, err := r.discoverSharedProtos(mapKeys(allProtoFiles))
		if err != nil {
			r.logger.Warn().Err(err).Msg("Shared proto discovery failed")
		} else {
			for _, path := range discovered {
				allProtoFiles[path] = true
			}
		}
	}

	// Load all proto files
	for protoFile := range allProtoFiles {
		if err := r.loadProtoFile(protoFile); err != nil {
			r.logger.Warn().Err(err).Str("file", protoFile).Msg("Failed to load proto file")
		}
	}

	// Resolve each service
	for _, service := range services {
		if err := r.ResolveService(service, service.ProtoFiles); err != nil {
			r.logger.Warn().Err(err).Str("service", service.Name).Msg("Service resolution failed")
		}
	}

	// Generate virtual services if enabled
	if r.config.GenerateVirtualServices {
		virtualServices := r.generateVirtualServices()
		r.logger.Info().Int("count", len(virtualServices)).Msg("Generated virtual services")
	}

	return nil
}

// HandleServiceWithoutProtos creates a service definition for services without proto files.
func (r *SharedResourceResolver) HandleServiceWithoutProtos(serviceName string) (*docgen.Service, error) {
	r.logger.Info().Str("service", serviceName).Msg("Handling service without proto files")

	// Try to find proto files that might belong to this service
	candidates := r.findProtoFilesForService(serviceName)

	if len(candidates) == 0 {
		// Generate a minimal service definition
		return r.generateMinimalService(serviceName), nil
	}

	// Load and parse candidate files
	service := &docgen.Service{
		Name:        serviceName,
		Package:     inferPackageName(serviceName),
		ProtoFiles:  candidates,
		Methods:     []docgen.Method{},
		Messages:    []docgen.Message{},
	}

	for _, protoFile := range candidates {
		if err := r.loadProtoFile(protoFile); err != nil {
			r.logger.Warn().Err(err).Str("file", protoFile).Msg("Failed to load proto file")
			continue
		}

		// Extract service definition from proto file
		if proto, ok := r.protoCache[protoFile]; ok {
			for _, protoSvc := range proto.Services {
				if strings.EqualFold(protoSvc.Name, serviceName) {
					service.Methods = append(service.Methods, protoSvc.Methods...)
				}
			}
			service.Messages = append(service.Messages, proto.Messages...)
		}
	}

	// Add shared resources
	r.addSharedMessages(service)

	return service, nil
}

// discoverSharedProtos discovers shared proto files.
func (r *SharedResourceResolver) discoverSharedProtos(knownProtos []string) ([]string, error) {
	discovered := make(map[string]bool)

	// Add configured shared proto paths
	for _, path := range r.config.SharedProtoPaths {
		protos, err := r.scanDirectory(path, ".proto")
		if err != nil {
			r.logger.Warn().Err(err).Str("path", path).Msg("Failed to scan shared proto path")
			continue
		}
		for _, proto := range protos {
			discovered[proto] = true
		}
	}

	// Look for common directory patterns
	for _, knownProto := range knownProtos {
		dir := filepath.Dir(knownProto)
		for _, pattern := range r.config.CommonDirPatterns {
			commonDir := filepath.Join(dir, "..", pattern)
			if protos, err := r.scanDirectory(commonDir, ".proto"); err == nil {
				for _, proto := range protos {
					discovered[proto] = true
				}
			}
		}
	}

	return mapKeys(discovered), nil
}

// loadProtoFile loads and parses a proto file.
func (r *SharedResourceResolver) loadProtoFile(path string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check cache
	if _, exists := r.protoCache[path]; exists {
		return nil
	}

	// Load file
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read proto file: %w", err)
	}

	// Parse proto file (basic parsing)
	protoFile := r.parseProtoFile(path, string(content))
	r.protoCache[path] = protoFile

	// Add to import graph
	r.importGraph.AddNode(path, protoFile.Imports)

	// Store shared messages
	for i := range protoFile.Messages {
		msg := &protoFile.Messages[i]
		key := fmt.Sprintf("%s.%s", protoFile.Package, msg.Name)
		r.sharedMessages[key] = msg
	}

	// Store shared enums
	for _, enum := range protoFile.Enums {
		key := fmt.Sprintf("%s.%s", protoFile.Package, enum.Name)
		r.sharedEnums[key] = enum
	}

	return nil
}

// parseProtoFile performs basic parsing of a proto file.
func (r *SharedResourceResolver) parseProtoFile(path, content string) *ProtoFile {
	proto := &ProtoFile{
		Path:     path,
		Messages: []docgen.Message{},
		Services: []*ProtoService{},
		Enums:    []*SharedEnum{},
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Extract package
		if strings.HasPrefix(line, "package ") {
			proto.Package = extractPackageName(line)
		}

		// Extract imports
		if strings.HasPrefix(line, "import ") {
			importPath := extractImportPath(line)
			if importPath != "" {
				proto.Imports = append(proto.Imports, importPath)
			}
		}

		// Basic message detection
		if strings.HasPrefix(line, "message ") {
			msgName := extractMessageName(line)
			if msgName != "" {
				proto.Messages = append(proto.Messages, docgen.Message{
					Name:        msgName,
					Description: fmt.Sprintf("Message %s from %s", msgName, filepath.Base(path)),
					Fields:      []docgen.Field{},
				})
			}
		}

		// Basic service detection
		if strings.HasPrefix(line, "service ") {
			svcName := extractServiceName(line)
			if svcName != "" {
				proto.Services = append(proto.Services, &ProtoService{
					Name:    svcName,
					Methods: []docgen.Method{},
				})
			}
		}
	}

	return proto
}

// resolveImports resolves import dependencies for a service.
func (r *SharedResourceResolver) resolveImports(service *docgen.Service) error {
	for _, protoFile := range service.ProtoFiles {
		deps := r.importGraph.GetDependencies(protoFile)
		for _, dep := range deps {
			if err := r.loadProtoFile(dep); err != nil {
				r.logger.Warn().Err(err).Str("import", dep).Msg("Failed to load import")
			}
		}
	}
	return nil
}

// deduplicateMessages removes duplicate messages across services.
func (r *SharedResourceResolver) deduplicateMessages(service *docgen.Service) {
	seen := make(map[string]bool)
	unique := []docgen.Message{}

	for _, msg := range service.Messages {
		key := fmt.Sprintf("%s.%s", service.Package, msg.Name)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, msg)
		}
	}

	service.Messages = unique
}

// addSharedMessages adds shared messages to a service.
func (r *SharedResourceResolver) addSharedMessages(service *docgen.Service) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, msg := range r.sharedMessages {
		// Check if message is relevant to this service
		if r.isMessageRelevant(service, msg) {
			service.Messages = append(service.Messages, *msg)
		}
	}
}

// isMessageRelevant checks if a message is relevant to a service.
func (r *SharedResourceResolver) isMessageRelevant(service *docgen.Service, msg *docgen.Message) bool {
	// Check if any method uses this message
	for _, method := range service.Methods {
		if strings.Contains(method.InputType, msg.Name) ||
			strings.Contains(method.OutputType, msg.Name) {
			return true
		}
	}

	// Check if message name suggests common use (e.g., Request, Response suffixes)
	return strings.HasSuffix(msg.Name, "Request") ||
		strings.HasSuffix(msg.Name, "Response") ||
		strings.HasPrefix(msg.Name, "Common")
}

// generateVirtualServices creates virtual services from shared proto files.
func (r *SharedResourceResolver) generateVirtualServices() []*docgen.Service {
	var virtualServices []*docgen.Service

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Group messages by package
	packageMessages := make(map[string][]docgen.Message)
	for _, proto := range r.protoCache {
		if len(proto.Services) == 0 && len(proto.Messages) > 0 {
			// This is a shared proto without services
			packageMessages[proto.Package] = append(packageMessages[proto.Package], proto.Messages...)
		}
	}

	// Create virtual services
	for pkg, messages := range packageMessages {
		if len(messages) > 0 {
			virtualServices = append(virtualServices, &docgen.Service{
				Name:        fmt.Sprintf("Shared_%s", sanitizeName(pkg)),
				Package:     pkg,
				Description: fmt.Sprintf("Shared messages and types for %s", pkg),
				Messages:    messages,
				Methods:     []docgen.Method{},
			})
		}
	}

	return virtualServices
}

// findProtoFilesForService finds proto files that might belong to a service.
func (r *SharedResourceResolver) findProtoFilesForService(serviceName string) []string {
	var candidates []string

	// Search in known proto files
	r.mu.RLock()
	defer r.mu.RUnlock()

	for path, proto := range r.protoCache {
		for _, svc := range proto.Services {
			if strings.EqualFold(svc.Name, serviceName) {
				candidates = append(candidates, path)
				break
			}
		}
	}

	return candidates
}

// generateMinimalService creates a minimal service definition.
func (r *SharedResourceResolver) generateMinimalService(serviceName string) *docgen.Service {
	return &docgen.Service{
		Name:        serviceName,
		Package:     inferPackageName(serviceName),
		Description: fmt.Sprintf("%s service (auto-generated)", serviceName),
		Methods: []docgen.Method{
			{
				Name:        "Get",
				Description: fmt.Sprintf("Retrieves %s data", serviceName),
				InputType:   fmt.Sprintf("Get%sRequest", serviceName),
				OutputType:  fmt.Sprintf("Get%sResponse", serviceName),
			},
		},
		Messages: []docgen.Message{
			{
				Name:        fmt.Sprintf("Get%sRequest", serviceName),
				Description: fmt.Sprintf("Request message for Get%s", serviceName),
				Fields: []docgen.Field{
					{Name: "id", Type: "string", Description: "Resource ID", Number: 1},
				},
			},
			{
				Name:        fmt.Sprintf("Get%sResponse", serviceName),
				Description: fmt.Sprintf("Response message for Get%s", serviceName),
				Fields: []docgen.Field{
					{Name: "data", Type: serviceName, Description: "Resource data", Number: 1},
				},
			},
		},
	}
}

// scanDirectory scans a directory for files with a specific extension.
func (r *SharedResourceResolver) scanDirectory(dir, ext string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}
		if !info.IsDir() && strings.HasSuffix(path, ext) {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// Import graph methods

// AddNode adds a node to the import graph.
func (ig *ImportGraph) AddNode(path string, imports []string) {
	ig.mu.Lock()
	defer ig.mu.Unlock()

	node := &ImportNode{
		Path:         path,
		Dependencies: imports,
		Dependents:   []string{},
	}

	ig.nodes[path] = node

	// Update dependents
	for _, imp := range imports {
		if depNode, exists := ig.nodes[imp]; exists {
			depNode.Dependents = append(depNode.Dependents, path)
		}
	}
}

// GetDependencies returns all dependencies for a path.
func (ig *ImportGraph) GetDependencies(path string) []string {
	ig.mu.RLock()
	defer ig.mu.RUnlock()

	if node, exists := ig.nodes[path]; exists {
		return node.Dependencies
	}
	return []string{}
}

// Helper functions

func mapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func extractPackageName(line string) string {
	// Extract from: package foo.bar.v1;
	parts := strings.Fields(line)
	if len(parts) >= 2 {
		pkg := strings.TrimSuffix(parts[1], ";")
		return strings.TrimSpace(pkg)
	}
	return ""
}

func extractImportPath(line string) string {
	// Extract from: import "path/to/file.proto";
	start := strings.Index(line, "\"")
	end := strings.LastIndex(line, "\"")
	if start != -1 && end != -1 && start < end {
		return line[start+1 : end]
	}
	return ""
}

func extractMessageName(line string) string {
	// Extract from: message Foo {
	parts := strings.Fields(line)
	if len(parts) >= 2 {
		return strings.TrimSuffix(parts[1], "{")
	}
	return ""
}

func extractServiceName(line string) string {
	// Extract from: service Foo {
	parts := strings.Fields(line)
	if len(parts) >= 2 {
		return strings.TrimSuffix(parts[1], "{")
	}
	return ""
}

func inferPackageName(serviceName string) string {
	// Convert ServiceName -> service.v1
	lower := strings.ToLower(serviceName)
	lower = strings.ReplaceAll(lower, "service", "")
	return lower + ".v1"
}

func sanitizeName(name string) string {
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, "/", "_")
	return name
}
