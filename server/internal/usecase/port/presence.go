package port

import (
	"context"

	"messaging/server/internal/domain"
)

// PresenceStore tracks which devices are currently online.
// Episode 1: implemented by an in-memory map.
// Episode 3: swapped for the Redis adapter without touching any use case.
type PresenceStore interface {
	// SetOnline marks a device as online and associates it with a gateway ID.
	// The gatewayID is used by Episode 3 for cross-gateway routing.
	SetOnline(ctx context.Context, deviceID domain.DeviceID, gatewayID string) error

	// SetOffline marks a device as offline.
	SetOffline(ctx context.Context, deviceID domain.DeviceID) error

	// IsOnline reports whether a device is currently online.
	IsOnline(ctx context.Context, deviceID domain.DeviceID) (bool, error)

	// GatewayOf returns the gatewayID hosting a device, or "" if offline.
	// Episode 1 returns "" for all; Episode 3 returns the real Redis value.
	GatewayOf(ctx context.Context, deviceID domain.DeviceID) (string, error)

	// OnlineDevices returns the IDs of all currently-online devices.
	OnlineDevices(ctx context.Context) ([]domain.DeviceID, error)
}
