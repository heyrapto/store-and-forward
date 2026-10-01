# 0006 – State Synchronization via `seq` Cursor

**Status:** Accepted  
**Date:** 2026-10-01  
**Deciders:** Series author

## Context
When a client goes offline and later reconnects, it needs to reliably retrieve any messages delivered to its mailbox during the disconnected period. We need a robust synchronization mechanism to drain the mailbox in the correct order, without missing messages (gaps) or duplicating them at the protocol level.

## Decision
We will use a sequence number (`seq`) cursor for state synchronization.

- Each delivery row in the database will have a `seq` number, which is a per-recipient, monotonically increasing integer.
- On reconnect, the client includes its last known `seq` number in the `hello` WebSocket frame.
- The server queries the durable storage (SQLite): `SELECT * FROM deliveries WHERE recipient_id = ? AND seq > ? AND status = pending ORDER BY seq ASC`.
- The server streams these results to the client via a `sync_batch` frame.

### Rejected Alternatives
- **Timestamp-based Cursors:** Relying on timestamps (`created_at > ?`) is vulnerable to clock skew between the server and devices, making strict ordering and exact resumption unreliable.

## Consequences
- **Reliability:** This approach ensures ordered, gap-free mailbox drain. It is immune to clock skew issues.
- **Client Deduping:** While this prevents protocol-level duplication on sync, clients still need to deduplicate based on message IDs to handle edge cases (e.g., receiving a live message right as a sync batch arrives).
- **Database Consistency:** The `seq` must be assigned atomically. In SQLite, this requires a transaction that inserts the delivery row and simultaneously increments the recipient's sequence counter to prevent race conditions.
