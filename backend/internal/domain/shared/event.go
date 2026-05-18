package shared

import (
	"context"
	"time"
)

// Event is the base interface for all domain events.
type Event interface {
	EventName() string
	OccurredAt() time.Time
}

// EventHandler is a callback that processes a domain event.
type EventHandler func(ctx context.Context, event Event) error

// EventBus provides publish/subscribe semantics for domain events.
type EventBus interface {
	Publish(ctx context.Context, events ...Event) error
	Subscribe(eventName string, handler EventHandler)
}
