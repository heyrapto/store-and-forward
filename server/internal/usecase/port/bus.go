package port

import (
	"context"

	"messaging/server/internal/domain"
)

// EventBus publishes domain events to all interested subscribers.
// Episode 1: in-memory fan-out (subscribers register at startup).
// Episode 3: Redis pub/sub adapter — cross-gateway routing uses this same port.
type EventBus interface {
	// Publish emits an event. Non-blocking: slow subscribers are skipped.
	Publish(ctx context.Context, event domain.Event) error

	// Subscribe registers a handler for all events of the given types.
	// The handler is called in a dedicated goroutine per subscriber.
	// Returns an unsubscribe function.
	Subscribe(types []domain.EventType, handler func(domain.Event)) (unsubscribe func())
}

// Router routes a delivery to a specific device, regardless of which gateway
// the device is connected to.
// Episode 1: local push only (via ConnectionRegistry).
// Episode 3: checks PresenceStore for the gateway ID, publishes via EventBus if remote.
type Router interface {
	// Route attempts to deliver a message frame to a device.
	// If the device is not locally connected, it is a no-op (mailbox covers it).
	Route(ctx context.Context, deviceID domain.DeviceID, frame []byte) error
}
