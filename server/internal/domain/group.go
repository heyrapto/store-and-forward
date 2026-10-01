package domain

import "time"

// GroupID is the opaque identifier for a group conversation.
type GroupID string

// Group is a named collection of devices.
// For Episode 1, there is exactly one group: "general".
// All devices are auto-joined on registration.
type Group struct {
	ID        GroupID
	Name      string
	CreatedAt time.Time
}

// GeneralGroupID is the seeded default group every device auto-joins.
const GeneralGroupID GroupID = "general"

// GroupMember links a device to a group.
type GroupMember struct {
	GroupID  GroupID
	DeviceID DeviceID
}
