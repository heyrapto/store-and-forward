# Architecture Decision Records

Architecture Decision Records (ADRs) capture the significant design choices made in this project and the reasoning behind them.

| Number | Title | Status | Summary |
| :--- | :--- | :--- | :--- |
| 0000 | Record Architecture Decisions | Accepted | Use standard ADR format to document significant architectural choices. |
| 0001 | Clean Architecture | Accepted | Structure the application following clean architecture principles (domain, usecases, ports, adapters). |
| 0002 | Go and React/TypeScript Stack | Accepted | Use Go for backend services and React/TypeScript for the frontend client. |
| 0003 | SQLite and Redis Storage | Accepted | Use SQLite for durable storage and Redis for real-time coordination and pub/sub. |
| 0004 | Auth Approach — Device Identity Without User Accounts | Accepted | Use generated device IDs for identity to keep the focus on message delivery mechanics. |
| 0005 | One VPS Behind Caddy (Demo Deployment) | Accepted | Deploy 3 Go gateways behind Caddy with Docker Compose on a single VPS for a simple multi-instance demo. |
| 0006 | State Synchronization via `seq` Cursor | Accepted | Use a per-recipient monotonically increasing integer for reliable, gap-free offline mailbox draining. |

## Adding a New ADR

To add a new ADR:
1. Copy the ADR template or an existing ADR file.
2. Name it sequentially using the next available number (e.g., `0007-your-title.md`).
3. Set the initial Status to **Proposed**.
4. Once reviewed and approved, update the Status to **Accepted**.
5. Add the new ADR to the index table in this file.
