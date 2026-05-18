// tasks.md: T080 | spec.md: Circuit Breaker pattern
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

func TestNewCircuitBreaker(t *testing.T) {
	t.Parallel()

	breaker := NewCircuitBreaker(5, 30*time.Second)
	assert.NotNil(t, breaker)
}

func TestCircuitBreaker_ClosedState(t *testing.T) {
	t.Parallel()

	breaker := NewCircuitBreaker(5, 100*time.Millisecond)
	ctx := context.Background()

	// In closed state, all requests should pass
	for i := 0; i < 10; i++ {
		err := breaker.Execute(ctx, func() error {
			return nil
		})
		assert.NoError(t, err)
	}
}

func TestCircuitBreaker_OpenAfterConsecutiveFailures(t *testing.T) {
	t.Parallel()

	breaker := NewCircuitBreaker(3, 100*time.Millisecond)
	ctx := context.Background()

	// Simulate 3 consecutive failures
	failFunc := func() error { return errors.New("test error") }

	for i := 0; i < 3; i++ {
		_ = breaker.Execute(ctx, failFunc)
	}

	// Next request should fail immediately (circuit is open)
	err := breaker.Execute(ctx, func() error { return nil })
	assert.Error(t, err)
}

func TestCircuitBreaker_HalfOpen(t *testing.T) {
	t.Parallel()

	breaker := NewCircuitBreaker(3, 50*time.Millisecond)
	ctx := context.Background()

	// Trip the breaker
	failFunc := func() error { return errors.New("fail") }
	for i := 0; i < 3; i++ {
		_ = breaker.Execute(ctx, failFunc)
	}

	// Verify breaker is open
	assert.Equal(t, StateOpen, breaker.GetState())

	// Wait for half-open timeout
	time.Sleep(100 * time.Millisecond)

	// First request in half-open should be allowed and succeed
	successFunc := func() error { return nil }
	err := breaker.Execute(ctx, successFunc)
	assert.NoError(t, err)
	// Success in half-open should close the circuit
	assert.Equal(t, StateClosed, breaker.GetState())
}

func TestCircuitBreaker_ConcurrentExecution(t *testing.T) {
	t.Parallel()

	breaker := NewCircuitBreaker(5, 1*time.Second)
	ctx := context.Background()
	var successCount atomic.Int32

	// Run 100 concurrent requests
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = breaker.Execute(ctx, func() error {
				successCount.Add(1)
				return nil
			})
		}()
	}
	wg.Wait()

	// Should handle concurrency without race
	assert.True(t, successCount.Load() > 0)
}
