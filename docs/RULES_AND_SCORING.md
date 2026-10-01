# Delivery Rules and Receipt Semantics

This document defines the formal rules for how messages move through the system. Think of these as the "rules of the game" governing the distributed state machine.

## 1. The Fundamental Rule
A message stays in the mailbox (server-side persistence) until the recipient device explicitly acknowledges (`ack`) it. **No ack = no removal.**

## 2. Sender-Side Rules
- **Retry Mechanism:** The client must retry a `send` operation with the same message ID until it receives an `ack_server` frame.
- **Offline-First Display:** The client must store the message in its local outbox before attempting to send it, displaying it immediately to the user (often with a clock icon).
- **Reconnect Flush:** Upon reconnecting, the client flushes the outbox in strict insertion order.

## 3. Server-Side Rules
- **Persist Before Ack:** The server must safely persist the message to disk/DB before returning an `ack_server` to the sender.
- **Fan-Out Timing:** Fan-out (creating one delivery row per recipient in a group or 1:1) happens *after* the initial persist.
- **Strict Ordering:** The server must never re-order messages within a specific conversation.

## 4. Delivery Rules
- **At-Least-Once Contract:** The delivery guarantee is at-least-once. The network may duplicate messages.
- **Deduplication:** The client is strictly responsible for deduplication using the unique message ID.
- **Backpressure:** A delivery attempt to a full or slow channel results in dropping the message from the active queue, and the target device is marked for a "reconnect drain" (forcing it to pull state when ready).

## 5. Receipt Rules (Ticks)
- **✓ (Single Tick):** The server successfully persisted the message.
- **✓✓ (Double Tick):** The recipient device acknowledged receipt of the message.
- **Blue ✓✓ (Blue Double Tick):** The recipient device sent a `read` frame (user opened the chat).
- **Group Receipts:** 
  - ✓✓ appears when *ALL* group members have acked the message.
  - Blue ✓✓ appears when *ALL* group members have read the message.

## 6. TTL Rules
- **Expiry:** Delivery rows expire after 30 days in a real system (set to 60 seconds in demo mode for visibility).
- **Sweeper:** A background sweeper task runs every 10 minutes to clean up.
- **Hard Deletion:** Expired rows are hard-deleted from the database.
- **Sender Visibility:** The sender will see no further receipt updates for expired deliveries (they remain at whatever state they were in).

## 7. Seq Rules
- **Monotonicity:** The `seq` (sequence) number is per-recipient and monotonically increasing.
- **Atomic Assignment:** The server assigns the `seq` atomically when adding a message to a recipient's mailbox.
- **Sync Protocol:** On reconnect, the client sends its last known `seq`. The server responds by sending everything strictly after that `seq` in order.

## 8. Presence Rules
- **Online Criteria:** A device is considered "online" if and only if it has a live WebSocket connection AND a valid heartbeat TTL in Redis.
- **Offline Threshold:** A device that fails 3 consecutive heartbeats is marked offline.
- **Broadcast:** Presence events (online, offline, typing) are broadcast to all connected devices that share active conversations with the user.
