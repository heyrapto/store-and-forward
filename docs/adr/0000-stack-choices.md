# 0000 – Stack Choices

**Status:** Accepted  
**Date:** 2026-10-01  
**Deciders:** Series author

## Context
We are building a WhatsApp-style offline message delivery system as part of a System Design Series. We need to choose the technology stack for the backend, frontend, storage, real-time coordination, and other tooling. The choices need to balance production-readiness, educational value, operational simplicity, and architectural correctness.

## Decision
We have decided to use the following stack:
*   **Backend:** Go. We chose Go for its `goroutines` model which maps perfectly to a per-connection concurrency model for WebSockets. We will use the standard library `http` and `coder/websocket` for networking.
*   **Frontend:** React, TypeScript, Vite, Zustand, and Tailwind CSS. This provides a modern, fast, and scalable foundation for building a robust client application with predictable state management.
*   **Durable Storage:** SQLite in WAL (Write-Ahead Logging) mode. It provides durable, zero-ops storage suitable for a standalone educational project. By putting it behind repository interfaces (ports), a more traditional database like Postgres becomes a drop-in replacement later. We will use `modernc.org/sqlite` as it is a pure Go implementation (no cgo) making cross-compilation easy. We will use `goose` for migrations.
*   **Real-time Coordination:** Redis. It will be used for ephemeral presence, managing connection TTLs, and pub/sub routing across gateways.
*   **Identifiers:** ULIDs (Universally Unique Lexicographically Sortable Identifiers) for globally unique, sortable IDs.

**Rejected Alternatives:**
*   Node.js: Rejected in favor of Go's goroutine model for managing highly concurrent WebSocket connections.
*   PostgreSQL: Rejected as overkill for the initial demo. SQLite allows us to be honest about the tradeoffs while keeping the setup trivial.
*   Firebase: Rejected because it is too opaque and not instructive for teaching backend system design principles.

## Consequences
*   The system is highly self-contained and easy to run locally without complex external dependencies (save for Redis for coordination).
*   Pure Go SQLite simplifies the build pipeline.
*   The use of repository ports ensures that the decision to use SQLite can be reversed without impacting the domain logic if the project outgrows it.
