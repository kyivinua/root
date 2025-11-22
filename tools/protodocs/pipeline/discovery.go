package pipeline

import (
	"fmt"
	"os/exec"
	"strings"
)

// Scope represents the scope of proto files to process.
type Scope struct {
	ProtoFiles    []string // paths to changed .proto files
	ProtoPackages []string // affected packages
}

// DiscoverChangedProto discovers changed .proto files between two git refs.
//
// This is used to implement incremental documentation generation for PRs.
//
// Args:
//   - protoRoot: root directory for proto files (e.g., "./proto")
//   - baseRef: base git ref to compare against (e.g., "main", "HEAD~1")
//   - headRef: head git ref (e.g., "HEAD")
//
// Returns:
//   - Scope with changed files and affected packages
func DiscoverChangedProto(protoRoot, baseRef, headRef string) (*Scope, error) {
	// Run git diff to find changed files
	cmd := exec.Command("git", "diff", "--name-only", fmt.Sprintf("%s..%s", baseRef, headRef))
	output, err := cmd.Output()
	if err != nil {
		// If git diff fails, fall back to full scope
		return DiscoverAllProto(protoRoot)
	}

	lines := strings.Split(string(output), "\n")
	scope := &Scope{
		ProtoFiles:    []string{},
		ProtoPackages: []string{},
	}

	packageSet := make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Filter only .proto files under protoRoot
		if strings.HasPrefix(line, protoRoot+"/") && strings.HasSuffix(line, ".proto") {
			scope.ProtoFiles = append(scope.ProtoFiles, line)

			// Extract package from path
			// e.g., "proto/user/v1/user_service.proto" → "user.v1"
			pkg := extractPackageFromPath(line, protoRoot)
			if pkg != "" && !packageSet[pkg] {
				scope.ProtoPackages = append(scope.ProtoPackages, pkg)
				packageSet[pkg] = true
			}
		}
	}

	// If no proto files changed, return empty scope
	if len(scope.ProtoFiles) == 0 {
		return scope, nil
	}

	return scope, nil
}

// DiscoverAllProto discovers all .proto files in the proto root.
//
// This is used for full documentation generation (e.g., on main branch).
func DiscoverAllProto(protoRoot string) (*Scope, error) {
	// Use find command to list all .proto files
	cmd := exec.Command("find", protoRoot, "-name", "*.proto", "-type", "f")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("find proto files: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	scope := &Scope{
		ProtoFiles:    []string{},
		ProtoPackages: []string{},
	}

	packageSet := make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		scope.ProtoFiles = append(scope.ProtoFiles, line)

		pkg := extractPackageFromPath(line, protoRoot)
		if pkg != "" && !packageSet[pkg] {
			scope.ProtoPackages = append(scope.ProtoPackages, pkg)
			packageSet[pkg] = true
		}
	}

	return scope, nil
}

// extractPackageFromPath extracts a package name from a file path.
//
// Example:
//   - "proto/user/v1/user_service.proto" → "user.v1"
//   - "proto/billing/v1/billing.proto" → "billing.v1"
//
// This is a heuristic and may not work for all path structures.
// Ideally, we should parse the .proto file to get the actual package name.
func extractPackageFromPath(path, protoRoot string) string {
	// Remove proto root prefix
	if !strings.HasPrefix(path, protoRoot+"/") {
		return ""
	}

	relative := strings.TrimPrefix(path, protoRoot+"/")

	// Split by / and take all but the last component (filename)
	parts := strings.Split(relative, "/")
	if len(parts) < 2 {
		return ""
	}

	// Join parts with . to form package name
	// e.g., ["user", "v1", "user_service.proto"] → "user.v1"
	pkgParts := parts[:len(parts)-1]
	return strings.Join(pkgParts, ".")
}

// IsEmpty returns true if the scope is empty (no files to process).
func (s *Scope) IsEmpty() bool {
	return len(s.ProtoFiles) == 0
}

// String returns a human-readable representation of the scope.
func (s *Scope) String() string {
	return fmt.Sprintf("Scope{files: %d, packages: %d}", len(s.ProtoFiles), len(s.ProtoPackages))
}
