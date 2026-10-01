package memory

import (
	"context"
	"sync"

	"messaging/server/internal/domain"
)

// Bus is an in-memory implementation of port.EventBus.
type Bus struct {
	mu   sync.RWMutex
	subs map[domain.EventType][]func(domain.Event)
}

func NewBus() *Bus {
	return &Bus{
		subs: make(map[domain.EventType][]func(domain.Event)),
	}
}

func (b *Bus) Publish(ctx context.Context, event domain.Event) error {
	b.mu.RLock()
	handlers := b.subs[event.Type]
	b.mu.RUnlock()

	// Fan out to handlers asynchronously
	for _, h := range handlers {
		go h(event)
	}
	return nil
}

func (b *Bus) Subscribe(types []domain.EventType, handler func(domain.Event)) func() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, t := range types {
		b.subs[t] = append(b.subs[t], handler)
	}

	// Simplified unsubscribe (not strictly needed for Episode 1 static wiring)
	return func() {
		// In a real impl, we'd remove the handler pointer
	}
}
