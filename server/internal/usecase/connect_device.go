package usecase

import (
	"context"
	"fmt"

	"messaging/server/internal/domain"
	"messaging/server/internal/usecase/port"
)

// ConnectDevice handles the hello handshake when a device opens (or re-opens) a WebSocket.
// It:
//  1. Validates the device exists.
//  2. Kicks any existing connection for the same device (single active session).
//  3. Marks the device online in the PresenceStore.
//  4. Drains the mailbox: returns all pending deliveries with seq > lastSeq.
//  5. Emits DeviceOnline so all other connected clients get a presence update.
type ConnectDevice struct {
	Devices   port.DeviceRepo
	Deliveries port.DeliveryRepo
	Presence  port.PresenceStore
	Registry  port.ConnectionRegistry
	Bus       port.EventBus
	Clock     port.Clock
	GatewayID string // identity of this gateway node
}

// ConnectDeviceInput carries the hello frame payload.
type ConnectDeviceInput struct {
	DeviceID domain.DeviceID
	LastSeq  uint64 // highest seq the client has already processed
	SendCh   chan<- []byte // the writer goroutine's outbound channel
}

// ConnectDeviceOutput carries the data needed to build hello_ok + sync_batch.
type ConnectDeviceOutput struct {
	Device           domain.Device
	PendingDeliveries []domain.Delivery
}

// Execute runs the use case.
func (uc *ConnectDevice) Execute(ctx context.Context, in ConnectDeviceInput) (ConnectDeviceOutput, error) {
	// 1. Validate device exists.
	device, err := uc.Devices.FindByID(ctx, in.DeviceID)
	if err != nil {
		return ConnectDeviceOutput{}, fmt.Errorf("connect device: find: %w", err)
	}

	// 2. Register the connection (kicks old session if one exists).
	uc.Registry.Register(in.DeviceID, in.SendCh)

	// 3. Mark online in presence store.
	if err := uc.Presence.SetOnline(ctx, in.DeviceID, uc.GatewayID); err != nil {
		return ConnectDeviceOutput{}, fmt.Errorf("connect device: set online: %w", err)
	}

	// 4. Stamp last seen.
	_ = uc.Devices.UpdateLastSeen(ctx, in.DeviceID)

	// 5. Drain mailbox — all pending deliveries after the client's cursor.
	pending, err := uc.Deliveries.FindPending(ctx, in.DeviceID, in.LastSeq)
	if err != nil {
		return ConnectDeviceOutput{}, fmt.Errorf("connect device: find pending: %w", err)
	}

	// 6. Emit DeviceOnline so peers update their presence dots.
	now := uc.Clock.Now()
	_ = uc.Bus.Publish(ctx, domain.Event{
		Type:       domain.EventDeviceOnline,
		OccurredAt: now,
		Payload: domain.DeviceOnlinePayload{
			DeviceID: device.ID,
			Name:     device.Name,
		},
	})

	return ConnectDeviceOutput{
		Device:            device,
		PendingDeliveries: pending,
	}, nil
}
