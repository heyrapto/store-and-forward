package ws

import (
	"encoding/json"

	"messaging/server/internal/domain"
)

// Envelope is the standard format for all WebSocket frames.
type Envelope struct {
	Type    string          `json:"type"`
	ID      string          `json:"id"`
	Ts      string          `json:"ts"` // ISO8601
	Payload json.RawMessage `json:"payload"`
}

// HelloPayload is sent by the client upon connecting.
type HelloPayload struct {
	DeviceID domain.DeviceID `json:"device_id"`
	LastSeq  uint64          `json:"last_seq"`
}

// HelloOkPayload is sent by the server acknowledging the connection.
type HelloOkPayload struct {
	ConnectionID string `json:"connection_id"`
	ServerTime   string `json:"server_time"`
}

// SendPayload is sent by the client to send a message.
type SendPayload struct {
	To   string `json:"to"`
	Text string `json:"text"`
}

// AckServerPayload is sent by the server to confirm persistence.
type AckServerPayload struct {
	RefID string `json:"ref_id"`
}

// AckDevicePayload is sent by the recipient client to confirm receipt.
type AckDevicePayload struct {
	MessageID domain.MessageID `json:"message_id"`
}

// ReadPayload is sent by the recipient client to confirm read, 
// and forwarded by the server to the sender.
type ReadPayload struct {
	MessageID domain.MessageID `json:"message_id"`
}

// SyncBatchPayload is sent by the server to push missed messages.
type SyncBatchPayload struct {
	Messages []DeliverPayload `json:"messages"`
	HasMore  bool             `json:"has_more"`
}

// DeliverPayload represents a message pushed to a recipient.
type DeliverPayload struct {
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	From           string `json:"from"`
	Text           string `json:"text"`
	SentAt         string `json:"sent_at"`
	ServerAt       string `json:"server_at"`
	Seq            uint64 `json:"seq"`
}
