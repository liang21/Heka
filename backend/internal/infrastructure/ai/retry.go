// tasks.md: T083 | spec.md: Retry with exponential backoff
package ai

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrMaxAttemptsReached = errors.New("max retry attempts reached")
)

// Retry executes the given function with exponential backoff retry
func Retry(ctx context.Context, fn func() error, maxAttempts int) error {
	var lastErr error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Check context before each attempt
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("retry cancelled: %w", err)
		}

		// Execute the function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't sleep after the last attempt
		if attempt < maxAttempts-1 {
			// Calculate backoff delay: 2^attempt seconds, capped at 30s
			delay := time.Duration(1<<uint(attempt)) * time.Second
			if delay > 30*time.Second {
				delay = 30 * time.Second
			}

			// Wait for backoff or context cancellation
			select {
			case <-time.After(delay):
				// Continue to next attempt
			case <-ctx.Done():
				return fmt.Errorf("retry cancelled during backoff: %w", ctx.Err())
			}
		}
	}

	return fmt.Errorf("%w: %v", ErrMaxAttemptsReached, lastErr)
}
