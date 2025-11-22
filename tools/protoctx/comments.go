package protoctx

import (
	"fmt"
	"strings"
	"sync"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

// Comments holds the comments for a descriptor.
//
// Comments are extracted from SourceCodeInfo.Location in the FileDescriptorProto.
type Comments struct {
	// Leading is the main comment block above the element.
	Leading string

	// Trailing is the comment on the same line after the element.
	Trailing string

	// LeadingDetached are detached comment blocks before the element.
	LeadingDetached []string
}

// IsEmpty returns true if all comment fields are empty.
func (c Comments) IsEmpty() bool {
	return c.Leading == "" && c.Trailing == "" && len(c.LeadingDetached) == 0
}

// Summary returns the first line of the leading comment (up to 160 chars).
//
// This is useful for generating brief descriptions in documentation.
func (c Comments) Summary() string {
	if c.Leading == "" {
		return ""
	}

	lines := strings.Split(strings.TrimSpace(c.Leading), "\n")
	if len(lines) == 0 {
		return ""
	}

	summary := strings.TrimSpace(lines[0])
	if len(summary) > 160 {
		summary = summary[:157] + "..."
	}
	return summary
}

// Description returns the full leading comment.
func (c Comments) Description() string {
	return strings.TrimSpace(c.Leading)
}

// CommentsIndex indexes comments by FQN (Fully Qualified Name) and path.
//
// The path format is: "file_name:path[0].path[1]..."
// The FQN format is: "package.Message" or "package.Service.Method"
//
// All methods are thread-safe.
type CommentsIndex struct {
	mu    sync.RWMutex
	byFQN map[string]Comments
}

// NewCommentsIndex creates a new empty CommentsIndex.
func NewCommentsIndex() *CommentsIndex {
	return &CommentsIndex{
		byFQN: make(map[string]Comments),
	}
}

// AddFile extracts comments from a FileDescriptorProto and adds them to the index.
//
// It processes SourceCodeInfo.Location entries and associates comments with
// their corresponding descriptors by computing the FQN.
func (idx *CommentsIndex) AddFile(fdProto *descriptorpb.FileDescriptorProto, files *protoregistry.Files) error {
	if fdProto.SourceCodeInfo == nil {
		return nil // No source info, nothing to extract
	}

	fileName := fdProto.GetName()
	fd, err := files.FindFileByPath(fileName)
	if err != nil {
		return fmt.Errorf("find file %s: %w", fileName, err)
	}

	pkg := string(fd.Package())

	for _, loc := range fdProto.SourceCodeInfo.Location {
		if len(loc.Path) == 0 {
			continue
		}

		comments := Comments{
			Leading:         strings.TrimSpace(loc.GetLeadingComments()),
			Trailing:        strings.TrimSpace(loc.GetTrailingComments()),
			LeadingDetached: make([]string, len(loc.LeadingDetachedComments)),
		}

		for i, detached := range loc.LeadingDetachedComments {
			comments.LeadingDetached[i] = strings.TrimSpace(detached)
		}

		if comments.IsEmpty() {
			continue
		}

		// Compute FQN from path
		fqn := idx.computeFQN(fd, loc.Path, pkg)
		if fqn != "" {
			idx.mu.Lock()
			idx.byFQN[fqn] = comments
			idx.mu.Unlock()
		}
	}

	return nil
}

// GetByFQN returns comments for a fully qualified name.
//
// Example FQNs:
//   - "company.user.v1.UserService"
//   - "company.user.v1.UserService.CreateUser"
//   - "company.user.v1.UserProfile"
//   - "company.user.v1.UserProfile.id"
func (idx *CommentsIndex) GetByFQN(fqn string) (Comments, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	c, ok := idx.byFQN[fqn]
	return c, ok
}

// GetByDescriptor returns comments for a descriptor.
//
// This is a convenience method that computes the FQN from the descriptor.
func (idx *CommentsIndex) GetByDescriptor(d protoreflect.Descriptor) (Comments, bool) {
	fqn := string(d.FullName())
	return idx.GetByFQN(fqn)
}

// computeFQN computes the fully qualified name from a SourceCodeInfo.Location path.
//
// The path is a sequence of field numbers and indexes that identifies the location
// in the FileDescriptorProto tree structure.
//
// See: https://github.com/protocolbuffers/protobuf/blob/main/src/google/protobuf/descriptor.proto
func (idx *CommentsIndex) computeFQN(fd protoreflect.FileDescriptor, path []int32, pkg string) string {
	if len(path) == 0 {
		return ""
	}

	// path[0] indicates the top-level element type:
	// 4 = message_type
	// 5 = enum_type
	// 6 = service
	// 8 = extension

	const (
		filePackageTag    = 2
		fileMessageTag    = 4
		fileEnumTag       = 5
		fileServiceTag    = 6
		fileExtensionTag  = 8
		messageFieldTag   = 2
		messageNestedTag  = 3
		messageEnumTag    = 4
		messageOneofTag   = 8
		serviceMethodTag  = 2
		enumValueTag      = 2
	)

	switch path[0] {
	case filePackageTag:
		return pkg

	case fileMessageTag:
		if len(path) < 2 {
			return ""
		}
		msg := fd.Messages().Get(int(path[1]))
		return idx.computeMessageFQN(msg, path[2:])

	case fileEnumTag:
		if len(path) < 2 {
			return ""
		}
		enum := fd.Enums().Get(int(path[1]))
		return idx.computeEnumFQN(enum, path[2:])

	case fileServiceTag:
		if len(path) < 2 {
			return ""
		}
		service := fd.Services().Get(int(path[1]))
		return idx.computeServiceFQN(service, path[2:])

	default:
		return ""
	}
}

func (idx *CommentsIndex) computeMessageFQN(msg protoreflect.MessageDescriptor, path []int32) string {
	fqn := string(msg.FullName())

	if len(path) == 0 {
		return fqn
	}

	const (
		messageFieldTag  = 2
		messageNestedTag = 3
		messageEnumTag   = 4
		messageOneofTag  = 8
	)

	switch path[0] {
	case messageFieldTag:
		if len(path) < 2 {
			return fqn
		}
		field := msg.Fields().Get(int(path[1]))
		return string(field.FullName())

	case messageNestedTag:
		if len(path) < 2 {
			return fqn
		}
		nested := msg.Messages().Get(int(path[1]))
		return idx.computeMessageFQN(nested, path[2:])

	case messageEnumTag:
		if len(path) < 2 {
			return fqn
		}
		enum := msg.Enums().Get(int(path[1]))
		return idx.computeEnumFQN(enum, path[2:])

	case messageOneofTag:
		if len(path) < 2 {
			return fqn
		}
		oneof := msg.Oneofs().Get(int(path[1]))
		return string(oneof.FullName())

	default:
		return fqn
	}
}

func (idx *CommentsIndex) computeEnumFQN(enum protoreflect.EnumDescriptor, path []int32) string {
	fqn := string(enum.FullName())

	if len(path) == 0 {
		return fqn
	}

	const enumValueTag = 2

	if path[0] == enumValueTag && len(path) >= 2 {
		value := enum.Values().Get(int(path[1]))
		return string(value.FullName())
	}

	return fqn
}

func (idx *CommentsIndex) computeServiceFQN(service protoreflect.ServiceDescriptor, path []int32) string {
	fqn := string(service.FullName())

	if len(path) == 0 {
		return fqn
	}

	const serviceMethodTag = 2

	if path[0] == serviceMethodTag && len(path) >= 2 {
		method := service.Methods().Get(int(path[1]))
		return string(method.FullName())
	}

	return fqn
}
