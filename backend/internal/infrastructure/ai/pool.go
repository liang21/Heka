// tasks.md: T086 | spec.md: Worker Pool pattern
package ai

import (
	"context"
	"errors"
	"log"
	"sync"
	"sync/atomic"
)

var (
	ErrPoolStopped = errors.New("worker pool is stopped")
)

// Task represents a unit of work
type Task func() error

// ErrorHandler is called when a task returns an error
type ErrorHandler func(error)

// Pool implements a worker pool pattern
type Pool struct {
	wg             sync.WaitGroup
	queue          chan Task
	stop           chan struct{}
	stopped        atomic.Bool
	errorHandler   ErrorHandler
}

// NewPool creates a new worker pool
func NewPool(workers int, queueSize int) *Pool {
	return NewPoolWithErrorHandler(workers, queueSize, nil)
}

// NewPoolWithErrorHandler creates a new worker pool with custom error handler
func NewPoolWithErrorHandler(workers int, queueSize int, errorHandler ErrorHandler) *Pool {
	p := &Pool{
		queue:        make(chan Task, queueSize),
		stop:         make(chan struct{}),
		errorHandler: errorHandler,
	}

	// Start workers
	for i := 0; i < workers; i++ {
		p.wg.Add(1)
		go p.worker(context.Background())
	}

	return p
}

// worker processes tasks from the queue
func (p *Pool) worker(ctx context.Context) {
	defer p.wg.Done()

	for {
		select {
		case task, ok := <-p.queue:
			if !ok {
				return
			}
			// Execute task and handle error
			if err := task(); err != nil {
				if p.errorHandler != nil {
					p.errorHandler(err)
				} else {
					// Default: log to stderr
					log.Printf("worker pool task error: %v", err)
				}
			}
		case <-p.stop:
			// Drain remaining tasks from queue before exiting
			for {
				select {
				case task, ok := <-p.queue:
					if !ok {
						return
					}
					if err := task(); err != nil {
						if p.errorHandler != nil {
							p.errorHandler(err)
						}
					}
				default:
					return
				}
			}
		case <-ctx.Done():
			// Context cancelled - exit worker
			return
		}
	}
}

// Submit submits a task to the pool
func (p *Pool) Submit(ctx context.Context, task Task) error {
	if p.stopped.Load() {
		return ErrPoolStopped
	}

	select {
	case p.queue <- task:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-p.stop:
		return ErrPoolStopped
	}
}

// Stop gracefully shuts down the pool
func (p *Pool) Stop() {
	if !p.stopped.CompareAndSwap(false, true) {
		return
	}

	close(p.stop)
	p.wg.Wait()
	close(p.queue)
}
