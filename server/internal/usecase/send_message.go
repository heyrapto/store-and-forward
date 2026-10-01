package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"messaging/server/internal/domain"
	"messaging/server/internal/usecase/port"
)

// SendMessage is the core use case.
// It:
//  1. Persists the message (crash-safe: ack sent only after this).
//  2. Resolves recipients: DM → 1 recipient, group → all members except sender.
//  3. Fans out: one delivery row per recipient.
//  4. Pushes deliver frames to online recipients via the Registry.
//  5. Emits MessageQueued for the god-view.
type SendMessage struct {
	Messages   port.MessageRepo
	Deliveries port.DeliveryRepo
	Devices    port.DeviceRepo
	Groups     port.GroupRepo
	Presence   port.PresenceStore
	Registry   port.ConnectionRegistry
	Bus        port.EventBus
	Clock      port.Clock
	IDs        port.IDGenerator

	// TTL controls how long an undelivered message stays in the mailbox.
	TTL time.Duration
}

// SendMessageInput maps directly from the `send` WebSocket frame payload.
type SendMessageInput struct {
	// ClientID is the client-generated idempotency key (from the frame's `id` field).
	ClientID domain.MessageID

	// To is either a DeviceID (for DM) or a GroupID string (for group).
	To string

	// SenderID is the authenticated device.
	SenderID domain.DeviceID

	// Payload is the raw message body (plaintext in Ep1, ciphertext in Ep4).
	Payload string

	// SentAt is the client-side timestamp for display.
	SentAt time.Time
}

// SendMessageOutput contains the IDs needed to build ack_server.
type SendMessageOutput struct {
	MessageID domain.MessageID
	ServerAt  time.Time
}

// Execute runs the use case.
func (uc *SendMessage) Execute(ctx context.Context, in SendMessageInput) (SendMessageOutput, error) {
	now := uc.Clock.Now()

	// Resolve conversation and recipients.
	convID, recipients, err := uc.resolveRecipients(ctx, in)
	if err != nil {
		return SendMessageOutput{}, err
	}

	// Build the message.
	msg := domain.Message{
		ID:             in.ClientID, // honour client-generated ID for dedup
		ConversationID: convID,
		SenderID:       in.SenderID,
		Payload:        in.Payload,
		Encoding:       domain.EncodingPlaintext,
		SentAt:         in.SentAt,
		ServerAt:       now,
	}

	// 1. Persist before acking — crash safety.
	if err := uc.Messages.Save(ctx, msg); err != nil {
		// ErrDuplicateMessage means the client retried. Treat as success.
		if !isDuplicate(err) {
			return SendMessageOutput{}, fmt.Errorf("send message: save: %w", err)
		}
	}

	// 2. Fan out — one delivery row per recipient.
	expiresAt := now.Add(uc.TTL)
	for _, recipientID := range recipients {
		seq, err := uc.Deliveries.NextSeq(ctx, recipientID)
		if err != nil {
			return SendMessageOutput{}, fmt.Errorf("send message: next seq for %s: %w", recipientID, err)
		}

		delivery := domain.Delivery{
			MessageID:   msg.ID,
			RecipientID: recipientID,
			Seq:         seq,
			Status:      domain.DeliveryPending,
			ExpiresAt:   expiresAt,
		}

		if err := uc.Deliveries.Save(ctx, delivery); err != nil {
			// Duplicate delivery — client retry. Skip.
			continue
		}

		// 3. Push to online recipient.
		uc.pushDeliver(ctx, msg, delivery)
	}

	// 4. Emit event for god-view.
	_ = uc.Bus.Publish(ctx, domain.Event{
		Type:       domain.EventMessageQueued,
		OccurredAt: now,
		Payload: domain.MessageQueuedPayload{
			MessageID:      msg.ID,
			SenderID:       in.SenderID,
			ConversationID: convID,
			RecipientCount: len(recipients),
		},
	})

	return SendMessageOutput{MessageID: msg.ID, ServerAt: now}, nil
}

// resolveRecipients returns the conversationID and the list of recipient DeviceIDs.
// DM: "To" is a DeviceID string → conversation is "dm:<lo>:<hi>".
// Group: "To" starts with "grp:" → conversation is "grp:<id>", recipients are all members except sender.
func (uc *SendMessage) resolveRecipients(ctx context.Context, in SendMessageInput) (domain.ConversationID, []domain.DeviceID, error) {
	if strings.HasPrefix(in.To, "grp:") {
		groupID := domain.GroupID(strings.TrimPrefix(in.To, "grp:"))
		members, err := uc.Groups.Members(ctx, groupID)
		if err != nil {
			return "", nil, fmt.Errorf("send message: group members: %w", err)
		}
		// Exclude sender from recipients.
		recipients := make([]domain.DeviceID, 0, len(members))
		for _, m := range members {
			if m != in.SenderID {
				recipients = append(recipients, m)
			}
		}
		return domain.ConversationID("grp:" + string(groupID)), recipients, nil
	}

	// DM: sort IDs to form a stable conversation key.
	recipientID := domain.DeviceID(in.To)
	a, b := string(in.SenderID), string(recipientID)
	if a > b {
		a, b = b, a
	}
	convID := domain.ConversationID("dm:" + a + ":" + b)
	return convID, []domain.DeviceID{recipientID}, nil
}

// pushDeliver encodes and sends a `deliver` frame to a recipient if they are locally connected.
// If not connected, the delivery row stays pending — the mailbox will drain on reconnect.
func (uc *SendMessage) pushDeliver(ctx context.Context, msg domain.Message, d domain.Delivery) {
	frame, err := encodeDeliverFrame(msg, d)
	if err != nil {
		return
	}
	uc.Registry.Send(d.RecipientID, frame)
}

// encodeDeliverFrame serialises the `deliver` WebSocket frame.
func encodeDeliverFrame(msg domain.Message, d domain.Delivery) ([]byte, error) {
	payload := map[string]any{
		"message_id":      string(msg.ID),
		"conversation_id": string(msg.ConversationID),
		"from":            string(msg.SenderID),
		"text":            msg.Payload,
		"sent_at":         msg.SentAt.Format(time.RFC3339),
		"server_at":       msg.ServerAt.Format(time.RFC3339),
		"seq":             d.Seq,
	}
	envelope := map[string]any{
		"type":    "deliver",
		"id":      string(msg.ID),
		"ts":      msg.ServerAt.Format(time.RFC3339),
		"payload": payload,
	}
	return json.Marshal(envelope)
}

// isDuplicate checks whether an error wraps ErrDuplicateMessage.
func isDuplicate(err error) bool {
	return err != nil && strings.Contains(err.Error(), domain.ErrDuplicateMessage.Error())
}
