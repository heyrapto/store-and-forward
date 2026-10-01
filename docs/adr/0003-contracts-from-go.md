# 0003 – Protocol Contracts Defined in Go, Consumed by TypeScript

**Status:** Accepted  
**Date:** 2026-10-01  
**Deciders:** Series author

## Context
The client (TypeScript) and server (Go) communicate over WebSockets. We need a reliable way to define the shape of the messages (frames) passing over this connection to ensure both sides understand each other and to maintain type safety.

## Decision
The canonical WebSocket frame shapes are defined as Go structs (with `json` struct tags) in `internal/adapter/transport/ws/protocol.go`. The TypeScript types in the frontend (`client/src/core/socket/protocol.ts`) will be kept in sync manually for now.
Every frame uses a standard JSON envelope format: `{ type, id, ts, payload }`.
The `type` discriminant is a string enum defined on both the Go and TypeScript sides.

**Benefits:**
Go acts as the authoritative backend — the server strictly defines and validates which frames are valid and expected.

**Rejected Alternatives:**
*   Protobuf/gRPC: Rejected because it adds significant toolchain complexity. JSON is more human-readable and inspectable, which is highly beneficial for teaching and debugging in this educational context.

## Consequences
*   Any change to the protocol requires updating two separate files across two different languages.
*   There is a risk of drift between the Go structs and TypeScript definitions if not managed carefully.
*   A codegen script (e.g., using `tygo` or a hand-rolled solution) will be introduced in the future to automate synchronization and eliminate human error.
