// tasks.md: T082 | spec.md: Retry with exponential backoff
package ai

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetry_OnFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	attempts := atomic.Int32{}
	failCount := 3

	failFunc := func() error {
		attempts.Add(1)
		if attempts.Load() < int32(failCount) {
			return errors.New("temporary error")
		}
		return nil
	}

	err := Retry(ctx, failFunc, 3)
	// Should succeed after retries (fails twice, succeeds on 3rd attempt)
	assert.NoError(t, err)
	assert.Equal(t, int32(3), attempts.Load())
}

func TestRetry_ExponentialBackoff(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var delays []time.Duration
	startTime := time.Now()

	attemptFunc := func() error {
		return errors.New("always fails")
	}

	_ = Retry(ctx, attemptFunc, 3)
	elapsed := time.Since(startTime)

	// Should have delays: ~1s, ~2s (total ~3s minimum)
	assert.True(t, elapsed >= 3*time.Second)
	assert.True(t, elapsed < 5*time.Second) // Should not be too long

	for i, delay := range delays {
		expected := time.Duration(1<<uint(i)) * time.Second
		if expected > 30*time.Second {
			expected = 30 * time.Second
		}
		assert.Equal(t, expected, delay)
	}
}

func TestRetry_MaxAttempts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	attempts := atomic.Int32{}

	failFunc := func() error {
		attempts.Add(1)
		return errors.New("persistent error")
	}

	err := Retry(ctx, failFunc, 3)
	assert.Error(t, err)
	assert.Equal(t, int32(3), attempts.Load())
}

func TestRetry_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	attemptFunc := func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	err := Retry(ctx, attemptFunc, 3)
	assert.Error(t, err)
	// Error should be wrapped, so check with errors.Is
	assert.ErrorIs(t, err, context.Canceled)
}

func TestRetry_SuccessOnFirstAttempt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	attempts := atomic.Int32{}

	successFunc := func() error {
		attempts.Add(1)
		return nil
	}

	err := Retry(ctx, successFunc, 3)
	assert.NoError(t, err)
	assert.Equal(t, int32(1), attempts.Load())
}
