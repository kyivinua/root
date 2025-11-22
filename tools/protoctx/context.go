// Package protoctx provides runtime Context for working with Protocol Buffer descriptors.
//
// This package implements the ProtoContext system for loading and working with
// FileDescriptorSet at runtime, providing reflection, JSON/wire conversion,
// comments extraction, and schema compatibility checking.
package protoctx

import (
	"sync"

	"google.golang.org/protobuf/reflect/protoregistry"
)

// Context holds the runtime context for Protocol Buffer descriptors.
//
// It encapsulates:
//   - protoregistry.Files: registry of all loaded file descriptors
//   - FileIndex: fast lookup of files by name and package
//   - CommentsIndex: comments extracted from SourceCodeInfo
//
// All operations are thread-safe for concurrent reads.
type Context struct {
	files    *protoregistry.Files
	fileIdx  *FileIndex
	comments *CommentsIndex

	// mu protects context-level operations (not individual reads)
	mu sync.RWMutex
}

// NewContext creates a new empty Context.
func NewContext() *Context {
	return &Context{
		files:    new(protoregistry.Files),
		fileIdx:  NewFileIndex(),
		comments: NewCommentsIndex(),
	}
}

// Files returns the underlying protoregistry.Files.
//
// The returned Files can be used directly for advanced operations.
// Modifications to Files should be done carefully as they may affect
// the consistency of FileIndex and CommentsIndex.
func (c *Context) Files() *protoregistry.Files {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.files
}

// FileIndex returns the file index for fast lookup.
func (c *Context) FileIndex() *FileIndex {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.fileIdx
}

// CommentsIndex returns the comments index.
func (c *Context) CommentsIndex() *CommentsIndex {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.comments
}
