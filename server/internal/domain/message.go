package domain

import "time"

// MessageID is the client-generated, globally unique message identifier.
// Clients generate it (ULID), so the server can detect retries by ID.
type MessageID string

// ConversationID is either:
//   - "grp:<group_id>"       for group conversations
//   - "dm:<a>:<b>"           for DMs, where a < b lexicographically
type ConversationID string

// Encoding describes how the payload is encoded.
// "plaintext" for Episode 1; "aes-gcm" in Episode 4.
type Encoding string

const (
	EncodingPlaintext Encoding = "plaintext"
)

// Message is the immutable payload written once and never updated.
// The server routes and fans out; it never rewrites the message body.
type Message struct {
	ID             MessageID      `json:"id"`
	ConversationID ConversationID `json:"conversation_id"`
	SenderID       DeviceID       `json:"sender_id"`
	Payload        string         `json:"payload"`
	Encoding       Encoding       `json:"encoding"`
	SentAt         time.Time      `json:"sent_at"`
	ServerAt       time.Time      `json:"server_at"`
}
