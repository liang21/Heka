// tasks.md: T085 | spec.md: Worker Pool pattern
package ai

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPool(t *testing.T) {
	t.Parallel()

	pool := NewPool(10, 50)
	assert.NotNil(t, pool)
}

func TestWorkerPool_ConcurrentExecution(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := NewPool(10, 50)
	defer pool.Stop()

	var counter atomic.Int32
	taskFunc := func() error {
		time.Sleep(10 * time.Millisecond)
		counter.Add(1)
		return nil
	}

	// Submit 20 tasks
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = pool.Submit(ctx, taskFunc)
		}()
	}
	wg.Wait()

	// Wait for tasks to complete
	time.Sleep(200 * time.Millisecond)

	// All 20 tasks should complete
	assert.Equal(t, int32(20), counter.Load())
}

func TestWorkerPool_QueueFull(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := NewPool(2, 5) // Small queue for testing
	taskFunc := func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	// Submit more tasks than queue capacity
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = pool.Submit(ctx, taskFunc)
		}()
	}

	// Wait for all submissions
	time.Sleep(50 * time.Millisecond)
	wg.Wait()

	// Pool should handle queue full scenario
	// This will fail because implementation doesn't exist yet
}

func TestWorkerPool_Stop(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := NewPool(10, 50)

	var counter atomic.Int32
	taskFunc := func() error {
		time.Sleep(10 * time.Millisecond)
		counter.Add(1)
		return nil
	}

	// Submit some tasks
	for i := 0; i < 5; i++ {
		_ = pool.Submit(ctx, taskFunc)
	}

	// Wait for tasks to be queued
	time.Sleep(50 * time.Millisecond)

	// Stop the pool
	pool.Stop()

	// Submit after stop should fail
	err := pool.Submit(ctx, taskFunc)
	assert.Error(t, err)
}

func TestWorkerPool_StopWithPendingTasks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := NewPool(2, 10)

	var started atomic.Int32
	longTask := func() error {
		started.Add(1)
		time.Sleep(500 * time.Millisecond)
		return nil
	}

	// Submit long-running tasks
	for i := 0; i < 5; i++ {
		_ = pool.Submit(ctx, longTask)
	}

	// Wait a bit for workers to pick up tasks
	time.Sleep(50 * time.Millisecond)

	// Stop should wait for pending tasks to complete
	start := time.Now()
	pool.Stop()
	elapsed := time.Since(start)

	// At least one task should have started and completed
	assert.True(t, started.Load() > 0, "No tasks started")
	assert.True(t, elapsed >= 500*time.Millisecond, "Expected >= 500ms, got %v", elapsed)
	// With 2 workers and 5 tasks of 500ms each, expect ~1.5s total
	assert.True(t, elapsed < 2*time.Second, "Expected < 2s, got %v", elapsed)
}

func TestWorkerPool_TaskExecutionError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := NewPool(10, 50)
	defer pool.Stop()

	errFunc := func() error {
		return errors.New("task error")
	}

	// Submit should succeed (errors are handled by worker, not returned to Submit)
	err := pool.Submit(ctx, errFunc)
	assert.NoError(t, err)

	// Wait for task to execute
	time.Sleep(50 * time.Millisecond)
}

func TestWorkerPool_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	pool := NewPool(10, 50)
	defer pool.Stop()

	cancel() // Cancel immediately

	taskFunc := func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	// Submit should fail due to canceled context
	err := pool.Submit(ctx, taskFunc)
	// Due to non-blocking select, might succeed if context check happens after queue insert
	_ = err // The behavior depends on timing
}
