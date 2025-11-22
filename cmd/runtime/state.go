package main

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/kyivinua/docgen-tool/tools/protoctx"
)

// CtxStatus represents the state of the runtime context.
type CtxStatus int32

const (
	// StatusUninitialized is the initial state before loading.
	StatusUninitialized CtxStatus = iota

	// StatusLoading is the state during descriptor loading.
	StatusLoading

	// StatusReady is the state when the context is loaded and ready.
	StatusReady

	// StatusError is the state when loading failed.
	StatusError

	// StatusReloading is the state during hot reload.
	StatusReloading
)

func (s CtxStatus) String() string {
	switch s {
	case StatusUninitialized:
		return "UNINITIALIZED"
	case StatusLoading:
		return "LOADING"
	case StatusReady:
		return "READY"
	case StatusError:
		return "ERROR"
	case StatusReloading:
		return "RELOADING"
	default:
		return "UNKNOWN"
	}
}

// RuntimeState manages the runtime context with FSM (Finite State Machine).
//
// State transitions:
//   UNINITIALIZED -> LOADING -> READY
//   LOADING -> ERROR
//   READY -> RELOADING -> READY
//   RELOADING -> ERROR
//
// All operations are thread-safe.
type RuntimeState struct {
	mu     sync.RWMutex
	ctx    *protoctx.Context
	status atomic.Int32 // CtxStatus
	err    error
}

// NewRuntimeState creates a new RuntimeState.
func NewRuntimeState() *RuntimeState {
	rs := &RuntimeState{}
	rs.status.Store(int32(StatusUninitialized))
	return rs
}

// Init initializes the runtime context by loading descriptors from a file.
//
// This transitions the state from UNINITIALIZED -> LOADING -> READY (or ERROR).
func (rs *RuntimeState) Init(imagePath string) {
	rs.setStatus(StatusLoading, nil)

	ctx, err := protoctx.LoadFromDescriptorSetFile(imagePath)
	if err != nil {
		rs.setStatus(StatusError, fmt.Errorf("load descriptor set: %w", err))
		return
	}

	rs.mu.Lock()
	rs.ctx = ctx
	rs.mu.Unlock()

	rs.setStatus(StatusReady, nil)
}

// Reload reloads the runtime context from a file (hot reload).
//
// This transitions the state from READY -> RELOADING -> READY (or ERROR).
// The old context remains usable until the new one is successfully loaded.
func (rs *RuntimeState) Reload(imagePath string) {
	currentStatus := CtxStatus(rs.status.Load())
	if currentStatus != StatusReady {
		// Can only reload from READY state
		return
	}

	rs.setStatus(StatusReloading, nil)

	ctx, err := protoctx.LoadFromDescriptorSetFile(imagePath)
	if err != nil {
		rs.setStatus(StatusError, fmt.Errorf("reload descriptor set: %w", err))
		return
	}

	// Atomic swap of context
	rs.mu.Lock()
	rs.ctx = ctx
	rs.mu.Unlock()

	rs.setStatus(StatusReady, nil)
}

// WithContext executes a function with a snapshot of the current context.
//
// This is the primary way to access the context for request handling.
// It ensures that the context is in READY state before executing the function.
//
// Example:
//
//	err := rs.WithContext(func(ctx *protoctx.Context) error {
//	    jsonData, err := ctx.MarshalWireToJSON(typeName, wire)
//	    // ... use jsonData
//	    return err
//	})
func (rs *RuntimeState) WithContext(fn func(*protoctx.Context) error) error {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	status := CtxStatus(rs.status.Load())
	if status != StatusReady {
		return fmt.Errorf("context not ready (status: %s)", status)
	}

	if rs.ctx == nil {
		return fmt.Errorf("context is nil")
	}

	return fn(rs.ctx)
}

// Status returns the current status of the runtime state.
func (rs *RuntimeState) Status() CtxStatus {
	return CtxStatus(rs.status.Load())
}

// Error returns the last error that occurred.
func (rs *RuntimeState) Error() error {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.err
}

// IsReady returns true if the context is in READY state.
func (rs *RuntimeState) IsReady() bool {
	return rs.Status() == StatusReady
}

// setStatus sets the current status and error.
func (rs *RuntimeState) setStatus(status CtxStatus, err error) {
	rs.status.Store(int32(status))

	rs.mu.Lock()
	rs.err = err
	rs.mu.Unlock()
}
