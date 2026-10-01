-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS devices (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    avatar      TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL,  -- RFC3339 UTC
    last_seen_at TEXT NOT NULL  -- RFC3339 UTC
);

CREATE TABLE IF NOT EXISTS groups (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS group_members (
    group_id  TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    device_id TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, device_id)
);

-- messages is write-once. payload is opaque to the server
-- (plaintext now, ciphertext in Episode 4 — no migration needed).
CREATE TABLE IF NOT EXISTS messages (
    id              TEXT PRIMARY KEY,
    conversation_id TEXT NOT NULL,
    sender_id       TEXT NOT NULL REFERENCES devices(id),
    payload         TEXT NOT NULL,
    encoding        TEXT NOT NULL DEFAULT 'plaintext',
    sent_at         TEXT NOT NULL,   -- client clock (display only)
    server_at       TEXT NOT NULL    -- server clock (authoritative)
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages(conversation_id, server_at);

-- deliveries is the mailbox. status: pending | delivered | read.
-- The mailbox = rows WHERE status = 'pending'.
-- An ack flips the row. The TTL sweeper hard-deletes expired rows.
CREATE TABLE IF NOT EXISTS deliveries (
    message_id    TEXT NOT NULL REFERENCES messages(id),
    recipient_id  TEXT NOT NULL REFERENCES devices(id),
    seq           INTEGER NOT NULL,  -- per-recipient monotonic cursor
    status        TEXT NOT NULL DEFAULT 'pending',
    expires_at    TEXT NOT NULL,
    delivered_at  TEXT,
    read_at       TEXT,
    PRIMARY KEY (message_id, recipient_id)
);

CREATE INDEX IF NOT EXISTS idx_deliveries_pending
    ON deliveries(recipient_id, seq)
    WHERE status = 'pending';

-- seq_counters tracks the next seq value per recipient.
-- We use a dedicated table so the increment is atomic within a transaction.
CREATE TABLE IF NOT EXISTS seq_counters (
    device_id TEXT PRIMARY KEY REFERENCES devices(id) ON DELETE CASCADE,
    next_seq  INTEGER NOT NULL DEFAULT 1
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_deliveries_pending;
DROP INDEX IF EXISTS idx_messages_conversation;
DROP TABLE IF EXISTS deliveries;
DROP TABLE IF EXISTS seq_counters;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS devices;
-- +goose StatementEnd
