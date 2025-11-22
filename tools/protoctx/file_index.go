package protoctx

import (
	"sync"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// FileInfo holds metadata about a .proto file.
type FileInfo struct {
	Name    string   // file name (e.g., "user/v1/user_service.proto")
	Package string   // package name (e.g., "company.user.v1")
	Syntax  string   // "proto2" or "proto3"
	Imports []string // list of imported file names
}

// FileIndex provides fast lookup of files by name and package.
//
// It maintains two indexes:
//   - byName: file name → FileInfo
//   - byPackage: package name → list of FileInfo
//
// All methods are thread-safe.
type FileIndex struct {
	mu        sync.RWMutex
	byName    map[string]FileInfo
	byPackage map[string][]FileInfo
}

// NewFileIndex creates a new empty FileIndex.
func NewFileIndex() *FileIndex {
	return &FileIndex{
		byName:    make(map[string]FileInfo),
		byPackage: make(map[string][]FileInfo),
	}
}

// Add adds a file descriptor to the index.
func (idx *FileIndex) Add(fd protoreflect.FileDescriptor) {
	info := FileInfo{
		Name:    fd.Path(),
		Package: string(fd.Package()),
		Syntax:  string(fd.Syntax()),
		Imports: make([]string, fd.Imports().Len()),
	}

	for i := 0; i < fd.Imports().Len(); i++ {
		info.Imports[i] = fd.Imports().Get(i).Path()
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.byName[info.Name] = info
	idx.byPackage[info.Package] = append(idx.byPackage[info.Package], info)
}

// Get returns file info by name.
func (idx *FileIndex) Get(name string) (FileInfo, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	info, ok := idx.byName[name]
	return info, ok
}

// List returns all files in the index.
func (idx *FileIndex) List() []FileInfo {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	result := make([]FileInfo, 0, len(idx.byName))
	for _, info := range idx.byName {
		result = append(result, info)
	}
	return result
}

// ListByPackage returns all files in a package.
func (idx *FileIndex) ListByPackage(pkg string) []FileInfo {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	result := idx.byPackage[pkg]
	// Return a copy to avoid external modification
	cp := make([]FileInfo, len(result))
	copy(cp, result)
	return cp
}

// ListPackages returns all package names in the index.
func (idx *FileIndex) ListPackages() []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	result := make([]string, 0, len(idx.byPackage))
	for pkg := range idx.byPackage {
		result = append(result, pkg)
	}
	return result
}
