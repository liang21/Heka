package event

import (
	"context"
	"sync"

	"github.com/liang21/heka/internal/domain/shared"
)

type eventWrapper struct {
	event   shared.Event
	handler shared.EventHandler
}

type eventBus struct {
	mu       sync.RWMutex
	handlers map[string][]shared.EventHandler
	queue    chan eventWrapper
	workers  int
	stop     chan struct{}
	done     chan struct{}
}

func NewEventBus(workers int) shared.EventBus {
	if workers <= 0 {
		workers = 4
	}
	bus := &eventBus{
		handlers: make(map[string][]shared.EventHandler),
		queue:    make(chan eventWrapper, 100),
		workers:  workers,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
	bus.start()
	return bus
}

func (b *eventBus) start() {
	var wg sync.WaitGroup
	for i := 0; i < b.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case ew := <-b.queue:
					_ = ew.handler(context.Background(), ew.event)
				case <-b.stop:
					return
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(b.done)
	}()
}

func (b *eventBus) Publish(ctx context.Context, events ...shared.Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, event := range events {
		handlers := b.handlers[event.EventName()]
		for _, handler := range handlers {
			select {
			case b.queue <- eventWrapper{event: event, handler: handler}:
			default:
				// queue full, drop event
			}
		}
	}
	return nil
}

func (b *eventBus) Subscribe(eventName string, handler shared.EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], handler)
}

func (b *eventBus) Shutdown() {
	close(b.stop)
	<-b.done
}
