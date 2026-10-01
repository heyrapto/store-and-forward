# Build Plan

## Episodes

### Episode 1: Core
- **Theme:** Core Mechanics
- **Goal:** Establish a baseline store-and-forward messaging system.
- **Backend:** Single gateway, SQLite storage, WS transport, store-and-forward mailbox, auto-join General group.
- **Frontend:** React device spawning (+ button), people strip with live presence dots, DM + group chat, outbox (clock icon), notification banners, unread badges.
- **Demo:** Multiple devices spawn, send messages, see ticks (✓ ✓✓). Disconnect a device, send it messages, reconnect it, and watch the mailbox drain (messages arrive).
- **Key Concept:** Store-and-forward mailbox.

### Episode 2: Reliability
- **Theme:** Robustness
- **Goal:** Handle unreliable networks gracefully.
- **Backend:** Seq cursor sync (hello carries last_seq, server drains from there), TTL sweeper for expired deliveries, last seen timestamp, god-view v1 (gateways panel, sessions, mailbox depth, event log).
- **Frontend:** Retry + backoff (client retries send until ack_server), dedupe by message id.
- **Demo:** Network flake simulation. Messages queue locally, then send when back online. Duplicate messages (simulated) are deduplicated.
- **Key Concept:** At-least-once delivery + idempotency.

### Episode 3: Scale-out
- **Theme:** Distributed System
- **Goal:** Scale beyond a single gateway.
- **Backend:** Redis presence (device→gateway-id with TTL), Redis pub/sub for cross-gateway routing, Docker Compose with 3 gateways behind Caddy.
- **Frontend:** Handle reconnects gracefully when a gateway dies.
- **Demo:** 'Kill gateway' button in god-view. Devices disconnect, reconnect elsewhere, and their mailboxes drain.
- **Key Concept:** Stateless gateways + shared coordination layer.

### Episode 4: E2EE
- **Theme:** Security
- **Goal:** End-to-end encryption.
- **Backend:** Key directory (public identity keys + one-time prekeys uploaded at device creation).
- **Frontend:** WebCrypto API, ECDH + AES-GCM, X3DH, Double Ratchet. Wire-tap panel shows ciphertext vs plaintext. Group messages via pairwise encryption (simpler, alternative to sender keys discussed).
- **Demo:** E2EE messaging, inspecting the wire-tap panel to verify the server only sees ciphertext.
- **Key Concept:** End-to-end encryption, key exchange.

### Episode 5: Load and Chaos
- **Theme:** Resilience and Performance
- **Goal:** Validate system behavior under stress.
- **Backend:** Metrics (Prometheus + Grafana or slog), Go bot swarm (10k clients).
- **Frontend:** Load test results dashboard.
- **Demo:** 10k concurrent bots generating traffic. Latency injection and packet-loss simulation. Observing system limits.
- **Key Concept:** Capacity planning, chaos engineering.

## Episode 1 Build Order
1. Domain types
2. Ports
3. Use cases + fakes
4. SQLite repos + migrations
5. WS transport
6. Registry + presence
7. React device spawning
8. Socket runtime
9. People strip
10. Chat + ticks
11. Outbox
12. Notifications

## Decisions Locked In
| Category | Decision |
|---|---|
| Shared State | Redis |
| Durable Storage | SQLite |
| Groups | One default |
| Contacts | None |
| Architecture | Clean Architecture |
| WS Library | coder/websocket |
| Layout | project/client + project/server |
