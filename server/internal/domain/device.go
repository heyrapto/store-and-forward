package domain

import "time"

// DeviceID is the opaque, globally unique identifier for a device.
// We use ULIDs (string form) so they sort lexicographically by creation time.
type DeviceID string

// Device represents a simulated phone registered in the system.
// It is the root aggregate for identity — every message and delivery
// references a DeviceID, not this struct directly.
type Device struct {
	ID         DeviceID  `json:"id"`
	Name       string    `json:"name"`
	Avatar     string    `json:"avatar"`
	CreatedAt  time.Time `json:"created_at"`
	LastSeenAt time.Time `json:"last_seen_at"`
}

// IsZero reports whether this is an empty (unset) device.
func (d Device) IsZero() bool {
	return d.ID == ""
}
