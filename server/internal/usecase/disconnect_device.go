package usecase

import (
	"context"
	"fmt"

	"messaging/server/internal/domain"
	"messaging/server/internal/usecase/port"
)

// DisconnectDevice cleans up when a WebSocket closes (graceful or abrupt).
// It removes the connection from the registry and marks the device offline.
// Idempotent — safe to call if the device was never connected.
type DisconnectDevice struct {
	Presence  port.PresenceStore
	Registry  port.ConnectionRegistry
	Bus       port.EventBus
	Clock     port.Clock
}

// DisconnectDeviceInput names the device being disconnected.
type DisconnectDeviceInput struct {
	DeviceID domain.DeviceID
}

// Execute runs the use case.
func (uc *DisconnectDevice) Execute(ctx context.Context, in DisconnectDeviceInput) error {
	// Remove from local registry first.
	uc.Registry.Unregister(in.DeviceID)

	// Mark offline in the presence store.
	if err := uc.Presence.SetOffline(ctx, in.DeviceID); err != nil {
		return fmt.Errorf("disconnect device: set offline: %w", err)
	}

	// Broadcast presence change so other clients grey out the avatar dot.
	now := uc.Clock.Now()
	_ = uc.Bus.Publish(ctx, domain.Event{
		Type:       domain.EventDeviceOffline,
		OccurredAt: now,
		Payload: domain.DeviceOfflinePayload{
			DeviceID: in.DeviceID,
		},
	})

	return nil
}
