package domain

import (
	"errors"
	"time"
)

// DeliveryStatus is the state machine for a single recipient's delivery edge.
//
//	pending → delivered → read
//
// Transitions are one-way and enforced by Advance().
type DeliveryStatus string

const (
	DeliveryPending   DeliveryStatus = "pending"
	DeliveryDelivered DeliveryStatus = "delivered"
	DeliveryRead      DeliveryStatus = "read"
)

// Delivery is the fan-out edge: one row per (message, recipient) pair.
// The "mailbox" is simply the set of rows where Status == DeliveryPending.
// An ack flips the row; the TTL sweeper hard-deletes expired rows.
type Delivery struct {
	MessageID   MessageID      `json:"message_id"`
	RecipientID DeviceID       `json:"recipient_id"`
	Seq         uint64         `json:"seq"`
	Status      DeliveryStatus `json:"status"`
	ExpiresAt   time.Time      `json:"expires_at"`
	DeliveredAt *time.Time     `json:"delivered_at,omitempty"`
	ReadAt      *time.Time     `json:"read_at,omitempty"`
}

// ErrInvalidTransition is returned when Advance() is called with an
// illegal status transition.
var ErrInvalidTransition = errors.New("invalid delivery status transition")

// Advance moves the delivery through its state machine.
// Only forward transitions are allowed.
func (d *Delivery) Advance(next DeliveryStatus) error {
	switch {
	case d.Status == DeliveryPending && next == DeliveryDelivered:
		d.Status = DeliveryDelivered
		return nil
	case d.Status == DeliveryDelivered && next == DeliveryRead:
		d.Status = DeliveryRead
		return nil
	default:
		return ErrInvalidTransition
	}
}

// IsExpired reports whether this delivery has passed its TTL.
func (d *Delivery) IsExpired(now time.Time) bool {
	return now.After(d.ExpiresAt)
}
