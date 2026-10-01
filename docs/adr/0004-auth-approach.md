# 0004 – Auth Approach — Device Identity Without User Accounts

**Status:** Accepted  
**Date:** 2026-10-01  
**Deciders:** Series author

## Context
In a standard messaging application, users typically register with an email or phone number, authenticate to obtain a session token (like a JWT), and use that token for API calls and WebSocket connections. However, the primary focus of this "System Design Series" project is to demonstrate offline message delivery, mailbox draining, and state synchronization. Introducing full user account management and session-based authentication would add significant complexity that obscures the core lessons. We need a simpler way to identify devices and route messages.

## Decision
We will implement device identity without traditional user accounts, passwords, or sessions. 

- A new device registers via a `POST /devices` endpoint, providing a name and an avatar.
- The server generates and returns a unique `device_id` (using ULID).
- The client stores this `device_id` in `localStorage`.
- When establishing a WebSocket connection, the client sends a `hello` frame containing the `device_id` as its identity claim.
- The server implicitly trusts this `device_id` claim.

*Note: In Episode 4 (E2EE), the device registration process will be extended to include uploading a public identity key, at which point the server could verify key ownership.*

### Rejected Alternatives
- **JWT-based Auth:** Adds unnecessary complexity (token generation, signing, validation, refresh flows) that detracts from the core mailbox/delivery curriculum.
- **OAuth / Social Login:** Completely out of scope for a localized, in-browser simulator and educational demo.

## Consequences
- **Simplicity:** The authentication flow is trivial, allowing learners to focus entirely on message routing and delivery mechanics.
- **Security:** This approach is fundamentally insecure as any user could impersonate another by spoofing a `device_id`. This is accepted as a conscious trade-off for a purely educational demo.
- **State Loss:** If a user clears their browser's `localStorage`, they lose their `device_id` and their simulated identity, requiring them to register a new device.
