package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"messaging/server/internal/domain"
	"messaging/server/internal/usecase/port"
)

// AckDelivery handles the ack_device frame from a recipient.
// It transitions the delivery row from pending → delivered and
// sends a receipt frame to the original sender.
type AckDelivery struct {
	Messages   port.MessageRepo
	Deliveries port.DeliveryRepo
	Registry   port.ConnectionRegistry
	Bus        port.EventBus
	Clock      port.Clock
}

// AckDeliveryInput maps from the ack_device frame payload.
type AckDeliveryInput struct {
	MessageID   domain.MessageID
	RecipientID domain.DeviceID
}

// Execute runs the use case.
func (uc *AckDelivery) Execute(ctx context.Context, in AckDeliveryInput) error {
	now := uc.Clock.Now()

	// Advance the delivery state machine.
	err := uc.Deliveries.Advance(ctx, in.MessageID, in.RecipientID, domain.DeliveryDelivered)
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyDelivered) {
			// Idempotent — client retried after a lost ack. No-op.
			return nil
		}
		return fmt.Errorf("ack delivery: advance: %w", err)
	}

	// Look up the message to find the sender.
	msg, err := uc.Messages.FindByID(ctx, in.MessageID)
	if err != nil {
		// Message may have been swept by the TTL cleaner. Not fatal.
		return nil
	}

	// Push a receipt frame to the original sender.
	frame, err := encodeReceiptFrame(in.MessageID, "delivered", now)
	if err == nil {
		uc.Registry.Send(msg.SenderID, frame)
	}

	// Emit delivery.acked for the god-view.
	_ = uc.Bus.Publish(ctx, domain.Event{
		Type:       domain.EventDeliveryAcked,
		OccurredAt: now,
		Payload: domain.DeliveryAckedPayload{
			MessageID:   in.MessageID,
			RecipientID: in.RecipientID,
		},
	})

	return nil
}

// encodeReceiptFrame builds the `receipt` frame sent to the original sender.
func encodeReceiptFrame(messageID domain.MessageID, status string, now time.Time) ([]byte, error) {
	envelope := map[string]any{
		"type": "receipt",
		"id":   string(messageID) + "-rcpt",
		"ts":   now.Format(time.RFC3339),
		"payload": map[string]any{
			"message_id": string(messageID),
			"status":     status,
		},
	}
	return json.Marshal(envelope)
}
