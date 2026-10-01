package domain

import "errors"

// Sentinel errors for the domain layer.
// Use cases wrap these with fmt.Errorf("...: %w", err) to add context.

var (
	// ErrDeviceNotFound is returned when a DeviceID references no known device.
	ErrDeviceNotFound = errors.New("device not found")

	// ErrGroupNotFound is returned when a GroupID references no known group.
	ErrGroupNotFound = errors.New("group not found")

	// ErrMessageNotFound is returned when a MessageID has no matching row.
	ErrMessageNotFound = errors.New("message not found")

	// ErrDeliveryNotFound is returned when a (message, recipient) pair has no row.
	ErrDeliveryNotFound = errors.New("delivery not found")

	// ErrAlreadyDelivered is returned when ack_device arrives for an already-delivered row.
	// The use case should treat this as a no-op (idempotent ack).
	ErrAlreadyDelivered = errors.New("delivery already acknowledged")

	// ErrDuplicateMessage is returned when a message with the same ID already exists.
	// The server must treat duplicate sends idempotently.
	ErrDuplicateMessage = errors.New("duplicate message id")

	// ErrDeviceAlreadyConnected is returned when a second hello arrives for a device
	// that already has a live session. The old session should be terminated.
	ErrDeviceAlreadyConnected = errors.New("device already connected")
)
