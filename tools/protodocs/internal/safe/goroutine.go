package safe

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"sync"
	"time"
)

// PanicHandler is called when a panic is recovered.
type PanicHandler func(interface{}, []byte)

// DefaultPanicHandler logs the panic with stack trace.
func DefaultPanicHandler(recovered interface{}, stack []byte) {
	log.Printf("PANIC RECOVERED: %v\nStack trace:\n%s", recovered, stack)
}

// Go runs a function in a goroutine with panic recovery.
func Go(fn func(), handler PanicHandler) {
	if handler == nil {
		handler = DefaultPanicHandler
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				handler(r, debug.Stack())
			}
		}()
		fn()
	}()
}

// GoWithContext runs a function in a goroutine with context and panic recovery.
func GoWithContext(ctx context.Context, fn func(context.Context), handler PanicHandler) {
	if handler == nil {
		handler = DefaultPanicHandler
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				handler(r, debug.Stack())
			}
		}()
		fn(ctx)
	}()
}

// Pool represents a worker pool that safely executes tasks.
type Pool struct {
	workers       int
	tasks         chan func()
	wg            sync.WaitGroup
	panicHandler  PanicHandler
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewPool creates a new worker pool.
func NewPool(workers int, handler PanicHandler) *Pool {
	if workers <= 0 {
		workers = 1
	}
	if handler == nil {
		handler = DefaultPanicHandler
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &Pool{
		workers:      workers,
		tasks:        make(chan func(), workers*2),
		panicHandler: handler,
		ctx:          ctx,
		cancel:       cancel,
	}

	pool.start()
	return pool
}

// start starts the worker goroutines.
func (p *Pool) start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		Go(func() {
			defer p.wg.Done()
			for {
				select {
				case <-p.ctx.Done():
					return
				case task, ok := <-p.tasks:
					if !ok {
						return
					}
					p.executeTask(task)
				}
			}
		}, p.panicHandler)
	}
}

// executeTask executes a task with panic recovery.
func (p *Pool) executeTask(task func()) {
	defer func() {
		if r := recover(); r != nil {
			p.panicHandler(r, debug.Stack())
		}
	}()
	task()
}

// Submit submits a task to the pool.
func (p *Pool) Submit(task func()) error {
	select {
	case <-p.ctx.Done():
		return fmt.Errorf("pool is closed")
	case p.tasks <- task:
		return nil
	}
}

// SubmitWithTimeout submits a task with a timeout.
func (p *Pool) SubmitWithTimeout(task func(), timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(p.ctx, timeout)
	defer cancel()

	select {
	case <-ctx.Done():
		return fmt.Errorf("submit timeout: %w", ctx.Err())
	case p.tasks <- task:
		return nil
	}
}

// Close gracefully shuts down the pool.
func (p *Pool) Close() {
	p.cancel()
	close(p.tasks)
	p.wg.Wait()
}

// CloseWithTimeout closes the pool with a timeout.
func (p *Pool) CloseWithTimeout(timeout time.Duration) error {
	done := make(chan struct{})
	go func() {
		p.Close()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("pool close timeout after %v", timeout)
	}
}

// Retry executes a function with retry logic and panic recovery.
func Retry(fn func() error, maxRetries int, backoff time.Duration, handler PanicHandler) error {
	if handler == nil {
		handler = DefaultPanicHandler
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(backoff * time.Duration(attempt))
		}

		err := func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					handler(r, debug.Stack())
					err = fmt.Errorf("panic recovered: %v", r)
				}
			}()
			return fn()
		}()

		if err == nil {
			return nil
		}

		lastErr = err
	}

	return fmt.Errorf("all %d retries failed: %w", maxRetries, lastErr)
}

// WaitGroup is a safe wrapper around sync.WaitGroup with panic recovery.
type WaitGroup struct {
	wg           sync.WaitGroup
	panicHandler PanicHandler
}

// NewWaitGroup creates a new safe WaitGroup.
func NewWaitGroup(handler PanicHandler) *WaitGroup {
	if handler == nil {
		handler = DefaultPanicHandler
	}
	return &WaitGroup{
		panicHandler: handler,
	}
}

// Go runs a function in a goroutine with panic recovery.
func (w *WaitGroup) Go(fn func()) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				w.panicHandler(r, debug.Stack())
			}
		}()
		fn()
	}()
}

// GoWithContext runs a function in a goroutine with context and panic recovery.
func (w *WaitGroup) GoWithContext(ctx context.Context, fn func(context.Context)) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				w.panicHandler(r, debug.Stack())
			}
		}()
		fn(ctx)
	}()
}

// Wait waits for all goroutines to complete.
func (w *WaitGroup) Wait() {
	w.wg.Wait()
}

// WaitWithTimeout waits for all goroutines with a timeout.
func (w *WaitGroup) WaitWithTimeout(timeout time.Duration) error {
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("wait timeout after %v", timeout)
	}
}
