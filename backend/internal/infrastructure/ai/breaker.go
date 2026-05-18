// tasks.md: T081 | spec.md: Circuit Breaker pattern
package ai

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrBreakerOpen = errors.New("circuit breaker is open")
)

// State represents the circuit breaker state
type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateHalfOpen:
		return "half-open"
	case StateOpen:
		return "open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu                sync.Mutex
	maxFailures      int
	resetTimeout     time.Duration
	state            State
	failures         int
	lastFailureTime  time.Time
	halfOpenAttempts int
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
	}
}

// Execute runs the given function if the circuit breaker allows it
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	cb.mu.Lock()

	// Check if we should transition from open to half-open
	if cb.state == StateOpen && time.Since(cb.lastFailureTime) > cb.resetTimeout {
		cb.state = StateHalfOpen
		cb.halfOpenAttempts = 0
	}

	// Fail immediately if circuit is open
	if cb.state == StateOpen {
		cb.mu.Unlock()
		return ErrBreakerOpen
	}

	currentState := cb.state
	cb.mu.Unlock()

	// Execute the function
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailureTime = time.Now()

		// Trip the breaker if threshold reached
		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
		} else if currentState == StateHalfOpen {
			// Failure in half-open trips back to open
			cb.state = StateOpen
		}
		return err
	}

	// Success resets failures in closed state
	if currentState == StateClosed {
		cb.failures = 0
	} else if currentState == StateHalfOpen {
		// Success in half-open closes the circuit
		cb.halfOpenAttempts++
		if cb.halfOpenAttempts >= 1 {
			cb.state = StateClosed
			cb.failures = 0
		}
	}

	return nil
}

// GetState returns the current state (for testing/monitoring)
func (cb *CircuitBreaker) GetState() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
