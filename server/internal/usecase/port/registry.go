package port

import "messaging/server/internal/domain"

// ConnectionRegistry tracks live WebSocket connections on this gateway.
// Each gateway owns its own registry — sockets cannot be serialised across processes.
// The registry is sharded internally (64 shards) to minimise lock contention.
type ConnectionRegistry interface {
	// Register associates a device with a send channel.
	// The channel carries pre-serialised JSON frames.
	// If the device already has a connection, the old send channel is closed
	// (the old goroutines will drain and exit) before the new one is stored.
	Register(deviceID domain.DeviceID, send chan<- []byte)

	// Unregister removes a device's entry. Safe to call if not registered.
	Unregister(deviceID domain.DeviceID)

	// Send attempts to write a frame to a device's outbound channel.
	// Returns false if the device is not registered or the channel is full.
	Send(deviceID domain.DeviceID, frame []byte) bool

	// IsLocal reports whether this gateway has a live connection for the device.
	IsLocal(deviceID domain.DeviceID) bool

	// ConnectedDevices returns the IDs of all locally connected devices.
	ConnectedDevices() []domain.DeviceID

	// Count returns the total number of locally connected devices.
	Count() int
}
