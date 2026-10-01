# Content Guide: WhatsApp-Style Offline Message Delivery

This guide provides internal direction for the series creator on how to present, narrate, and structure the "WhatsApp-Style Offline Message Delivery" system design series.

## 1. Series Identity
**Title:** WhatsApp-Style Offline Message Delivery: System Design Series
**Target Audience:** Intermediate developers who know how to build a basic app (CRUD, typical MVC) but want to understand distributed systems design in practice. 
**Tone & Philosophy:** This is not a CRUD tutorial. This is a system design exploration where code is the medium for demonstrating architecture. We show *why* things break at scale and *how* real systems solve them.

## 2. Narrative Arc
Every episode follows a specific narrative structure:
1. **The Problem:** We encounter something that breaks or is missing.
2. **The Design Phase:** We explore how real-world systems like WhatsApp approach this problem.
3. **The Build Phase:** We implement the solution using clean, deliberate architecture.
4. **The Demo:** Every episode ends with a visceral, on-screen demonstration of the concept in action.

## 3. Episode Script Outline

### Episode 1: The Offline Problem & Core Delivery
- **Hook:** What happens when you send a message on an airplane? Standard REST APIs fail.
- **Research:** How WhatsApp uses a store-and-forward architecture.
- **Design Decision:** In-memory mailboxes, ack-based delivery, and clean architecture separation.
- **Build Walkthrough:** Core domain, in-memory repositories, and basic WebSocket routing.
- **Demo:** (See detailed script below)

### Episode 2: Persistence & Receipts
- **Hook:** If the server restarts, messages are lost. Senders don't know if messages arrived.
- **Research:** How delivery receipts (ticks) work in distributed systems.
- **Design Decision:** SQLite for persistence, seq numbers, and ack/read frames.
- **Build Walkthrough:** Swapping in-memory repos for SQLite, adding receipt routing.
- **Demo:** Show server restart surviving messages; show single tick, double tick, blue tick.

### Episode 3: Scaling the Gateway & Pub/Sub
- **Hook:** One server is a bottleneck. What if Alice and Bob are connected to different servers?
- **Research:** Scaling stateful WebSocket connections and Redis Pub/Sub.
- **Design Decision:** Adding Redis for cross-node routing and presence.
- **Build Walkthrough:** Implementing the Redis adapter, handling presence events.
- **Demo:** Boot two terminal servers. Alice connects to A, Bob to B. Messages flow across.

### Episode 4: End-to-End Encryption (E2EE)
- **Hook:** The server can read everything. How do we ensure privacy?
- **Research:** The Signal Protocol and WebCrypto.
- **Design Decision:** Implementing ECDH key exchange and AES-GCM payload encryption.
- **Build Walkthrough:** Key generation, key distribution via server, encryption at the edge.
- **Demo:** Inspect the server database to show raw ciphertext; clients display plaintext.

### Episode 5: The Bot Swarm (Load & Scale)
- **Hook:** Can our architecture actually handle traffic?
- **Research:** Load testing stateful systems.
- **Design Decision:** Building a Go-based bot swarm to simulate concurrent users.
- **Build Walkthrough:** The swarm orchestrator and metrics collection.
- **Demo:** Launching 1,000+ bots, visualizing throughput, and showing the system gracefully handle the load.

## 4. Demo Script for Episode 1
**Objective:** Make the "store-and-forward" concept visceral.
**Steps:**
1. Spawn 3 phones (simulated clients): Alice, Bob, Carol.
2. Make Bob go **offline** (disconnect his WebSocket).
3. Send multiple messages from Alice to Bob.
4. **Show:** A clock icon in Alice's outbox when Alice herself goes offline. (Demonstrate offline-first outbox).
5. Bring Alice online (messages flush to server).
6. Bring Bob **online**.
7. **Watch:** The mailbox drain on the server side + ticks update on Alice's side.
8. Send a group message to "General".
9. **Watch:** All connected phones receive it instantly.

## 5. Key Talking Points per Episode
- **Ep 1:** Store-and-forward, decoupling senders from receivers, offline-first client design.
- **Ep 2:** Idempotency, exactly-once delivery semantics (via at-least-once + deduplication), read receipts.
- **Ep 3:** Stateful vs. stateless scaling, Pub/Sub fan-out, presence management.
- **Ep 4:** Zero-knowledge architectures, key exchange, public key infrastructure (PKI).
- **Ep 5:** Concurrency, backpressure, observing system limits.

## 6. What NOT to Show
- **Don't show trivial CRUD.** Skip the user registration flow.
- **Don't show config boilerplate.** Pre-write the Dockerfiles and Makefile.
- **Don't apologize for simplifications.** (e.g., using SQLite instead of Cassandra). *Do* acknowledge them, explain the tradeoff, and explain why it's the right choice for the educational scope.

## 7. Code Tour Philosophy
When walking through code, follow the dependency inversion principle explicitly:
1. **Domain First:** Show the pure logic (Entities, Use Cases).
2. **Ports:** Show the interfaces/contracts.
3. **Adapters:** Show the implementation (SQLite, Redis, WebSockets).
4. **Wire-up:** Show `main.go` bringing it all together. 
Make the architecture visible in how you present it.
