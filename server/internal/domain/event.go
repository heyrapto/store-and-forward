package domain

import "time"

// EventType is a string discriminant for domain events.
type EventType string

const (
	EventMessageQueued  EventType = "message.queued"
	EventDeviceOnline   EventType = "device.online"
	EventDeviceOffline  EventType = "device.offline"
	EventDeliveryAcked  EventType = "delivery.acked"
	EventDeliveryRead   EventType = "delivery.read"
	EventDeviceJoined   EventType = "device.joined"
	EventMailboxDrained EventType = "mailbox.drained"
)

// Event is the base envelope for all domain events.
// Use cases emit these through the EventBus port so adapters
// (the admin SSE stream, the god-view, future Redis pub/sub) can react
// without being imported by the use case.
type Event struct {
	Type      EventType
	OccurredAt time.Time
	Payload   any
}

// --- Concrete event payloads ---

// MessageQueuedPayload is emitted when a message has been persisted
// and deliveries have been fanned out.
type MessageQueuedPayload struct {
	MessageID      MessageID
	SenderID       DeviceID
	ConversationID ConversationID
	RecipientCount int
}

// DeviceOnlinePayload is emitted when a device completes its hello handshake.
type DeviceOnlinePayload struct {
	DeviceID DeviceID
	Name     string
}

// DeviceOfflinePayload is emitted when a device disconnects or times out.
type DeviceOfflinePayload struct {
	DeviceID DeviceID
}

// DeliveryAckedPayload is emitted when a device sends ack_device.
type DeliveryAckedPayload struct {
	MessageID   MessageID
	RecipientID DeviceID
}

// DeliveryReadPayload is emitted when a device sends a read frame.
type DeliveryReadPayload struct {
	MessageID   MessageID
	RecipientID DeviceID
}

// DeviceJoinedPayload is emitted when a new device registers.
type DeviceJoinedPayload struct {
	Device Device
}

// MailboxDrainedPayload is emitted after sync_batch completes on reconnect.
type MailboxDrainedPayload struct {
	DeviceID DeviceID
	Count    int
}
