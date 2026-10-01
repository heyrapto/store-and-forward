# 0001 – Server Is the Referee

**Status:** Accepted  
**Date:** 2026-10-01  
**Deciders:** Series author

## Context
In a messaging system, establishing a reliable, strictly ordered timeline of events is critical. Both clients and the server generate events, and their clocks can differ significantly due to clock skew, latency, or adversarial manipulation. We need to decide who dictates the true order of messages and their delivery state.

## Decision
The server acts as the absolute referee and single source of truth.
*   The server assigns authoritative `server_at` timestamps and strictly increasing sequence (`seq`) numbers to all messages.
*   Clients cannot dictate delivery order.
*   The server maintains the official state of a message (pending, delivered, read).
*   Client-generated clocks (`sent_at`) are recorded and stored, but they are strictly used for UI display purposes and local approximation, not for system ordering.
*   The server must persist a message to durable storage before sending an acknowledgment (ack) back to the sender. This ensures crash safety.

This mimics WhatsApp's model where the server acknowledges the sender only after persistence, and the recipient's device acknowledges the server upon receipt.

**Rejected Alternatives:**
*   Client-ordered timestamps: Rejected due to inevitable clock skew and the possibility of cheating or manipulation in adversarial contexts (e.g., a client backdating messages).

## Consequences
*   The backend logic becomes the strict enforcer of consistency.
*   Server persistence latency directly impacts the sender's perceived latency before receiving a "sent" checkmark.
*   Clients must rely on server acks to confirm state transitions rather than assuming immediate success.
