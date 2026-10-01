package usecase

import (
	"context"
	"errors"
	"fmt"

	"messaging/server/internal/domain"
	"messaging/server/internal/usecase/port"
)

// MarkRead handles the `read` frame from a recipient (blue ticks).
// It transitions the delivery row from delivered → read and
// sends a read receipt to the original sender.
type MarkRead struct {
	Messages   port.MessageRepo
	Deliveries port.DeliveryRepo
	Registry   port.ConnectionRegistry
	Bus        port.EventBus
	Clock      port.Clock
}

// MarkReadInput maps from the `read` frame payload.
type MarkReadInput struct {
	MessageID   domain.MessageID
	RecipientID domain.DeviceID
}

// Execute runs the use case.
func (uc *MarkRead) Execute(ctx context.Context, in MarkReadInput) error {
	now := uc.Clock.Now()

	err := uc.Deliveries.Advance(ctx, in.MessageID, in.RecipientID, domain.DeliveryRead)
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyDelivered) || errors.Is(err, domain.ErrInvalidTransition) {
			// Already read or already in a later state. Idempotent no-op.
			return nil
		}
		return fmt.Errorf("mark read: advance: %w", err)
	}

	// Find the sender to push a blue-tick receipt.
	msg, err := uc.Messages.FindByID(ctx, in.MessageID)
	if err != nil {
		return nil // swept or not found — not fatal
	}

	frame, err := encodeReceiptFrame(in.MessageID, "read", now)
	if err == nil {
		uc.Registry.Send(msg.SenderID, frame)
	}

	_ = uc.Bus.Publish(ctx, domain.Event{
		Type:       domain.EventDeliveryRead,
		OccurredAt: now,
		Payload: domain.DeliveryReadPayload{
			MessageID:   in.MessageID,
			RecipientID: in.RecipientID,
		},
	})

	return nil
}
