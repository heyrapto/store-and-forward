# Execution Game Plan

This document outlines the strategic high-level roadmap and decision log for building the "WhatsApp-Style Offline Message Delivery" series.

## 1. The Goal
Build a production-quality system design educational series that teaches real distributed systems concepts through a working implementation. 
**This is not a toy, and it's not full production.** It is a deliberate, focused educational demo designed to make abstract architectural concepts concrete.

## 2. The Bet
**Clean architecture is the key to incremental learning.** 
The bet is that by using strict clean architecture, each episode can extend the previous one without rewriting the core domain. 
- Redis slots in where in-memory maps were.
- SQLite repositories swap to Postgres (if needed) without touching use cases.
This structure is the proof that dependency inversion works in practice.

## 3. Milestones

| Milestone | Rough Timeline | Definition of Done |
| :--- | :--- | :--- |
| **Ep 1: Core Delivery** | Week 1 | In-memory system works. Alice can send to offline Bob. Bob reconnects and gets messages. |
| **Ep 2: Persistence** | Week 2 | Server can restart without losing messages. Full receipt tick lifecycle (✓, ✓✓, blue ✓✓) works. |
| **Ep 3: Pub/Sub Gateway** | Week 3 | Multi-node setup works. Alice and Bob on different ports can chat seamlessly. |
| **Ep 4: E2EE** | Week 4 | End-to-end encryption implemented. DB shows ciphertext, clients show plaintext. |
| **Ep 5: Scale & Load** | Week 5 | Bot swarm can sustain high throughput. Metrics visualize the load. |

## 4. Risk Register

| Risk | Mitigation Strategy |
| :--- | :--- |
| **SQLite Concurrency under 3 gateways** | Enable WAL mode and busy timeouts. Have an honest discussion about this tradeoff vs Postgres. |
| **WebCrypto Complexity in Ep 4** | Stage the rollout. Implement ECDH first, establish shared secrets, then add AES-GCM. |
| **10k Bots for Ep 5 crashing OS limits** | Go goroutines are cheap, but file descriptors aren't. Start with 1k bots. Tune `ulimit` for the demo if necessary. |

## 5. Non-Goals
To keep the series focused, we explicitly will **NOT** build:
- Production-ready security (auth/authz beyond basic tokens)
- Real-world, globally distributed scalability (no multi-region routing)
- Multi-tenant SaaS capabilities
- Native mobile apps (React/HTML5 clients are sufficient for the demo)

## 6. Definition of Done per Episode
An episode is only "Done" when its core demo runs flawlessly on screen:
- **Ep 1:** Store and forward works visibly for offline clients.
- **Ep 2:** A server restart proves persistence; tick states update correctly.
- **Ep 3:** Cross-node routing via Redis succeeds.
- **Ep 4:** Payload inspection proves the server cannot read the messages.
- **Ep 5:** The bot swarm graph shows stable throughput without crashing.

## 7. Open Decisions
- **Group E2EE Strategy:** Do we use pairwise encryption (easier to explain) or Sender Keys (Signal style, more accurate but complex)? *Pending review.*
- **Protocol Types:** Should we add a protobuf/codegen step for protocol types, or stick to raw JSON for educational simplicity? *Pending review.*
- **Load Testing Tooling:** Will Episode 5 use standard `k6` or a custom-built Go bot swarm tailored to our WebSocket protocol? *Pending review.*
