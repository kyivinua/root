package protoctx

import (
	"fmt"
	"io"
	"os"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// MergeStrategy defines how to handle conflicts when loading multiple descriptor sets.
type MergeStrategy int

const (
	// FailOnConflict returns an error if duplicate file names are found.
	FailOnConflict MergeStrategy = iota

	// LastWins keeps the last version of duplicate files.
	LastWins
)

// LoadFromDescriptorSetBytes loads a Context from FileDescriptorSet bytes.
//
// This is the primary loading method. It:
//   1. Unmarshals the FileDescriptorSet protobuf
//   2. Registers all files into protoregistry.Files
//   3. Builds FileIndex for fast lookup
//   4. Extracts comments into CommentsIndex
//
// Returns an error if unmarshaling fails or if file registration fails.
func LoadFromDescriptorSetBytes(data []byte) (*Context, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty descriptor set data")
	}

	var fds descriptorpb.FileDescriptorSet
	if err := proto.Unmarshal(data, &fds); err != nil {
		return nil, fmt.Errorf("unmarshal FileDescriptorSet: %w", err)
	}

	return loadFromFileDescriptorSet(&fds)
}

// LoadFromDescriptorSetReader loads a Context from a reader.
func LoadFromDescriptorSetReader(r io.Reader) (*Context, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read descriptor set: %w", err)
	}

	return LoadFromDescriptorSetBytes(data)
}

// LoadFromDescriptorSetFile loads a Context from a file path.
//
// This is the most common method for runtime services.
// Example:
//
//	ctx, err := protoctx.LoadFromDescriptorSetFile("api-docs/descriptors/image.bin")
func LoadFromDescriptorSetFile(path string) (*Context, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", path, err)
	}

	return LoadFromDescriptorSetBytes(data)
}

// LoadFromMultipleDescriptorSetsBytes loads and merges multiple descriptor sets.
//
// This is useful when you have multiple independent FileDescriptorSets
// (e.g., from different modules) that need to be combined.
//
// The merge strategy determines how conflicts are handled:
//   - FailOnConflict: returns error if duplicate files are found
//   - LastWins: keeps the last version of duplicate files
func LoadFromMultipleDescriptorSetsBytes(list [][]byte, merge MergeStrategy) (*Context, error) {
	if len(list) == 0 {
		return nil, fmt.Errorf("no descriptor sets provided")
	}

	var allFiles []*descriptorpb.FileDescriptorProto
	fileMap := make(map[string]*descriptorpb.FileDescriptorProto) // for conflict detection

	for i, data := range list {
		var fds descriptorpb.FileDescriptorSet
		if err := proto.Unmarshal(data, &fds); err != nil {
			return nil, fmt.Errorf("unmarshal descriptor set %d: %w", i, err)
		}

		for _, fdProto := range fds.File {
			name := fdProto.GetName()
			if _, exists := fileMap[name]; exists {
				if merge == FailOnConflict {
					return nil, fmt.Errorf("duplicate file %s in descriptor set %d", name, i)
				}
				// LastWins: replace existing
				for j, f := range allFiles {
					if f.GetName() == name {
						allFiles[j] = fdProto
						break
					}
				}
			} else {
				allFiles = append(allFiles, fdProto)
			}
			fileMap[name] = fdProto
		}
	}

	fds := &descriptorpb.FileDescriptorSet{File: allFiles}
	return loadFromFileDescriptorSet(fds)
}

// loadFromFileDescriptorSet is the internal loading logic.
func loadFromFileDescriptorSet(fds *descriptorpb.FileDescriptorSet) (*Context, error) {
	ctx := NewContext()

	// Register all files into protoregistry.Files
	files, err := protodesc.NewFiles(fds)
	if err != nil {
		return nil, fmt.Errorf("create file registry: %w", err)
	}

	ctx.files = files

	// Build FileIndex
	files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		ctx.fileIdx.Add(fd)
		return true
	})

	// Extract comments from SourceCodeInfo
	for _, fdProto := range fds.File {
		if err := ctx.comments.AddFile(fdProto, files); err != nil {
			// Log warning but don't fail - comments are optional
			// In production, you'd use a proper logger here
			_ = err
		}
	}

	return ctx, nil
}
