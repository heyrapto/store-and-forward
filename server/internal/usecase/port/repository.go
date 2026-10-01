package port

import (
	"context"

	"messaging/server/internal/domain"
)

// DeviceRepo is the read/write interface for device persistence.
type DeviceRepo interface {
	// Save inserts a new device. Returns ErrDuplicateMessage if the ID is taken.
	Save(ctx context.Context, d domain.Device) error

	// FindByID fetches a device by its ID. Returns ErrDeviceNotFound if absent.
	FindByID(ctx context.Context, id domain.DeviceID) (domain.Device, error)

	// ListAll returns every registered device, ordered by created_at ASC.
	ListAll(ctx context.Context) ([]domain.Device, error)

	// UpdateLastSeen stamps the device's last_seen_at column.
	UpdateLastSeen(ctx context.Context, id domain.DeviceID) error

	// Delete removes a device and all cascade records
	Delete(ctx context.Context, id domain.DeviceID) error
}

// GroupRepo is the read interface for group and membership data.
type GroupRepo interface {
	// FindByID fetches a group. Returns ErrGroupNotFound if absent.
	FindByID(ctx context.Context, id domain.GroupID) (domain.Group, error)

	// Members returns the device IDs of all members of a group.
	Members(ctx context.Context, id domain.GroupID) ([]domain.DeviceID, error)

	// AddMember adds a device to a group (idempotent — no error if already a member).
	AddMember(ctx context.Context, groupID domain.GroupID, deviceID domain.DeviceID) error
}

// MessageRepo is the write-once interface for message persistence.
type MessageRepo interface {
	// Save persists a new message. Returns ErrDuplicateMessage if ID is taken.
	Save(ctx context.Context, m domain.Message) error

	// FindByID fetches a message by ID. Returns ErrMessageNotFound if absent.
	FindByID(ctx context.Context, id domain.MessageID) (domain.Message, error)
}

// DeliveryRepo is the read/write interface for the mailbox (deliveries table).
type DeliveryRepo interface {
	// Save inserts a new delivery row with status = pending.
	Save(ctx context.Context, d domain.Delivery) error

	// FindPending returns all pending deliveries for a recipient with
	// seq > afterSeq, ordered by seq ASC. Used for mailbox drain on reconnect.
	FindPending(ctx context.Context, recipientID domain.DeviceID, afterSeq uint64) ([]domain.Delivery, error)

	// Advance updates the status of a delivery row.
	// If the row is already at or past the target status, it returns nil (idempotent).
	Advance(ctx context.Context, messageID domain.MessageID, recipientID domain.DeviceID, next domain.DeliveryStatus) error

	// DeleteExpired removes all delivery rows whose expires_at is before now.
	// Returns the number of rows deleted.
	DeleteExpired(ctx context.Context) (int64, error)

	// NextSeq atomically increments and returns the next seq value for a recipient.
	NextSeq(ctx context.Context, recipientID domain.DeviceID) (uint64, error)

	// MailboxDepth returns the count of pending deliveries for a recipient.
	// Used by the god-view dashboard.
	MailboxDepth(ctx context.Context, recipientID domain.DeviceID) (int64, error)
}
